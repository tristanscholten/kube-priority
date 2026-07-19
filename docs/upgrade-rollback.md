# Upgrade and rollback

This project has no CRD for the core feature, so upgrades mainly affect controller/webhook behavior and manifests.

## Upgrade checklist

- Read release notes.
- Back up webhook configurations and manager Deployment.
- Apply manifests or Helm chart.
- Wait for two replicas to become ready.
- Confirm both webhooks use the expected `failurePolicy` and namespace selector.
- Create a test Deployment with `kube-priority.hstr.nl: "100"` and verify mutation.

## Rollback checklist

- Roll back the manager image or Helm release.
- Keep existing managed PriorityClasses unless you have manually verified no Pods/workloads reference them.
- If webhooks block cluster changes, temporarily set `failurePolicy: Ignore` or remove the webhook configurations while investigating.
