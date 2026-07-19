package mutation

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tristanscholten/kube-priority/internal/priorityclass"
	"github.com/tristanscholten/kube-priority/internal/resource"
)

func TestMutatePriorityIdempotent(t *testing.T) {
	spec, _ := resource.ByAdmissionKind("apps", "v1", "Deployment")
	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "x", Annotations: map[string]string{}},
	}
	obj.Spec.Template.Spec.PriorityClassName = "pc"
	changed, err := MutatePriority(obj, spec, "pc")
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("expected no mutation")
	}
}

func TestMutatePriorityAddsCronJobField(t *testing.T) {
	spec, _ := resource.ByAdmissionKind("batch", "v1", "CronJob")
	obj := &batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{Name: "x"}}
	changed, err := MutatePriority(obj, spec, "pc")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected mutation")
	}
	if got := obj.Spec.JobTemplate.Spec.Template.Spec.PriorityClassName; got != "pc" {
		t.Fatalf("got %q", got)
	}
	if obj.Annotations[priorityclass.ManagedPriorityClassAnnotation] != "pc" {
		t.Fatalf("missing ownership annotation")
	}
}
