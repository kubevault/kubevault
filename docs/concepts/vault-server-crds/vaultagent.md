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
2. **Standalone**: create the `VaultAgent` by hand in the spoke cluster, supplying the join material yourself.

When a `VaultAgent` is created, the KubeVault operator in the spoke cluster provisions a ServiceAccount, the spoke-agent Pod, and an AppBinding pointing back at the hub VaultServer (unless a hub-authored AppBinding already exists).

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
- `caBundle`: optional PEM bundle used to verify TLS on the hub Vault API endpoint. Takes precedence over the `caBundle` key of `spec.bootstrap.joinSecretRef`.

#### spec.bootstrap

`spec.bootstrap` is an optional field that configures the automated `bao agent join` trust bootstrap. When set, the spoke-agent Pod runs a join init container that exchanges a hub-issued bootstrap token for mTLS client credentials before the long-running agent starts. Credentials live on an emptyDir; the agent renews its own certificate in place, and a Pod restart simply re-joins with the current token.

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

Exactly one of `spec.bootstrap` or `spec.tls.certSecret` (pre-provisioned credentials) should be used.

#### spec.tokenSecretRef

`spec.tokenSecretRef` is an optional field referencing a Secret with a `token` key holding a Vault token. When set, the AppBinding created for this agent authenticates to the hub Vault with that token instead of kubernetes auth.

#### spec.image

`spec.image` is an optional field that overrides the spoke-agent container image.

#### spec.tls

`spec.tls` is an optional field for pre-provisioned mTLS credentials, as an alternative to `spec.bootstrap`:

- `caSecret`: Secret with `ca.crt` (the hub's spoke-CA certificate).
- `certSecret`: Secret with `tls.crt` and `tls.key` (a client certificate whose CN equals `spec.spokeName`, signed by the spoke-CA).

#### spec.reconnect

`spec.reconnect` is an optional field controlling automatic reconnection to the hub. Defaults: enabled, 5s initial backoff, 300s max backoff.

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

- `phase`: one of `Pending`, `Connected`, `Disconnected`, `Error`.
- `podName`: name of the spoke-agent Pod.
- `appBindingRef`: the AppBinding created for the hub Vault. Spoke-side consumers (SecretEngine, database role CRDs, KubeDB) use this AppBinding.
- `lastHeartbeat`: timestamp of the last successful heartbeat to the hub.
- `certExpiry`: expiry of the current spoke client certificate, if known.
- `conditions`: the latest available observations of the VaultAgent's state.

In hub-managed deployments, `status.phase` is scraped back to the hub through `ManifestWork` status feedback and aggregated into the hub VaultServer's `status.agentPlacement`.

## What the operator creates

For every `VaultAgent`, the spoke-side KubeVault operator provisions:

- a **ServiceAccount** for the agent Pod
- the **spoke-agent Pod**: with `spec.bootstrap` set, an init container runs `bao agent join` (verifying the hub via the bootstrap token's JWS signature plus the SPKI pin) and writes credentials to an emptyDir; the main container runs `bao agent run -server=<hub>:<grpcPort> -credentials-dir=...`
- an **AppBinding** for the hub VaultServer, unless a hub-authored AppBinding (label `app.kubernetes.io/managed-by: kubevault-hub`) already exists. The AppBinding parameters carry `deploymentMode: RemoteAgent` and `spokeName`, which the secret engine machinery uses to route database mounts through the hub's `remote-<db>-plugin` proxies. See the [AppBinding concept](/docs/concepts/vault-server-crds/appbinding.md).

## Next Steps

- Deploy the full hub-spoke model with OCM: [guide](/docs/guides/hub-spoke/deploy-hub-spoke.md).
- Learn about the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD and its `spec.agentPlacementRef` field.
