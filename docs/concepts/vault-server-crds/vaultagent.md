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
2. **Standalone**: create the `VaultAgent` by hand in the spoke cluster. You must set `spec.bootstrap` and supply the join Secret it references yourself (see [spec.bootstrap](#specbootstrap)).

`spec.bootstrap` is required: the spoke agent obtains its mTLS credentials by running `bao agent join`. When a `VaultAgent` is reconciled, the KubeVault operator in the spoke cluster provisions a ServiceAccount and the spoke-agent Pod. The AppBinding pointing back at the hub VaultServer is **not** authored by this reconciler — in hub-managed deployments it is delivered and owned by the hub's `ManifestWork`.

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

`spec.bootstrap` configures the automated `bao agent join` trust bootstrap. The spoke-agent Pod runs a join init container that exchanges a hub-issued bootstrap token for mTLS client credentials before the long-running agent starts. Credentials live on an emptyDir; the agent renews its own certificate in place, and a Pod restart simply re-joins with the current token.

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

> **Note:** `spec.bootstrap` is currently required. The operator rejects a `VaultAgent` that does not set it, because the spoke agent has no other supported way to obtain its mTLS credentials.

#### spec.image

`spec.image` is an optional field that overrides the spoke-agent container image.

#### spec.tokenSecretRef

`spec.tokenSecretRef` is an optional field referencing a Secret with a `token` key holding a Vault token.

> **Note:** Reserved for future use. The current operator does not author the spoke AppBinding (it is delivered by the hub `ManifestWork`), so this field has no effect today.

#### spec.tls

`spec.tls` describes pre-provisioned mTLS credentials as an alternative to the `spec.bootstrap` join flow:

- `caSecret`: Secret with `ca.crt` (the hub's spoke-CA certificate).
- `certSecret`: Secret with `tls.crt` and `tls.key` (a client certificate whose CN equals `spec.spokeName`, signed by the spoke-CA).

> **Note:** Not yet implemented. The current operator requires `spec.bootstrap`; pre-provisioned credentials are not consumed yet.

#### spec.reconnect

`spec.reconnect` describes automatic reconnection to the hub (defaults: enabled, 5s initial backoff, 300s max backoff).

> **Note:** Reserved. These fields are not yet wired into the spoke-agent Pod; reconnection is handled by the agent's own defaults.

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
- the **spoke-agent Pod**: an init container runs `bao agent join` (verifying the hub via the bootstrap token's JWS signature plus the SPKI pin) and writes credentials to an emptyDir; the main container runs `bao agent run -server=<hub>:<grpcPort> -credentials-dir=...`. The Pod is replaced when its template changes (image, podTemplate, resources, or the referenced join Secret).

The **AppBinding** for the hub VaultServer is **not** authored by this reconciler. In hub-managed deployments it is delivered and owned by the hub's `ManifestWork` (labeled `app.kubernetes.io/managed-by: kubevault-hub`). Its parameters carry `deploymentMode: RemoteAgent` and `spokeName`, which the secret engine machinery uses to route database mounts through the hub's `remote-<db>-plugin` proxies. See the [AppBinding concept](/docs/concepts/vault-server-crds/appbinding.md).

## Next Steps

- Deploy the full hub-spoke model with OCM: [guide](/docs/guides/hub-spoke/deploy-hub-spoke.md).
- Learn about the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD and its `spec.agentPlacementRef` field.
