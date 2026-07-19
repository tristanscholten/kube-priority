package e2e

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestKindDeploymentAnnotation(t *testing.T) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl not installed")
	}
	if err := exec.Command("kubectl", "cluster-info").Run(); err != nil {
		t.Skip("no Kubernetes cluster available")
	}
	run(t, "kubectl", "apply", "-f", "https://github.com/cert-manager/cert-manager/releases/download/v1.19.1/cert-manager.yaml")
	run(t, "kubectl", "-n", "cert-manager", "rollout", "status", "deploy/cert-manager", "--timeout=180s")
	run(t, "kubectl", "-n", "cert-manager", "rollout", "status", "deploy/cert-manager-webhook", "--timeout=180s")
	run(t, "kubectl", "-n", "cert-manager", "rollout", "status", "deploy/cert-manager-cainjector", "--timeout=180s")

	img := os.Getenv("E2E_IMAGE")
	if img == "" {
		img = "kube-priority-manager:local"
	}
	if _, err := exec.LookPath("docker"); err == nil {
		//nolint:gosec // Test harness invokes trusted local docker binary with a test image tag.
		_ = exec.Command("docker", "build", "-t", img, "../..").Run()
	}
	if _, err := exec.LookPath("kind"); err == nil {
		//nolint:gosec // Test harness invokes trusted local kind binary with a test image tag.
		_ = exec.Command("kind", "load", "docker-image", img).Run()
	}

	applyKustomizeWithImage(t, img)
	run(t, "kubectl", "-n", "kube-priority-manager-system", "rollout", "status", "deploy/kube-priority-manager", "--timeout=180s")

	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: e2e-priority
  annotations:
    kube-priority.hstr.nl: "500000"
spec:
  replicas: 1
  selector:
    matchLabels: {app: e2e-priority}
  template:
    metadata:
      labels: {app: e2e-priority}
    spec:
      containers:
        - name: web
          image: nginx:stable
`
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("apply deployment: %v\n%s", err, out)
	}
	for i := 0; i < 30; i++ {
		out, _ := exec.Command("kubectl", "get", "deployment", "e2e-priority", "-o", "jsonpath={.spec.template.spec.priorityClassName}").CombinedOutput()
		if string(out) == "kube-priority-hstr-nl-500000" {
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("priorityClassName not set")
}

func applyKustomizeWithImage(t *testing.T, img string) {
	t.Helper()
	out, err := exec.Command("kubectl", "kustomize", "../../config/default").CombinedOutput()
	if err != nil {
		t.Fatalf("kubectl kustomize: %v\n%s", err, out)
	}
	manifest := strings.ReplaceAll(string(out), "ghcr.io/tristanscholten/kube-priority-manager:latest", img)
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	applyOut, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("kubectl apply rendered manifests: %v\n%s", err, applyOut)
	}
}

func run(t *testing.T, name string, args ...string) {
	t.Helper()
	//nolint:gosec // Test helper invokes fixed, trusted binaries with controlled argument lists.
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
}
