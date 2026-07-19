package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	ReconcileTotal         = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "kube_priority_reconcile_total", Help: "Total kube-priority reconciliations."}, []string{"kind", "result"})
	ValidationFailures     = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "kube_priority_validation_failures_total", Help: "Total invalid kube-priority annotations."}, []string{"kind"})
	Conflicts              = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "kube_priority_conflicts_total", Help: "Total priorityClassName conflicts."}, []string{"kind", "policy"})
	PriorityClassesCreated = prometheus.NewCounter(prometheus.CounterOpts{Name: "kube_priority_priorityclasses_created_total", Help: "Managed PriorityClasses created."})
)

func Register() {
	metrics.Registry.MustRegister(ReconcileTotal, ValidationFailures, Conflicts, PriorityClassesCreated)
}
