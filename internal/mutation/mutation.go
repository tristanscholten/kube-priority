package mutation

import (
	"encoding/json"

	"github.com/tristanscholten/kube-priority/internal/priorityclass"
	"github.com/tristanscholten/kube-priority/internal/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Operation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func PatchPriority(u *unstructured.Unstructured, spec resource.KindSpec, className string) ([]byte, bool, error) {
	current, exists, err := resource.CurrentPriorityClassName(u, spec)
	if err != nil {
		return nil, false, err
	}
	if exists && current == className {
		return nil, false, nil
	}
	op := "add"
	if exists {
		op = "replace"
	}
	ops := []Operation{{Op: op, Path: resource.JSONPointer(spec.PriorityPath), Value: className}}
	annPath := "/metadata/annotations/" + escape(priorityclass.ManagedPriorityClassAnnotation)
	if u.GetAnnotations() == nil {
		ops = append([]Operation{{Op: "add", Path: "/metadata/annotations", Value: map[string]string{}}}, ops...)
	}
	ops = append(ops, Operation{Op: "add", Path: annPath, Value: className})
	b, err := json.Marshal(ops)
	return b, true, err
}

func escape(s string) string {
	out := ""
	for _, r := range s {
		switch r {
		case '~':
			out += "~0"
		case '/':
			out += "~1"
		default:
			out += string(r)
		}
	}
	return out
}
