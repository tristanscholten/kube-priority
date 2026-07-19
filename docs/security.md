# Security

Security posture:

- Non-root distroless runtime image.
- Read-only root filesystem.
- Linux capabilities dropped.
- `allowPrivilegeEscalation: false`.
- `seccompProfile: RuntimeDefault`.
- TLS-only webhook serving; no `InsecureSkipVerify`.
- Exact webhook rules; no wildcard API groups/resources.
- Exact RBAC; no wildcard permissions.
- `kube-system` and the manager namespace are excluded by default.
- Webhooks log metadata only, never full resource bodies.
- The annotation value is strictly parsed and bounded.

RBAC note: PriorityClass delete is not granted by default. If `--enable-priorityclass-gc=true`, grant delete only for `priorityclasses` and only after accepting the documented race trade-offs.

Certificate options:

1. **cert-manager** — default manifests include an Issuer and Certificate with CA injection annotations.
2. **Without cert-manager** — create a TLS Secret named `kube-priority-webhook-server-cert` containing `tls.crt` and `tls.key`, and inject the CA bundle into both webhook configurations using your GitOps/cert rotation system.
