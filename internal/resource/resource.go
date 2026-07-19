package resource

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type KindSpec struct {
	GVK          schema.GroupVersionKind
	Resource     string
	PriorityPath []string
}

var Supported = []KindSpec{
	{schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}, "pods", []string{"spec", "priorityClassName"}},
	{schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ReplicationController"}, "replicationcontrollers", []string{"spec", "template", "spec", "priorityClassName"}},
	{schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, "deployments", []string{"spec", "template", "spec", "priorityClassName"}},
	{schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}, "statefulsets", []string{"spec", "template", "spec", "priorityClassName"}},
	{schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"}, "daemonsets", []string{"spec", "template", "spec", "priorityClassName"}},
	{schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"}, "replicasets", []string{"spec", "template", "spec", "priorityClassName"}},
	{schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"}, "jobs", []string{"spec", "template", "spec", "priorityClassName"}},
	{schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "CronJob"}, "cronjobs", []string{"spec", "jobTemplate", "spec", "template", "spec", "priorityClassName"}},
}

func ByGVK(gvk schema.GroupVersionKind) (KindSpec, bool) {
	for _, s := range Supported {
		if s.GVK == gvk {
			return s, true
		}
	}
	return KindSpec{}, false
}

func CurrentPriorityClassName(u *unstructured.Unstructured, spec KindSpec) (string, bool, error) {
	return unstructured.NestedString(u.Object, spec.PriorityPath...)
}

func SetPriorityClassName(u *unstructured.Unstructured, spec KindSpec, name string) error {
	return unstructured.SetNestedField(u.Object, name, spec.PriorityPath...)
}

func RemovePriorityClassName(u *unstructured.Unstructured, spec KindSpec) {
	unstructured.RemoveNestedField(u.Object, spec.PriorityPath...)
}

func JSONPointer(path []string) string {
	parts := make([]string, len(path))
	for i, p := range path {
		parts[i] = strings.ReplaceAll(strings.ReplaceAll(p, "~", "~0"), "/", "~1")
	}
	return "/" + strings.Join(parts, "/")
}

func EnsureTemplatePath(u *unstructured.Unstructured, spec KindSpec) error {
	if len(spec.PriorityPath) == 0 {
		return fmt.Errorf("empty priority path")
	}
	if _, ok, _ := unstructured.NestedMap(u.Object, spec.PriorityPath[:len(spec.PriorityPath)-1]...); !ok {
		return fmt.Errorf("resource %s does not contain expected pod spec path %s", spec.GVK.String(), strings.Join(spec.PriorityPath[:len(spec.PriorityPath)-1], "."))
	}
	return nil
}
