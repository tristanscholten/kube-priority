package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/tristanscholten/kube-priority/internal/config"
	"github.com/tristanscholten/kube-priority/internal/events"
	"github.com/tristanscholten/kube-priority/internal/metrics"
	"github.com/tristanscholten/kube-priority/internal/priorityclass"
	"github.com/tristanscholten/kube-priority/internal/resource"
	"github.com/tristanscholten/kube-priority/internal/validation"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Spec     resource.KindSpec
	Opts     config.Options
	Log      logr.Logger
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("kind", r.Spec.GVK.Kind, "name", req.String())
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(r.Spec.GVK)
	if err := r.Get(ctx, req.NamespacedName, u); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		metrics.ReconcileTotal.WithLabelValues(r.Spec.GVK.Kind, "get_error").Inc()
		return ctrl.Result{}, err
	}

	if _, excluded := r.Opts.ExcludedNamespaces()[u.GetNamespace()]; excluded {
		return ctrl.Result{}, nil
	}

	res, err := validation.Evaluate(u, r.Spec, r.Opts)
	if err != nil {
		metrics.ValidationFailures.WithLabelValues(r.Spec.GVK.Kind).Inc()
		r.Recorder.Event(u, "Warning", events.ReasonValidationFailed, err.Error())
		log.Info("invalid annotation", "error", err.Error())
		return ctrl.Result{}, nil
	}
	if !res.Annotated {
		return r.handleAnnotationRemoved(ctx, u)
	}
	if res.Preserved {
		metrics.Conflicts.WithLabelValues(r.Spec.GVK.Kind, string(r.Opts.ConflictPolicy)).Inc()
		r.Recorder.Eventf(u, "Warning", events.ReasonConflict, "preserved existing priorityClassName because conflict-policy=preserve")
		return ctrl.Result{}, nil
	}
	if err := r.ensurePriorityClass(ctx, res.Value); err != nil {
		metrics.ReconcileTotal.WithLabelValues(r.Spec.GVK.Kind, "priorityclass_error").Inc()
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}
	current, exists, err := resource.CurrentPriorityClassName(u, r.Spec)
	if err != nil {
		return ctrl.Result{}, err
	}
	if exists && current == res.ClassName {
		metrics.ReconcileTotal.WithLabelValues(r.Spec.GVK.Kind, "noop").Inc()
		return ctrl.Result{}, nil
	}
	before := u.DeepCopy()
	if err := resource.SetPriorityClassName(u, r.Spec, res.ClassName); err != nil {
		return ctrl.Result{}, err
	}
	ann := u.GetAnnotations()
	if ann == nil {
		ann = map[string]string{}
	}
	ann[priorityclass.ManagedPriorityClassAnnotation] = res.ClassName
	u.SetAnnotations(ann)
	if err := r.Patch(ctx, u, client.MergeFrom(before)); err != nil {
		metrics.ReconcileTotal.WithLabelValues(r.Spec.GVK.Kind, "patch_error").Inc()
		return ctrl.Result{}, err
	}
	r.Recorder.Eventf(u, "Normal", events.ReasonApplied, "set priorityClassName to %s", res.ClassName)
	metrics.ReconcileTotal.WithLabelValues(r.Spec.GVK.Kind, "applied").Inc()
	return ctrl.Result{}, nil
}

func (r *Reconciler) handleAnnotationRemoved(ctx context.Context, u *unstructured.Unstructured) (ctrl.Result, error) {
	if r.Opts.RemovalPolicy != config.RemovalClearManaged {
		return ctrl.Result{}, nil
	}
	ann := u.GetAnnotations()
	managed := ann[priorityclass.ManagedPriorityClassAnnotation]
	if managed == "" {
		return ctrl.Result{}, nil
	}
	current, exists, err := resource.CurrentPriorityClassName(u, r.Spec)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !exists || current != managed {
		return ctrl.Result{}, nil
	}
	before := u.DeepCopy()
	resource.RemovePriorityClassName(u, r.Spec)
	delete(ann, priorityclass.ManagedPriorityClassAnnotation)
	u.SetAnnotations(ann)
	if err := r.Patch(ctx, u, client.MergeFrom(before)); err != nil {
		return ctrl.Result{}, err
	}
	r.Recorder.Eventf(u, "Normal", events.ReasonApplied, "cleared managed priorityClassName %s after annotation removal", managed)
	return ctrl.Result{}, nil
}

func (r *Reconciler) ensurePriorityClass(ctx context.Context, value int32) error {
	preemption := corev1.PreemptionPolicy(r.Opts.PreemptionPolicy)
	desired, err := priorityclass.Build(value, preemption)
	if err != nil {
		return err
	}
	pc := &schedulingv1.PriorityClass{}
	if err := r.Get(ctx, client.ObjectKey{Name: desired.Name}, pc); err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.Create(ctx, desired); err != nil {
				return err
			}
			metrics.PriorityClassesCreated.Inc()
			return nil
		}
		return err
	}
	if pc.Value != value {
		return fmt.Errorf("PriorityClass %q has value %d, expected %d", pc.Name, pc.Value, value)
	}
	if !priorityclass.IsManaged(pc) {
		return fmt.Errorf("PriorityClass %q exists but is not managed by kube-priority", pc.Name)
	}
	return nil
}

func SetupAll(mgr ctrl.Manager, opts config.Options, log logr.Logger) error {
	for _, spec := range resource.Supported {
		if err := SetupOne(mgr, opts, log, spec); err != nil {
			return err
		}
	}
	return nil
}

func SetupOne(mgr ctrl.Manager, opts config.Options, log logr.Logger, spec resource.KindSpec) error {
	//nolint:staticcheck // client-go EventRecorder is still used here for broad typed/unstructured Event emission.
	r := &Reconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Recorder: mgr.GetEventRecorderFor("kube-priority"), Spec: spec, Opts: opts, Log: log.WithName("controller")}
	c, err := controller.New("kube-priority-"+spec.Resource, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(spec.GVK)
	pred := predicate.TypedFuncs[*unstructured.Unstructured]{
		CreateFunc: func(e event.TypedCreateEvent[*unstructured.Unstructured]) bool {
			return hasRelevantChange(e.Object, spec, opts)
		},
		UpdateFunc: func(e event.TypedUpdateEvent[*unstructured.Unstructured]) bool {
			return hasRelevantChange(e.ObjectNew, spec, opts) || priorityChanged(e.ObjectOld, e.ObjectNew, spec)
		},
		DeleteFunc: func(e event.TypedDeleteEvent[*unstructured.Unstructured]) bool { return false },
	}
	return c.Watch(source.Kind(mgr.GetCache(), obj, &handler.TypedEnqueueRequestForObject[*unstructured.Unstructured]{}, pred))
}

func hasRelevantChange(u *unstructured.Unstructured, spec resource.KindSpec, opts config.Options) bool {
	_, ok := u.GetAnnotations()[opts.AnnotationKey]
	return ok
}
func priorityChanged(oldObj, newObj *unstructured.Unstructured, spec resource.KindSpec) bool {
	old, _, _ := resource.CurrentPriorityClassName(oldObj, spec)
	neu, _, _ := resource.CurrentPriorityClassName(newObj, spec)
	return old != neu
}

var _ reconcile.Reconciler = (*Reconciler)(nil)
var _ = schema.GroupVersionKind{}
