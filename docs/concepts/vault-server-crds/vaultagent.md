---
title: Vault Agent | KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: vaultagent-vault-server-crds
    name: Vault Agent
    parent: vault-server-crds-concepts
    weight: 12
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# VaultAgent

## What is VaultAgent

A `VaultAgent` is a Kubernetes `CustomResourceDefinition` (CRD) which deploys an OpenBao spoke agent (`bao agent run`) in a Kubernetes cluster. The spoke agent connects to a central hub `VaultServer` over mTLS gRPC and runs database secret engine plugins in-process, next to databases that are only reachable from the spoke cluster.

This is the spoke half of the KubeVault hub-spoke model:

- The **hub** cluster runs a `VaultServer`. Its `agent/` backend issues spoke client certificates and brokers database plugin calls to connected spokes through `remote-<db>-plugin` proxy plugins.
- Each **spoke** cluster runs a `VaultAgent`. Databases in the spoke cluster never need to be reachable from the hub; credential operations travel over the agent's outbound mTLS connection.

A `VaultAgent` can be created two ways:

1. **Hub-managed (recommended)**: set `spec.agentPlacementRef` on the hub `VaultServer`. The KubeVault operator resolves the referenced OCM `Placement` and delivers a fully wired `VaultAgent` (plus its AppBinding and credentials) to every selected managed cluster via `ManifestWork`. See the [hub-spoke deployment guide](/docs/guides/hub-spoke/deploy-hub-spoke.md).
2. **Standalone**: create the `VaultAgent` by hand in the spoke cluster, providing the credential material yourself — either a [`spec.bootstrap`](#specbootstrap) join Secret or pre-provisioned [`spec.tls`](#spectls) certificates.

When a `VaultAgent` is reconciled, the KubeVault operator in the spoke cluster provisions a ServiceAccount, the spoke-agent Pod, and — for standalone agents — an AppBinding pointing back at the hub VaultServer. In hub-managed deployments the AppBinding is instead delivered and owned by the hub's `ManifestWork`, and the operator defers to that copy.

## VaultAgent CRD Specification

Like any official Kubernetes resource, a `VaultAgent` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `VaultAgent` object is shown below:

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultAgent
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

Here, we are going to describe the various sections of the `VaultAgent` crd.

### VaultAgent Spec

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
- `grpcPort`: the port of the hub's spoke-agent gRPC proxy. Defaults to `50053`.
- `caBundle`: optional PEM bundle used to verify TLS on the hub Vault API endpoint. When set, the `bao agent join` init container is told to verify the hub against the CA delivered in the `caBundle` key of `spec.bootstrap.joinSecretRef`.

#### spec.bootstrap

`spec.bootstrap` configures the automated `bao agent join` trust bootstrap. When set, the spoke-agent Pod runs a join init container that exchanges a hub-issued bootstrap token for mTLS client credentials before the long-running agent starts. Credentials live on an emptyDir; the agent renews its own certificate in place, and a Pod restart simply re-joins with the current token.

```yaml
spec:
  bootstrap:
    joinSecretRef:
      name: vault-agent-join
```

The referenced Secret must carry:

| key | value |
|---|---|
| `token` | hub bootstrap token (`<id>.<secret>`), minted by the hub's `agent/bootstrap-tokens` endpoint |
| `hubCertHash` | `sha256:<hex>` SPKI pin of the hub's spoke-CA |
| `caBundle` | optional PEM CA bundle for the hub Vault API endpoint |

In hub-managed deployments the operator creates and rotates this Secret automatically (tokens default to a 24h TTL and are rotated when less than a quarter of the TTL remains).

> **Note:** Set exactly one credential source — `spec.bootstrap` (this join flow) or [`spec.tls`](#spectls) (pre-provisioned certificates). The operator rejects a `VaultAgent` that sets neither or both.

#### spec.image

`spec.image` is an optional field that overrides the spoke-agent container image.

#### spec.tokenSecretRef

`spec.tokenSecretRef` is an optional field referencing a Secret with a `token` key holding a Vault token. For standalone agents — where the operator authors the AppBinding — that AppBinding authenticates to the hub Vault with this token instead of the agent's ServiceAccount. It has no effect in hub-managed deployments, where the AppBinding is delivered by the hub `ManifestWork`.

#### spec.tls

`spec.tls` provides pre-provisioned mTLS credentials as an alternative to the `spec.bootstrap` join flow. When set (and `spec.bootstrap` is not), the spoke-agent Pod skips `bao agent join` and runs `bao agent run` directly against these credentials:

- `caSecret`: Secret with `ca.crt` (the hub's spoke-CA certificate).
- `certSecret`: Secret with `tls.crt` and `tls.key` (a client certificate whose CN equals `spec.spokeName`, signed by the spoke-CA).

The operator projects these into the agent's credentials directory as `cert.pem`/`key.pem`/`ca.pem` (read-only) and disables in-agent certificate renewal (`-renew-check-every=0`). You rotate the certificate Secrets out-of-band; changing them rolls the Pod.

#### spec.reconnect

`spec.reconnect` configures automatic reconnection to the hub (defaults: enabled, 5s initial backoff, 300s max backoff). The operator surfaces these to the agent container as `BAO_AGENT_RECONNECT_ENABLED`, `BAO_AGENT_RECONNECT_BACKOFF_SECONDS`, and `BAO_AGENT_RECONNECT_MAX_BACKOFF_SECONDS` environment variables (the `bao agent run` command has no equivalent CLI flags).

#### spec.podTemplate

`spec.podTemplate` is an optional configuration (resources, nodeSelector, etc) for the spoke-agent Pod.

### VaultAgent Status

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

- `phase`: one of `Pending`, `Connected`, `Disconnected`, `Error`. The operator derives this from the spoke-agent Pod's readiness — `Connected` only once the Pod is Running and Ready.
- `podName`: name of the spoke-agent Pod.
- `appBindingRef`: references the AppBinding for the hub Vault (delivered by the hub `ManifestWork` in hub-managed deployments). Spoke-side consumers (SecretEngine, database role CRDs, KubeDB) use this AppBinding.
- `lastHeartbeat`: timestamp of the last successful heartbeat to the hub, if reported by the agent.
- `certExpiry`: expiry of the current spoke client certificate, if known.
- `conditions`: the latest available observations of the VaultAgent's state.

In hub-managed deployments, `status.phase` is scraped back to the hub through `ManifestWork` status feedback and aggregated into the hub VaultServer's `status.agentPlacement`.

## What the operator creates

For every `VaultAgent`, the spoke-side KubeVault operator provisions:

- a **ServiceAccount** for the agent Pod
- the **spoke-agent Pod**: with `spec.bootstrap`, an init container runs `bao agent join` (verifying the hub via the bootstrap token's JWS signature plus the SPKI pin) and writes credentials to an emptyDir; with `spec.tls`, the pre-provisioned certificates are projected read-only and there is no init container. The main container runs `bao agent run -server=<hub>:<grpcPort> -credentials-dir=...`. The Pod is replaced when its template changes (image, podTemplate, resources, or the referenced Secrets).
- an **AppBinding** for the hub VaultServer (standalone agents only). Its parameters carry `deploymentMode: RemoteAgent` and `spokeName`, which the secret engine machinery uses to route database mounts through the hub's `remote-<db>-plugin` proxies. In hub-managed deployments this AppBinding is instead delivered and owned by the hub's `ManifestWork` (labeled `app.kubernetes.io/managed-by: kubevault-hub`), and the operator defers to that copy. See the [AppBinding concept](/docs/concepts/vault-server-crds/appbinding.md).

## Next Steps

- Deploy the full hub-spoke model with OCM: [guide](/docs/guides/hub-spoke/deploy-hub-spoke.md).
- Learn about the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD and its `spec.agentPlacementRef` field.
