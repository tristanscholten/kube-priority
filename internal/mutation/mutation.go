package mutation

import (
	"encoding/json"

	"github.com/tristanscholten/kube-priority/internal/priorityclass"
	"github.com/tristanscholten/kube-priority/internal/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MutatePriority(obj client.Object, spec resource.KindSpec, className string) (bool, error) {
	current, exists, err := resource.CurrentPriorityClassName(obj, spec)
	if err != nil {
		return false, err
	}
	if exists && current == className {
		return false, nil
	}
	if err := resource.SetPriorityClassName(obj, spec, className); err != nil {
		return false, err
	}
	ann := obj.GetAnnotations()
	if ann == nil {
		ann = map[string]string{}
	}
	ann[priorityclass.ManagedPriorityClassAnnotation] = className
	obj.SetAnnotations(ann)
	return true, nil
}

func RawPatch(original []byte, obj client.Object) ([]byte, bool, error) {
	mutated, err := json.Marshal(obj)
	if err != nil {
		return nil, false, err
	}
	return mutated, string(original) != string(mutated), nil
}
