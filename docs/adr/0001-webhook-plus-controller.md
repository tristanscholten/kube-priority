# ADR 0001: Webhook plus controller

## Status

Accepted.

## Decision

kube-priority-manager implements both admission webhooks and reconcilers.

## Options considered

| Option | Pros | Cons |
|---|---|---|
| Mutating webhook only | Immediate mutation for new requests | Misses existing resources, cannot repair drift, cannot safely own PriorityClasses over time |
| Controller only | Recovers existing resources and repairs drift | New Pods/workloads may be created before priorityClassName is patched |
| Webhook + controller | Immediate mutation/rejection plus eventual repair/provisioning | More moving parts, webhook availability matters |

## Selected design

Use a mutating webhook for CREATE/UPDATE of supported resources, a validating webhook for clear rejections, and controllers for every supported built-in kind.

The webhook never performs Kubernetes writes. It injects the deterministic PriorityClass name and validates the annotation/conflict policy. The controller creates/validates the cluster-scoped PriorityClass, patches existing resources, emits Events, and repairs drift.

## PriorityClass race

There is a race between setting `priorityClassName` and the PriorityClass existing.

Selected strategy:

- Workload-template resources are admitted and patched; their controller-created Pods may briefly fail until kube-priority-manager creates the PriorityClass. Native Kubernetes controllers retry Pod creation.
- Direct Pod requests are rejected by the validating webhook when the required managed PriorityClass does not exist yet, because there is no parent workload object for this controller to reconcile after rejection.
- Administrators can pre-provision by annotating a workload-template resource, or by creating the managed PriorityClass manifest directly.

This keeps admission paths fast and avoids fragile writes from webhooks while making the consistency window explicit.
