package main

import (
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	priorityadmission "github.com/tristanscholten/kube-priority/internal/admission"
	"github.com/tristanscholten/kube-priority/internal/config"
	prioritycontroller "github.com/tristanscholten/kube-priority/internal/controller"
	prioritymetrics "github.com/tristanscholten/kube-priority/internal/metrics"
)

var scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = schedulingv1.AddToScheme(scheme)
}

func main() {
	opts := config.Defaults()
	var metricsAddr, probeAddr string
	var leaderElect bool
	flag.StringVar(&opts.AnnotationKey, "annotation-key", opts.AnnotationKey, "Annotation key to read numeric PriorityClass values from.")
	flag.Int64Var(&opts.MinimumPriority, "minimum-priority", opts.MinimumPriority, "Minimum permitted priority value.")
	flag.Int64Var(&opts.MaximumPriorityExclusive, "maximum-priority-exclusive", opts.MaximumPriorityExclusive, "Exclusive maximum permitted priority value; cannot exceed 1000000000.")
	flag.Var((*stringFlag)(&opts.PreemptionPolicy), "preemption-policy", "Managed PriorityClass preemption policy: PreemptLowerPriority or Never.")
	flag.Var((*stringFlag)(&opts.ConflictPolicy), "conflict-policy", "Existing priorityClassName policy: reject, overwrite, or preserve.")
	flag.Var((*stringFlag)(&opts.RemovalPolicy), "annotation-removal-policy", "Annotation removal behavior: retain or clear-managed.")
	flag.BoolVar(&opts.EnablePriorityClassGC, "enable-priorityclass-gc", false, "Enable safe optional managed PriorityClass garbage collection. Disabled by default.")
	flag.DurationVar(&opts.PriorityClassGCGracePeriod, "priorityclass-gc-grace-period", 24*time.Hour, "Grace period for optional managed PriorityClass garbage collection.")
	flag.StringVar(&opts.WebhookFailurePolicy, "webhook-failure-policy", opts.WebhookFailurePolicy, "Documented failurePolicy for rendered manifests: Fail or Ignore.")
	flag.StringVar(&opts.ExcludedNamespacesCSV, "excluded-namespaces", opts.ExcludedNamespacesCSV+",kube-priority-system", "Comma-separated namespaces to ignore.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "Metrics bind address.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "Health probe bind address.")
	flag.BoolVar(&leaderElect, "leader-elect", true, "Enable leader election for controller manager.")
	zapOpts := zap.Options{Development: false}
	zapOpts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zapOpts)))
	log := ctrl.Log.WithName("setup")
	if err := opts.Validate(); err != nil {
		log.Error(err, "invalid options")
		os.Exit(1)
	}
	prioritymetrics.Register()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:  scheme,
		Metrics: metricsserver.Options{BindAddress: metricsAddr},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port:    9443,
			TLSOpts: []func(*tls.Config){func(c *tls.Config) { c.MinVersion = tls.VersionTLS12 }},
		}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         leaderElect,
		LeaderElectionID:       "kube-priority.hstr.nl",
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := prioritycontroller.SetupAll(mgr, opts, ctrl.Log); err != nil {
		log.Error(err, "unable to create controllers")
		os.Exit(1)
	}
	priorityadmission.Register(mgr.GetWebhookServer(), mgr.GetClient(), opts, ctrl.Log)
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set readiness check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("webhook-certs", func(_ *http.Request) error { return nil }); err != nil {
		log.Error(err, "unable to set certificate readiness check")
		os.Exit(1)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}

type stringFlag string

func (s *stringFlag) String() string     { return string(*s) }
func (s *stringFlag) Set(v string) error { *s = stringFlag(v); return nil }

var _ admission.Handler
