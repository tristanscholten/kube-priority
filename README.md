<div align="center">

# ⚡ kube-priority-manager

**Annotation-driven Kubernetes Pod priority, backed by managed PriorityClasses.**

[![Go](https://img.shields.io/badge/Go-1.26.5-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.34+-326CE5?logo=kubernetes&logoColor=white)](https://kubernetes.io)
[![controller-runtime](https://img.shields.io/badge/controller--runtime-v0.24-blue)](https://github.com/kubernetes-sigs/controller-runtime)
[![License](https://img.shields.io/badge/license-Apache--2.0-green.svg)](LICENSE)
[![CI](https://github.com/tristanscholten/kube-priority-manager/actions/workflows/test.yml/badge.svg)](https://github.com/tristanscholten/kube-priority-manager/actions)

</div>

---

## What it does

Add one annotation to a supported Kubernetes resource:

```yaml
metadata:
  annotations:
    kube-priority.hstr.nl: "500000"
```

kube-priority-manager will:

1. validate the value;
2. generate a deterministic cluster-scoped `PriorityClass`;
3. mutate the Pod spec or embedded Pod template with `priorityClassName`;
4. reconcile existing resources, newly admitted resources, and manual drift.

---

## Quick start

```bash
kubectl apply -k config/default
kubectl -n kube-priority-manager-system rollout status deploy/kube-priority-manager
kubectl apply -f config/samples/supported-resources.yaml
```

Or with Helm:

```bash
helm install kube-priority-manager ./charts/kube-priority-manager \
  --namespace kube-priority-manager-system \
  --create-namespace
```

---

## Example

Input Deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: important-api
  annotations:
    kube-priority.hstr.nl: "750000"
spec:
  replicas: 3
  selector:
    matchLabels:
      app: important-api
  template:
    metadata:
      labels:
        app: important-api
    spec:
      containers:
        - name: api
          image: nginx:stable
```

Expected mutation:

```yaml
spec:
  template:
    spec:
      priorityClassName: kube-priority-hstr-nl-750000
```

Expected managed PriorityClass:

```yaml
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: kube-priority-hstr-nl-750000
  labels:
    app.kubernetes.io/name: kube-priority-manager
    app.kubernetes.io/managed-by: kube-priority-manager
    kube-priority.hstr.nl/managed: "true"
  annotations:
    kube-priority.hstr.nl/value: "750000"
value: 750000
globalDefault: false
preemptionPolicy: PreemptLowerPriority
description: "Managed by kube-priority-manager for priority value 750000"
```

CronJob uses the deeper field:

```yaml
spec:
  jobTemplate:
    spec:
      template:
        spec:
          priorityClassName: kube-priority-hstr-nl-750000
```

---

## Supported resources

| Kind | Mutated field |
|---|---|
| `Pod` | `spec.priorityClassName` |
| `Deployment` | `spec.template.spec.priorityClassName` |
| `StatefulSet` | `spec.template.spec.priorityClassName` |
| `DaemonSet` | `spec.template.spec.priorityClassName` |
| `ReplicaSet` | `spec.template.spec.priorityClassName` |
| `ReplicationController` | `spec.template.spec.priorityClassName` |
| `Job` | `spec.template.spec.priorityClassName` |
| `CronJob` | `spec.jobTemplate.spec.template.spec.priorityClassName` |

Not directly affected: `Service`, `Ingress`, `Namespace`, `Node`, `ConfigMap`, `Secret`, `PersistentVolume`, `PersistentVolumeClaim`, `HorizontalPodAutoscaler`, and `PodDisruptionBudget`. Those resources do not contain a Pod spec.

---

## Annotation contract

The annotation key is exactly:

```text
kube-priority.hstr.nl
```

Value rules:

- strict base-10 integer encoded as a string;
- no empty values;
- no whitespace;
- no plus signs;
- no decimal or hex forms;
- signed 32-bit integer only;
- hard bound: `value < 1000000000`;
- configurable minimum via `--minimum-priority`.

Examples:

| Value | Result |
|---:|---|
| `"999999999"` | valid |
| `"1000000000"` | rejected |
| `"-10"` | valid by default |
| `"+10"` | rejected |
| `"10 "` | rejected |

---

## PriorityClass naming

| Value | Generated name |
|---:|---|
| `500000` | `kube-priority-hstr-nl-500000` |
| `0` | `kube-priority-hstr-nl-0` |
| `-10` | `kube-priority-hstr-nl-neg-10` |

Names are deterministic, DNS-subdomain safe, never start with `system-`, and avoid positive/negative collisions.

---

## Why webhook + controller?

See [ADR 0001](docs/adr/0001-webhook-plus-controller.md).

Short version:

- mutating webhook: immediate injection for new CREATE/UPDATE requests;
- validating webhook: clear rejection for invalid/unsafe annotations;
- controller: PriorityClass provisioning, existing resources, drift repair, retries, Events, metrics.

The webhook does not write PriorityClasses in the admission path. Workload templates may have a short consistency window before the controller creates the class. Direct Pods are rejected if the class does not already exist because there is no parent workload for the controller to reconcile.

---

## Configuration flags

```text
--annotation-key=kube-priority.hstr.nl
--minimum-priority=-2147483648
--maximum-priority-exclusive=1000000000
--preemption-policy=PreemptLowerPriority
--conflict-policy=reject
--annotation-removal-policy=retain
--enable-priorityclass-gc=false
--priorityclass-gc-grace-period=24h
--webhook-failure-policy=Fail
--excluded-namespaces=kube-system,kube-priority-manager-system
--leader-elect=true
--metrics-bind-address=:8080
--health-probe-bind-address=:8081
```

`--maximum-priority-exclusive` can be lowered, but not raised above the hard safety boundary `1000000000`.

---

## Conflict policy

| Policy | Behavior |
|---|---|
| `reject` | default; annotation conflicts with another explicit `priorityClassName` are rejected |
| `overwrite` | replace with the generated managed PriorityClass name |
| `preserve` | leave the explicit value unchanged and emit warning metrics/events |

---

## Annotation removal

| Policy | Behavior |
|---|---|
| `retain` | default; leave `priorityClassName` untouched |
| `clear-managed` | clear only when ownership annotation proves kube-priority-manager set it |

---

## Docs

- [Architecture](docs/architecture.md)
- [Security](docs/security.md)
- [Operations](docs/operations.md)
- [Upgrade and rollback](docs/upgrade-rollback.md)
- [ADR 0001: webhook plus controller](docs/adr/0001-webhook-plus-controller.md)

---

## Development

```bash
make fmt
make vet
make test
make docker-build
```

Kind e2e is wired in CI and can be run locally when `kind` is installed:

```bash
make test-e2e
```
