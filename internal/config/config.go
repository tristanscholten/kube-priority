package config

import (
	"fmt"
	"strings"
	"time"
)

const (
	DefaultAnnotationKey       = "kube-priority.hstr.nl"
	HardMaximumExclusive int64 = 1000000000
)

type ConflictPolicy string

const (
	ConflictReject    ConflictPolicy = "reject"
	ConflictOverwrite ConflictPolicy = "overwrite"
	ConflictPreserve  ConflictPolicy = "preserve"
)

type RemovalPolicy string

const (
	RemovalRetain       RemovalPolicy = "retain"
	RemovalClearManaged RemovalPolicy = "clear-managed"
)

type PreemptionPolicy string

const (
	PreemptLowerPriority PreemptionPolicy = "PreemptLowerPriority"
	PreemptNever         PreemptionPolicy = "Never"
)

type Options struct {
	AnnotationKey              string
	MinimumPriority            int64
	MaximumPriorityExclusive   int64
	PreemptionPolicy           PreemptionPolicy
	ConflictPolicy             ConflictPolicy
	RemovalPolicy              RemovalPolicy
	EnablePriorityClassGC      bool
	PriorityClassGCGracePeriod time.Duration
	WebhookFailurePolicy       string
	ExcludedNamespacesCSV      string
	ManagerNamespace           string
}

func Defaults() Options {
	return Options{
		AnnotationKey:              DefaultAnnotationKey,
		MinimumPriority:            -2147483648,
		MaximumPriorityExclusive:   HardMaximumExclusive,
		PreemptionPolicy:           PreemptLowerPriority,
		ConflictPolicy:             ConflictReject,
		RemovalPolicy:              RemovalRetain,
		PriorityClassGCGracePeriod: 24 * time.Hour,
		WebhookFailurePolicy:       "Fail",
		ExcludedNamespacesCSV:      "kube-system",
		ManagerNamespace:           "kube-priority-manager-system",
	}
}

func (o Options) Validate() error {
	if strings.TrimSpace(o.AnnotationKey) == "" {
		return fmt.Errorf("annotation key must not be empty")
	}
	if o.MaximumPriorityExclusive > HardMaximumExclusive {
		return fmt.Errorf("maximum priority exclusive cannot exceed hard boundary %d", HardMaximumExclusive)
	}
	if o.MaximumPriorityExclusive <= o.MinimumPriority {
		return fmt.Errorf("maximum priority exclusive must be greater than minimum")
	}
	switch o.PreemptionPolicy {
	case PreemptLowerPriority, PreemptNever:
	default:
		return fmt.Errorf("invalid preemption policy %q", o.PreemptionPolicy)
	}
	switch o.ConflictPolicy {
	case ConflictReject, ConflictOverwrite, ConflictPreserve:
	default:
		return fmt.Errorf("invalid conflict policy %q", o.ConflictPolicy)
	}
	switch o.RemovalPolicy {
	case RemovalRetain, RemovalClearManaged:
	default:
		return fmt.Errorf("invalid annotation removal policy %q", o.RemovalPolicy)
	}
	if o.WebhookFailurePolicy != "Fail" && o.WebhookFailurePolicy != "Ignore" {
		return fmt.Errorf("invalid webhook failure policy %q", o.WebhookFailurePolicy)
	}
	return nil
}

func (o Options) ExcludedNamespaces() map[string]struct{} {
	out := map[string]struct{}{}
	for _, n := range strings.Split(o.ExcludedNamespacesCSV, ",") {
		if n = strings.TrimSpace(n); n != "" {
			out[n] = struct{}{}
		}
	}
	if o.ManagerNamespace != "" {
		out[o.ManagerNamespace] = struct{}{}
	}
	return out
}
