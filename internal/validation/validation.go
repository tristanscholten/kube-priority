package validation

import (
	"fmt"

	"github.com/tristanscholten/kube-priority/internal/annotation"
	"github.com/tristanscholten/kube-priority/internal/config"
	"github.com/tristanscholten/kube-priority/internal/priorityclass"
	"github.com/tristanscholten/kube-priority/internal/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Result struct {
	Annotated bool
	Value     int32
	ClassName string
	Conflict  bool
	Preserved bool
}

func Evaluate(obj client.Object, spec resource.KindSpec, opts config.Options) (Result, error) {
	ann := obj.GetAnnotations()
	raw, ok := ann[opts.AnnotationKey]
	if !ok {
		return Result{}, nil
	}
	parsed, err := annotation.Parse(raw, annotation.Bounds{Minimum: opts.MinimumPriority, MaximumExclusive: opts.MaximumPriorityExclusive})
	if err != nil {
		return Result{Annotated: true}, err
	}
	name, err := priorityclass.NameForValue(parsed.Value)
	if err != nil {
		return Result{Annotated: true}, err
	}
	current, exists, err := resource.CurrentPriorityClassName(obj, spec)
	if err != nil {
		return Result{Annotated: true}, err
	}
	res := Result{Annotated: true, Value: parsed.Value, ClassName: name}
	if exists && current != "" && current != name {
		switch opts.ConflictPolicy {
		case config.ConflictReject:
			return res, fmt.Errorf("annotation %s conflicts with existing priorityClassName %q; expected %q", opts.AnnotationKey, current, name)
		case config.ConflictPreserve:
			res.Conflict = true
			res.Preserved = true
		case config.ConflictOverwrite:
			res.Conflict = true
		default:
			return res, fmt.Errorf("unsupported conflict policy %q", opts.ConflictPolicy)
		}
	}
	return res, nil
}
