---
title: Vault Relay | KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: vaultrelay-vault-server-crds
    name: Vault Relay
    parent: vault-server-crds-concepts
    weight: 12
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# VaultRelay

## What is VaultRelay

A `VaultRelay` is a Kubernetes `CustomResourceDefinition` (CRD) which deploys an OpenBao spoke relay (`bao relay run`) in a Kubernetes cluster. The spoke relay connects to a central hub `VaultServer` over mTLS gRPC and runs database secret engine plugins in-process, next to databases that are only reachable from the spoke cluster.

This is the spoke half of the KubeVault hub-spoke model:

- The **hub** cluster runs a `VaultServer`. Its `relay/` backend issues spoke client certificates and brokers database plugin calls to connected spokes through `remote-<db>-plugin` proxy plugins.
- Each **spoke** cluster runs a `VaultRelay`. Databases in the spoke cluster never need to be reachable from the hub; credential operations travel over the relay's outbound mTLS connection.

A `VaultRelay` can be created two ways:

1. **Hub-managed (recommended)**: set `spec.relayPlacementRef` on the hub `VaultServer`. The KubeVault operator resolves the referenced OCM `Placement` and delivers a fully wired `VaultRelay` (plus its AppBinding and credentials) to every selected managed cluster via `ManifestWork`. See the [hub-spoke deployment guide](/docs/guides/hub-spoke/deploy-hub-spoke.md).
2. **Standalone**: create the `VaultRelay` by hand in the spoke cluster, providing the credential material yourself — either a [`spec.bootstrap`](#specbootstrap) join Secret or pre-provisioned [`spec.tls`](#spectls) certificates.

When a `VaultRelay` is reconciled, the KubeVault operator in the spoke cluster provisions a ServiceAccount, the spoke-relay Pod, and — for standalone relays — an AppBinding pointing back at the hub VaultServer. In hub-managed deployments the AppBinding is instead delivered and owned by the hub's `ManifestWork`, and the operator defers to that copy.

## VaultRelay CRD Specification

Like any official Kubernetes resource, a `VaultRelay` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `VaultRelay` object is shown below:

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultRelay
metadata:
  name: vault-agent
  namespace: demo
spec:
  spokeName: cluster-1
  hubVaultRef:
    name: vault
    namespace: demo
    address: https://bao.example.com:8200
    grpcPort: 50053
    caBundle: <base64 PEM bundle>
  bootstrap:
    joinSecretRef:
      name: vault-agent-join
  image: ghcr.io/kubevault/spoke-agent:v0.1.0
  reconnect:
    enabled: true
    backoffSeconds: 5
    maxBackoffSeconds: 300
```

Here, we are going to describe the various sections of the `VaultRelay` crd.

### VaultRelay Spec

#### spec.spokeName

`spec.spokeName` is a required field that specifies the unique identity of this spoke cluster. The hub pins issued client certificates and bootstrap tokens to this name, and database mounts on the hub reference it as `spoke_name`. In hub-managed deployments this is the OCM `ManagedCluster` name.

```yaml
spec:
  spokeName: cluster-1
```

#### spec.hubVaultRef

`spec.hubVaultRef` is a required field that specifies how to reach the hub VaultServer.

```yaml
spec:
  hubVaultRef:
    name: vault                            # VaultServer name on the hub
    namespace: demo                        # VaultServer namespace on the hub
    address: https://bao.example.com:8200  # hub Vault API URL (LoadBalancer address)
    grpcPort: 50053                        # hub gRPC proxy port (default 50053)
    caBundle: <base64 PEM>                 # CA bundle to verify the hub Vault API endpoint
```

- `name`, `namespace`: identify the `VaultServer` object on the hub cluster.
- `address`: the externally reachable hub Vault API URL. In hub-managed deployments this is the hub's LoadBalancer address.
- `grpcPort`: the port of the hub's spoke-relay gRPC proxy. Defaults to `50053`.
- `caBundle`: optional PEM bundle used to verify TLS on the hub Vault API endpoint. When set, the `bao relay join` init container is told to verify the hub against the CA delivered in the `caBundle` key of `spec.bootstrap.joinSecretRef`.

#### spec.bootstrap

`spec.bootstrap` configures the automated `bao relay join` trust bootstrap. When set, the spoke-relay Pod runs a join init container that exchanges a hub-issued bootstrap token for mTLS client credentials before the long-running relay starts. Credentials live on an emptyDir; the relay renews its own certificate in place, and a Pod restart simply re-joins with the current token.

```yaml
spec:
  bootstrap:
    joinSecretRef:
      name: vault-agent-join
```

The referenced Secret must carry:

| key | value |
|---|---|
| `token` | hub bootstrap token (`<id>.<secret>`), minted by the hub's `relay/bootstrap-tokens` endpoint |
| `hubCertHash` | `sha256:<hex>` SPKI pin of the hub's spoke-CA |
| `caBundle` | optional PEM CA bundle for the hub Vault API endpoint |

In hub-managed deployments the operator creates and rotates this Secret automatically (tokens default to a 24h TTL and are rotated when less than a quarter of the TTL remains).

> **Note:** Set exactly one credential source — `spec.bootstrap` (this join flow) or [`spec.tls`](#spectls) (pre-provisioned certificates). The operator rejects a `VaultRelay` that sets neither or both.

#### spec.image

`spec.image` is an optional field that overrides the spoke-relay container image.

#### spec.tokenSecretRef

`spec.tokenSecretRef` is an optional field referencing a Secret with a `token` key holding a Vault token. For standalone relays — where the operator authors the AppBinding — that AppBinding authenticates to the hub Vault with this token instead of the relay's ServiceAccount. It has no effect in hub-managed deployments, where the AppBinding is delivered by the hub `ManifestWork`.

#### spec.tls

`spec.tls` provides pre-provisioned mTLS credentials as an alternative to the `spec.bootstrap` join flow. When set (and `spec.bootstrap` is not), the spoke-relay Pod skips `bao relay join` and runs `bao relay run` directly against these credentials:

- `caSecret`: Secret with `ca.crt` (the hub's spoke-CA certificate).
- `certSecret`: Secret with `tls.crt` and `tls.key` (a client certificate whose CN equals `spec.spokeName`, signed by the spoke-CA).

Both `caSecret` and `certSecret` are required in this mode — the spoke-CA is needed to verify the hub's server certificate. The operator projects them into the relay's credentials directory as `cert.pem`/`key.pem`/`ca.pem` (read-only) and disables in-relay certificate renewal (`-renew-check-every=0`). You rotate the certificate Secrets out-of-band. Note the Pod template references the Secrets by name: changing a Secret *reference* rolls the Pod, but rotating a Secret's contents in place does **not** — restart the relay Pod for it to pick up the new credentials.

#### spec.reconnect

`spec.reconnect` controls whether the relay reconnects to the hub after the stream drops (defaults: enabled).

`bao relay run` has no internal reconnect loop — it exits when the hub stream drops and relies on the Pod to bring it back. The operator therefore maps `reconnect.enabled` onto the Pod's `restartPolicy`:

- **enabled** (default): `restartPolicy: Always` — the kubelet relaunches the relay, which re-dials the hub.
- **disabled**: `restartPolicy: OnFailure` — a clean hub disconnect leaves the relay stopped (the operator also leaves the completed Pod in place rather than recreating it, and reports `status.phase: Disconnected`); genuine crashes still recover in place.

> **Note:** `backoffSeconds` and `maxBackoffSeconds` are not honored — restart timing is governed by the kubelet's CrashLoopBackoff and cannot be tuned through these fields.

#### spec.podTemplate

`spec.podTemplate` is an optional configuration (resources, nodeSelector, etc) for the spoke-relay Pod.

### VaultRelay Status

```yaml
status:
  phase: Connected
  podName: vault-agent-agent
  appBindingRef:
    name: vault-agent-hub-vault
    namespace: demo
  lastHeartbeat: "2026-06-12T10:00:00Z"
  certExpiry: "2026-07-12T10:00:00Z"
```

- `phase`: one of `Pending`, `Connected`, `Disconnected`, `Error`. The operator derives this from the spoke-relay Pod's readiness — `Connected` only once the Pod is Running and Ready.
- `podName`: name of the spoke-relay Pod.
- `appBindingRef`: references the AppBinding for the hub Vault (delivered by the hub `ManifestWork` in hub-managed deployments). Spoke-side consumers (SecretEngine, database role CRDs, KubeDB) use this AppBinding.
- `lastHeartbeat`: timestamp of the last successful heartbeat to the hub, if reported by the relay.
- `certExpiry`: expiry of the spoke client certificate. The operator sets this for **pre-provisioned (`spec.tls`) relays** by reading `certSecret` directly. For **bootstrap relays** the certificate lives only in the pod, so the operator does not set it here — the hub reports it instead on the VaultServer's `status.relayPlacement.clusters[].certExpiry` (sourced from the relay backend's `relay/spokes` endpoint).
- `conditions`: the latest available observations of the VaultRelay's state.

In hub-managed deployments, `status.phase` is scraped back to the hub through `ManifestWork` status feedback and aggregated into the hub VaultServer's `status.relayPlacement`.

## What the operator creates

For every `VaultRelay`, the spoke-side KubeVault operator provisions:

- a **ServiceAccount** for the relay Pod
- the **spoke-relay Pod**: with `spec.bootstrap`, an init container runs `bao relay join` (verifying the hub via the bootstrap token's JWS signature plus the SPKI pin) and writes credentials to an emptyDir; with `spec.tls`, the pre-provisioned certificates are projected read-only and there is no init container. The main container runs `bao relay run -server=<hub>:<grpcPort> -credentials-dir=...`. The Pod is replaced when its template changes (image, podTemplate, resources, or the referenced Secrets).
- an **AppBinding** for the hub VaultServer (standalone relays only). Its parameters carry `deploymentMode: RemoteRelay` and `spokeName`, which the secret engine machinery uses to route database mounts through the hub's `remote-<db>-plugin` proxies. In hub-managed deployments this AppBinding is instead delivered and owned by the hub's `ManifestWork` (labeled `app.kubernetes.io/managed-by: kubevault-hub`), and the operator defers to that copy. See the [AppBinding concept](/docs/concepts/vault-server-crds/appbinding.md).

## Next Steps

- Deploy the full hub-spoke model with OCM: [guide](/docs/guides/hub-spoke/deploy-hub-spoke.md).
- Learn about the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD and its `spec.relayPlacementRef` field.
