package resource

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestSupportedFieldPaths(t *testing.T) {
	for _, spec := range Supported {
		t.Run(spec.GVK.Kind, func(t *testing.T) {
			u := fixture(spec.GVK)
			if err := SetPriorityClassName(u, spec, "pc"); err != nil {
				t.Fatal(err)
			}
			got, ok, err := CurrentPriorityClassName(u, spec)
			if err != nil {
				t.Fatal(err)
			}
			if !ok || got != "pc" {
				t.Fatalf("got %q ok=%v", got, ok)
			}
		})
	}
}

func TestJSONPointerCronJob(t *testing.T) {
	spec, _ := ByGVK(schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "CronJob"})
	want := "/spec/jobTemplate/spec/template/spec/priorityClassName"
	if got := JSONPointer(spec.PriorityPath); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func fixture(gvk schema.GroupVersionKind) *unstructured.Unstructured {
	u := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": gvk.GroupVersion().String(), "kind": gvk.Kind, "metadata": map[string]interface{}{"name": "x", "namespace": "default"}}}
	switch gvk.Kind {
	case "Pod":
		u.Object["spec"] = map[string]interface{}{}
	case "CronJob":
		u.Object["spec"] = map[string]interface{}{"jobTemplate": map[string]interface{}{"spec": map[string]interface{}{"template": map[string]interface{}{"spec": map[string]interface{}{}}}}}
	default:
		u.Object["spec"] = map[string]interface{}{"template": map[string]interface{}{"spec": map[string]interface{}{}}}
	}
	u.SetGroupVersionKind(gvk)
	return u
}
