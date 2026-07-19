package resource

import "testing"

func TestSupportedTypedAccessors(t *testing.T) {
	for _, spec := range Supported {
		t.Run(spec.Kind, func(t *testing.T) {
			obj := spec.NewObject()
			if err := SetPriorityClassName(obj, spec, "pc"); err != nil {
				t.Fatal(err)
			}
			got, ok, err := CurrentPriorityClassName(obj, spec)
			if err != nil {
				t.Fatal(err)
			}
			if !ok || got != "pc" {
				t.Fatalf("got %q ok=%v", got, ok)
			}
			if err := RemovePriorityClassName(obj, spec); err != nil {
				t.Fatal(err)
			}
			got, ok, err = CurrentPriorityClassName(obj, spec)
			if err != nil {
				t.Fatal(err)
			}
			if ok || got != "" {
				t.Fatalf("got %q ok=%v after clear", got, ok)
			}
		})
	}
}

func TestByAdmissionKind(t *testing.T) {
	spec, ok := ByAdmissionKind("batch", "v1", "CronJob")
	if !ok {
		t.Fatal("CronJob not found")
	}
	if spec.FieldPath != "spec.jobTemplate.spec.template.spec.priorityClassName" {
		t.Fatalf("bad CronJob field path %q", spec.FieldPath)
	}
	if _, ok := ByAdmissionKind("", "v1", "Service"); ok {
		t.Fatal("Service should not be supported")
	}
}
