package admission

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/tristanscholten/kube-priority/internal/config"
	"github.com/tristanscholten/kube-priority/internal/priorityclass"
	"github.com/tristanscholten/kube-priority/internal/resource"
	"github.com/tristanscholten/kube-priority/internal/validation"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrladmission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Handler struct {
	Client client.Client
	Opts   config.Options
	Log    logr.Logger
	Mode   string
}

func (h *Handler) Handle(ctx context.Context, req ctrladmission.Request) ctrladmission.Response {
	if len(req.Object.Raw) == 0 {
		return ctrladmission.Errored(http.StatusBadRequest, fmt.Errorf("empty admission object"))
	}
	u := &unstructured.Unstructured{}
	if err := json.Unmarshal(req.Object.Raw, &u.Object); err != nil {
		return ctrladmission.Errored(http.StatusBadRequest, fmt.Errorf("malformed object: %w", err))
	}
	gvk := schema.GroupVersionKind{Group: req.Kind.Group, Version: req.Kind.Version, Kind: req.Kind.Kind}
	u.SetGroupVersionKind(gvk)
	spec, ok := resource.ByGVK(gvk)
	if !ok || req.SubResource != "" {
		return ctrladmission.Allowed("unsupported kind or subresource ignored")
	}
	if _, excluded := h.Opts.ExcludedNamespaces()[u.GetNamespace()]; excluded {
		return ctrladmission.Allowed("namespace excluded")
	}
	res, err := validation.Evaluate(u, spec, h.Opts)
	if err != nil {
		return ctrladmission.Denied(err.Error())
	}
	if !res.Annotated || res.Preserved {
		return ctrladmission.Allowed("no mutation required")
	}
	if h.Mode == "validate" {
		if err := h.validatePriorityClassReadiness(ctx, gvk, res); err != nil {
			return ctrladmission.Denied(err.Error())
		}
		return ctrladmission.Allowed("valid kube-priority annotation")
	}
	current, exists, err := resource.CurrentPriorityClassName(u, spec)
	if err != nil {
		return ctrladmission.Errored(http.StatusBadRequest, err)
	}
	if exists && current == res.ClassName {
		return ctrladmission.Allowed("already mutated")
	}
	if err := resource.SetPriorityClassName(u, spec, res.ClassName); err != nil {
		return ctrladmission.Errored(http.StatusBadRequest, err)
	}
	ann := u.GetAnnotations()
	if ann == nil {
		ann = map[string]string{}
	}
	ann[priorityclass.ManagedPriorityClassAnnotation] = res.ClassName
	u.SetAnnotations(ann)
	mutated, err := json.Marshal(u.Object)
	if err != nil {
		return ctrladmission.Errored(http.StatusInternalServerError, err)
	}
	resp := ctrladmission.PatchResponseFromRaw(req.Object.Raw, mutated)
	resp.AuditAnnotations = map[string]string{"kube-priority.hstr.nl/mutated": "true"}
	return resp
}

func (h *Handler) validatePriorityClassReadiness(ctx context.Context, gvk schema.GroupVersionKind, res validation.Result) error {
	pc := &schedulingv1.PriorityClass{}
	if err := h.Client.Get(ctx, client.ObjectKey{Name: res.ClassName}, pc); err != nil {
		if apierrors.IsNotFound(err) {
			if gvk.Kind == "Pod" {
				return fmt.Errorf("PriorityClass %q does not exist yet; retry after it is created or annotate a supported controller resource so kube-priority can provision it", res.ClassName)
			}
			return nil
		}
		return fmt.Errorf("could not verify PriorityClass %q: %w", res.ClassName, err)
	}
	if strings.HasPrefix(pc.Name, "system-") {
		return fmt.Errorf("refusing system PriorityClass %q", pc.Name)
	}
	if pc.Value != res.Value {
		return fmt.Errorf("PriorityClass %q exists with value %d, expected %d", pc.Name, pc.Value, res.Value)
	}
	if !priorityclass.IsManaged(pc) {
		return fmt.Errorf("PriorityClass %q exists but is not managed by kube-priority", pc.Name)
	}
	return nil
}

func Register(mux interface{ Register(string, http.Handler) }, c client.Client, opts config.Options, log logr.Logger) {
	mutating := &ctrladmission.Webhook{Handler: &Handler{Client: c, Opts: opts, Log: log.WithName("mutating-webhook"), Mode: "mutate"}}
	validating := &ctrladmission.Webhook{Handler: &Handler{Client: c, Opts: opts, Log: log.WithName("validating-webhook"), Mode: "validate"}}
	mux.Register("/mutate", mutating)
	mux.Register("/validate", validating)
}

var _ ctrladmission.Handler = (*Handler)(nil)
var _ runtime.Object = (*metav1.Status)(nil)
