package resource

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KindSpec struct {
	Group     string
	Version   string
	Kind      string
	Resource  string
	FieldPath string
	NewObject func() client.Object
	Get       func(client.Object) (string, bool, error)
	Set       func(client.Object, string) error
	Clear     func(client.Object) error
}

var Supported = []KindSpec{
	{
		Group:     "",
		Version:   "v1",
		Kind:      "Pod",
		Resource:  "pods",
		FieldPath: "spec.priorityClassName",
		NewObject: func() client.Object { return &corev1.Pod{} },
		Get:       typedGet[*corev1.Pod](func(p *corev1.Pod) string { return p.Spec.PriorityClassName }),
		Set:       typedSet[*corev1.Pod](func(p *corev1.Pod, name string) { p.Spec.PriorityClassName = name }),
		Clear:     typedClear[*corev1.Pod](func(p *corev1.Pod) { p.Spec.PriorityClassName = "" }),
	},
	{
		Group:     "",
		Version:   "v1",
		Kind:      "ReplicationController",
		Resource:  "replicationcontrollers",
		FieldPath: "spec.template.spec.priorityClassName",
		NewObject: func() client.Object { return &corev1.ReplicationController{} },
		Get: typedGet[*corev1.ReplicationController](func(rc *corev1.ReplicationController) string {
			if rc.Spec.Template == nil {
				return ""
			}
			return rc.Spec.Template.Spec.PriorityClassName
		}),
		Set: typedSet[*corev1.ReplicationController](func(rc *corev1.ReplicationController, name string) {
			if rc.Spec.Template == nil {
				rc.Spec.Template = &corev1.PodTemplateSpec{}
			}
			rc.Spec.Template.Spec.PriorityClassName = name
		}),
		Clear: typedClear[*corev1.ReplicationController](func(rc *corev1.ReplicationController) {
			if rc.Spec.Template != nil {
				rc.Spec.Template.Spec.PriorityClassName = ""
			}
		}),
	},
	{
		Group:     "apps",
		Version:   "v1",
		Kind:      "Deployment",
		Resource:  "deployments",
		FieldPath: "spec.template.spec.priorityClassName",
		NewObject: func() client.Object { return &appsv1.Deployment{} },
		Get:       typedGet[*appsv1.Deployment](func(d *appsv1.Deployment) string { return d.Spec.Template.Spec.PriorityClassName }),
		Set:       typedSet[*appsv1.Deployment](func(d *appsv1.Deployment, name string) { d.Spec.Template.Spec.PriorityClassName = name }),
		Clear:     typedClear[*appsv1.Deployment](func(d *appsv1.Deployment) { d.Spec.Template.Spec.PriorityClassName = "" }),
	},
	{
		Group:     "apps",
		Version:   "v1",
		Kind:      "StatefulSet",
		Resource:  "statefulsets",
		FieldPath: "spec.template.spec.priorityClassName",
		NewObject: func() client.Object { return &appsv1.StatefulSet{} },
		Get:       typedGet[*appsv1.StatefulSet](func(s *appsv1.StatefulSet) string { return s.Spec.Template.Spec.PriorityClassName }),
		Set:       typedSet[*appsv1.StatefulSet](func(s *appsv1.StatefulSet, name string) { s.Spec.Template.Spec.PriorityClassName = name }),
		Clear:     typedClear[*appsv1.StatefulSet](func(s *appsv1.StatefulSet) { s.Spec.Template.Spec.PriorityClassName = "" }),
	},
	{
		Group:     "apps",
		Version:   "v1",
		Kind:      "DaemonSet",
		Resource:  "daemonsets",
		FieldPath: "spec.template.spec.priorityClassName",
		NewObject: func() client.Object { return &appsv1.DaemonSet{} },
		Get:       typedGet[*appsv1.DaemonSet](func(d *appsv1.DaemonSet) string { return d.Spec.Template.Spec.PriorityClassName }),
		Set:       typedSet[*appsv1.DaemonSet](func(d *appsv1.DaemonSet, name string) { d.Spec.Template.Spec.PriorityClassName = name }),
		Clear:     typedClear[*appsv1.DaemonSet](func(d *appsv1.DaemonSet) { d.Spec.Template.Spec.PriorityClassName = "" }),
	},
	{
		Group:     "apps",
		Version:   "v1",
		Kind:      "ReplicaSet",
		Resource:  "replicasets",
		FieldPath: "spec.template.spec.priorityClassName",
		NewObject: func() client.Object { return &appsv1.ReplicaSet{} },
		Get:       typedGet[*appsv1.ReplicaSet](func(r *appsv1.ReplicaSet) string { return r.Spec.Template.Spec.PriorityClassName }),
		Set:       typedSet[*appsv1.ReplicaSet](func(r *appsv1.ReplicaSet, name string) { r.Spec.Template.Spec.PriorityClassName = name }),
		Clear:     typedClear[*appsv1.ReplicaSet](func(r *appsv1.ReplicaSet) { r.Spec.Template.Spec.PriorityClassName = "" }),
	},
	{
		Group:     "batch",
		Version:   "v1",
		Kind:      "Job",
		Resource:  "jobs",
		FieldPath: "spec.template.spec.priorityClassName",
		NewObject: func() client.Object { return &batchv1.Job{} },
		Get:       typedGet[*batchv1.Job](func(j *batchv1.Job) string { return j.Spec.Template.Spec.PriorityClassName }),
		Set:       typedSet[*batchv1.Job](func(j *batchv1.Job, name string) { j.Spec.Template.Spec.PriorityClassName = name }),
		Clear:     typedClear[*batchv1.Job](func(j *batchv1.Job) { j.Spec.Template.Spec.PriorityClassName = "" }),
	},
	{
		Group:     "batch",
		Version:   "v1",
		Kind:      "CronJob",
		Resource:  "cronjobs",
		FieldPath: "spec.jobTemplate.spec.template.spec.priorityClassName",
		NewObject: func() client.Object { return &batchv1.CronJob{} },
		Get:       typedGet[*batchv1.CronJob](func(c *batchv1.CronJob) string { return c.Spec.JobTemplate.Spec.Template.Spec.PriorityClassName }),
		Set:       typedSet[*batchv1.CronJob](func(c *batchv1.CronJob, name string) { c.Spec.JobTemplate.Spec.Template.Spec.PriorityClassName = name }),
		Clear:     typedClear[*batchv1.CronJob](func(c *batchv1.CronJob) { c.Spec.JobTemplate.Spec.Template.Spec.PriorityClassName = "" }),
	},
}

