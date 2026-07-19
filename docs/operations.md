# Operations and troubleshooting

## Install

```bash
kubectl apply -k config/default
kubectl -n kube-priority-system rollout status deploy/kube-priority
```

## Verify

```bash
kubectl apply -f config/samples/supported-resources.yaml
kubectl get priorityclass | grep kube-priority-hstr-nl
kubectl get deploy important-api -o jsonpath='{.spec.template.spec.priorityClassName}'
```

## Common failures

### Invalid annotation

The validating webhook rejects empty, whitespace, decimal, hexadecimal, overflow, and `>=1000000000` values.

### Direct Pod rejected because PriorityClass missing

Create/annotate a workload resource first or pre-create the generated PriorityClass. Direct Pods have no persisted parent for the controller to reconcile after rejection.

### Conflict with priorityClassName

Default `--conflict-policy=reject` refuses annotation values that disagree with an explicit existing `priorityClassName`. Use `overwrite` or `preserve` only after reviewing ownership implications.

## Upgrade

1. Apply new manifests.
2. Wait for rollout.
3. Confirm webhook configurations point at the ready service.
4. Check Events and `kube_priority_*` metrics.

## Rollback

Roll back the image first. Do not delete PriorityClasses automatically; they are cluster-scoped and can be referenced by existing Pods.
