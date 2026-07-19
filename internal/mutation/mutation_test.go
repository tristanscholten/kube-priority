package mutation

import (
	"github.com/tristanscholten/kube-priority/internal/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
	"testing"
)

func TestPatchPriorityIdempotent(t *testing.T) {
	gvk := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
	spec, _ := resource.ByGVK(gvk)
	u := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "apps/v1", "kind": "Deployment", "metadata": map[string]interface{}{"name": "x", "annotations": map[string]interface{}{}}, "spec": map[string]interface{}{"template": map[string]interface{}{"spec": map[string]interface{}{"priorityClassName": "pc"}}}}}
	patch, changed, err := PatchPriority(u, spec, "pc")
	if err != nil {
		t.Fatal(err)
	}
	if changed || patch != nil {
		t.Fatalf("expected no patch")
	}
}

func TestPatchPriorityAddsPath(t *testing.T) {
	gvk := schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "CronJob"}
	spec, _ := resource.ByGVK(gvk)
	u := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "batch/v1", "kind": "CronJob", "metadata": map[string]interface{}{"name": "x", "annotations": map[string]interface{}{}}, "spec": map[string]interface{}{"jobTemplate": map[string]interface{}{"spec": map[string]interface{}{"template": map[string]interface{}{"spec": map[string]interface{}{}}}}}}}
	patch, changed, err := PatchPriority(u, spec, "pc")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatalf("expected change")
	}
	if !strings.Contains(string(patch), "/spec/jobTemplate/spec/template/spec/priorityClassName") {
		t.Fatalf("bad patch: %s", patch)
	}
}
