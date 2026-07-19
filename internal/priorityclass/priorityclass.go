package priorityclass

import (
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AppName                        = "kube-priority"
	ManagedLabel                   = "kube-priority.hstr.nl/managed"
	ValueAnnotation                = "kube-priority.hstr.nl/value"
	ManagedPriorityClassAnnotation = "kube-priority.hstr.nl/managed-priority-class"
)

func NameForValue(v int32) (string, error) {
	encoded := strconv.FormatInt(int64(v), 10)
	if v < 0 {
		encoded = "neg-" + strings.TrimPrefix(encoded, "-")
	}
	name := "kube-priority-hstr-nl-" + encoded
	if strings.HasPrefix(name, "system-") {
		return "", fmt.Errorf("generated PriorityClass name must not start with system-")
	}
	if len(name) > 253 {
		return "", fmt.Errorf("generated PriorityClass name exceeds DNS subdomain length")
	}
	return name, nil
}

func ManagedLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       AppName,
		"app.kubernetes.io/managed-by": AppName,
		ManagedLabel:                   "true",
	}
}

func Build(value int32, preemption corev1.PreemptionPolicy) (*schedulingv1.PriorityClass, error) {
	name, err := NameForValue(value)
	if err != nil {
		return nil, err
	}
	return &schedulingv1.PriorityClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      ManagedLabels(),
			Annotations: map[string]string{ValueAnnotation: strconv.FormatInt(int64(value), 10)},
		},
		Value:            value,
		GlobalDefault:    false,
		PreemptionPolicy: &preemption,
		Description:      fmt.Sprintf("Managed by kube-priority for priority value %d", value),
	}, nil
}

func IsManaged(pc *schedulingv1.PriorityClass) bool {
	return pc != nil && pc.Labels[ManagedLabel] == "true" && pc.Labels["app.kubernetes.io/managed-by"] == AppName
}
