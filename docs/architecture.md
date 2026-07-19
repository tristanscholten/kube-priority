# Architecture

kube-priority has four main pieces:

1. **Annotation parser** — validates `kube-priority.hstr.nl` as strict base-10 int32 with hard upper bound `< 1000000000`.
2. **PriorityClass namer/provisioner** — maps values to deterministic names such as `kube-priority-hstr-nl-neg-10`.
3. **Admission webhooks** — mutate supported resources and validate invalid/conflicting annotations before persistence.
4. **Reconcilers** — watch annotated supported resources, create/verify PriorityClasses, patch drift, and handle annotation removal policy.

Supported paths:

| Kind | Field |
|---|---|
| Pod | `spec.priorityClassName` |
| Deployment/StatefulSet/DaemonSet/ReplicaSet/ReplicationController/Job | `spec.template.spec.priorityClassName` |
| CronJob | `spec.jobTemplate.spec.template.spec.priorityClassName` |

Unsupported resources such as Service, Ingress, Namespace, Node, ConfigMap, Secret, PVC, HPA, and PodDisruptionBudget do not contain Pod specs and are not directly affected by PriorityClass.