func ByAdmissionKind(group, version, kind string) (KindSpec, bool) {
	for _, spec := range Supported {
		if spec.Group == group && spec.Version == version && spec.Kind == kind {
			return spec, true
		}
	}
	return KindSpec{}, false
}

func CurrentPriorityClassName(obj client.Object, spec KindSpec) (string, bool, error) {
	return spec.Get(obj)
}

func SetPriorityClassName(obj client.Object, spec KindSpec, name string) error {
	return spec.Set(obj, name)
}

func RemovePriorityClassName(obj client.Object, spec KindSpec) error {
	return spec.Clear(obj)
}

func typedGet[T client.Object](get func(T) string) func(client.Object) (string, bool, error) {
	return func(obj client.Object) (string, bool, error) {
		typed, ok := obj.(T)
		if !ok {
			return "", false, typeError[T](obj)
		}
		value := get(typed)
		return value, value != "", nil
	}
}

func typedSet[T client.Object](set func(T, string)) func(client.Object, string) error {
	return func(obj client.Object, name string) error {
		typed, ok := obj.(T)
		if !ok {
			return typeError[T](obj)
		}
		set(typed, name)
		return nil
	}
}

func typedClear[T client.Object](clear func(T)) func(client.Object) error {
	return func(obj client.Object) error {
		typed, ok := obj.(T)
		if !ok {
			return typeError[T](obj)
		}
		clear(typed)
		return nil
	}
}

func typeError[T client.Object](obj client.Object) error {
	var zero T
	return fmt.Errorf("resource accessor expected %T, got %T", zero, obj)
}
