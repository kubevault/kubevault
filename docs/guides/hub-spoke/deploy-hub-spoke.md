---
title: Deploy Vault in a Hub-Spoke Model | KubeVault
menu:
  docs_{{ .version }}:
    identifier: deploy-hub-spoke
    name: Deploy Hub-Spoke
    parent: hub-spoke-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Deploy Vault in a Hub-Spoke Model

This guide shows how to run one central Vault (OpenBao) server on a **hub** cluster and let workloads in many **spoke** clusters consume database credentials from it, even when the spoke databases are not reachable from the hub. KubeVault automates the whole rollout with [Open Cluster Management](https://open-cluster-management.io/) (OCM): you select spoke clusters with a `Placement`, and the operator does the rest.

## How it works

```
HUB CLUSTER                                       SPOKE CLUSTER (one of many)
+--------------------------------------+          +--------------------------------------+
|  VaultServer (OpenBao, HA)           |          |  VaultRelay (delivered via OCM)      |
|   - relay/ backend: spoke-CA, tokens |          |   - join init container (bootstrap)  |
|   - remote-<db>-plugin proxies       |  mTLS    |   - bao relay run daemon             |
|  Service (LoadBalancer)              |<==gRPC===|   - runs db plugins in-process       |
|   - 8200 Vault API                   |  :50053  |       |                              |
|   - 50053 spoke gRPC proxy           |<--HTTPS--|       +--> postgres (spoke-local)    |
|  Placement -> PlacementDecision      |  :8200   |  AppBinding (deploymentMode:         |
|  ManifestWork per selected cluster   |          |     RemoteRelay) for hub Vault       |
+--------------------------------------+          +--------------------------------------+
```

When you set `spec.relayPlacementRef` on the hub `VaultServer`, the operator:

1. initializes the OpenBao `relay/` backend (spoke-CA, gRPC proxy listener) and advertises the LoadBalancer address,
2. resolves the `Placement` through its `PlacementDecision`s to a set of managed clusters,
3. per cluster: creates a hub ServiceAccount (in the managed cluster's namespace on the hub) with a `VaultPolicy`/`VaultPolicyBinding` scoping its access, mints a rotating bootstrap token, and applies one `ManifestWork` carrying the `VaultRelay`, the AppBinding (with the LoadBalancer address and CA bundle), and the credential Secrets,
4. aggregates rollout state back into `status.relayPlacement` using ManifestWork status feedback.

On each spoke, the KubeVault operator reconciles the delivered `VaultRelay`: an init container runs `bao relay join` (verifying the hub with the bootstrap token's JWS signature plus the spoke-CA SPKI pin) and the main container runs `bao relay run`, connecting back to the hub over mTLS.

## Before you begin

- A hub cluster with the [OCM hub components](https://open-cluster-management.io/docs/getting-started/installation/start-the-control-plane/) installed (`clusteradm init`), and one or more managed clusters [registered](https://open-cluster-management.io/docs/getting-started/installation/register-a-cluster/) (`clusteradm join` + accept).
- The KubeVault operator installed on the hub **and** on every spoke cluster (the spoke operator reconciles the delivered `VaultRelay`). See the [setup guide](/docs/setup/README.md).
- [cert-manager](https://cert-manager.io/docs/installation/) on the hub (the VaultServer must run with TLS).
- A cloud LoadBalancer (or equivalent) on the hub cluster, reachable from the spokes.

In this guide the hub namespace is `demo` and the managed cluster is named `cluster-1`.

```bash
$ kubectl create namespace demo
```

## Step 1: Deploy the hub VaultServer

Create an Issuer for the VaultServer TLS certificates:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: vault-issuer
  namespace: demo
spec:
  selfSigned: {}
```

Create the `VaultServer` with a LoadBalancer service template and the agent placement reference:

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: 1.10.3
  replicas: 3
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: vault-issuer
  backend:
    raft:
      storage:
        storageClassName: "standard"
        resources:
          requests:
            storage: 1Gi
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  serviceTemplates:
    - alias: vault
      spec:
        type: LoadBalancer
  relayPlacementRef:
    name: db-spokes
  relayTemplate:
    bootstrapTokenTTL: 24h
  terminationPolicy: WipeOut
```

The `vault` Service exposes the Vault API on port 8200 and the spoke-relay gRPC proxy on port 50053; both travel through the same LoadBalancer.

## Step 2: Select spoke clusters with a Placement

A `Placement` selects from the `ManagedClusterSet`s bound to its namespace. Bind a cluster set to `demo` and create the Placement referenced above:

```yaml
apiVersion: cluster.open-cluster-management.io/v1beta2
kind: ManagedClusterSetBinding
metadata:
  name: default
  namespace: demo
spec:
  clusterSet: default
---
apiVersion: cluster.open-cluster-management.io/v1beta1
kind: Placement
metadata:
  name: db-spokes
  namespace: demo
spec:
  clusterSets:
    - default
  predicates:
    - requiredClusterSelector:
        labelSelector:
          matchLabels:
            purpose: database
```

Label the managed clusters that should run a spoke relay:

```bash
$ kubectl label managedcluster cluster-1 purpose=database
```

## Step 3: Watch the rollout

The operator waits for the VaultServer to be unsealed and the LoadBalancer address to be provisioned, then rolls out the spokes. Track progress on the hub:

```bash
$ kubectl get vaultserver vault -n demo -o jsonpath='{.status.relayPlacement}' | jq
{
  "placement": "db-spokes",
  "selected": 1,
  "applied": 1,
  "ready": 1,
  "clusters": [
    {
      "clusterName": "cluster-1",
      "phase": "Connected",
      "tokenExpiry": "2026-06-13T10:00:00Z",
      "certExpiry": "2026-07-12T10:00:00Z"
    }
  ]
}
```

The relevant condition types are `RelayPlacementResolved`, `RelayHubInitialized`, `RelayManifestWorksApplied`, and `RelaysReady`.

You can also inspect the per-cluster ManifestWork in the managed cluster's namespace on the hub:

```bash
$ kubectl get manifestwork -n cluster-1
NAME                 AGE
kv-demo-vault-agent  2m
```

And confirm the spoke is connected from inside a Vault pod:

```bash
$ kubectl exec -n demo vault-0 -c vault -- bao relay list
NAME       LAST SEEN  UPTIME  CERT EXP  HEALTH
cluster-1  1s ago     2m      29d       OK
```

## Step 4: Verify the spoke side

On the spoke cluster, the work agent has applied the payload and the KubeVault operator has started the relay:

```bash
$ kubectl get vaultrelay -n demo
NAME          SPOKE       STATUS      AGE
vault-agent   cluster-1   Connected   2m

$ kubectl get pods -n demo
NAME                READY   STATUS    RESTARTS   AGE
vault-agent-agent   1/1     Running   0          2m

$ kubectl get appbinding -n demo
NAME                    TYPE                 AGE
vault-agent-hub-vault   kubevault.com/vault  2m
```

The AppBinding points at the hub's LoadBalancer address and carries `deploymentMode: RemoteRelay` in its parameters. Spoke-side consumers authenticate to the hub Vault with a hub-issued ServiceAccount token via kubernetes auth; the matching Vault role and policy were created on the hub automatically.

## Step 5: Issue credentials for a spoke-local database

Now use the regular KubeVault secret engine workflow on the spoke, referencing the delivered AppBinding. The database itself only needs to be reachable from the spoke.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: postgres-engine
  namespace: demo
spec:
  vaultRef:
    name: vault-agent-hub-vault
    namespace: demo
  postgres:
    databaseRef:
      name: postgres-app
      namespace: demo
---
apiVersion: engine.kubevault.com/v1alpha1
kind: PostgresRole
metadata:
  name: postgres-readonly
  namespace: demo
spec:
  secretEngineRef:
    name: postgres-engine
  creationStatements:
    - CREATE ROLE "{{name}}" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';
    - GRANT SELECT ON ALL TABLES IN SCHEMA public TO "{{name}}";
  defaultTTL: 1h
```

Because the AppBinding's `deploymentMode` is `RemoteRelay`, the SecretEngine controller configures the hub mount with `plugin_name: remote-postgres-plugin` and `spoke_name: cluster-1`. The hub proxies every credential operation to the spoke relay, which runs the real `postgresql-database-plugin` in-process against the spoke-local database.

Postgres, MySQL, MariaDB, Redis, Valkey, and HANA are supported through the spoke relay. MongoDB and Elasticsearch are not; a SecretEngine for those against a `RemoteRelay` AppBinding is rejected on apply by the validating webhook.

Request credentials the usual way with a `SecretAccessRequest`; see the [secret engine guides](/docs/guides/secret-engines/postgres/overview.md).

## Day-2 operations

- **Adding or removing spokes**: label or relabel managed clusters; the Placement decision changes and the operator converges. Removed clusters get their ManifestWork deleted (the spoke loses the relay, AppBinding, and Secrets), their hub-side ServiceAccount and policies cleaned up, and their bootstrap token revoked. The spoke **Namespace is left in place** (orphaned) — it may hold non-KubeVault workloads — and the operator emits a `SpokeMountsRetained` warning Event since the hub-side database mounts the spoke configured are also retained.
- **Bootstrap token rotation**: automatic. Tokens default to a 24h TTL (`spec.relayTemplate.bootstrapTokenTTL`) and are rotated when less than a quarter of the TTL remains, so a restarting spoke Pod can always re-join.
- **Certificate renewal**: the spoke relay renews its own mTLS client certificate in place at half-life; no operator involvement. The current expiry is visible on `status.relayPlacement.clusters[].certExpiry` (the hub reads it from its `relay/spokes` endpoint, so it tracks renewals) and via `bao relay list`.
- **LoadBalancer address change**: the operator refreshes the advertised endpoint on the hub and pushes the new address into every ManifestWork. The changed hub address rolls the spoke-relay Pods (pod-template change), so they reconnect to the new endpoint.
- **VaultServer deletion**: under `Halt`, `Delete`, or `WipeOut`, hub-side finalizers tear down every ManifestWork (each spoke loses its VaultRelay, AppBinding, and Secrets) and revoke outstanding bootstrap tokens before the VaultServer itself is removed. A `DoNotTerminate` VaultServer cannot be deleted — the validating webhook rejects it, and even if that is bypassed the spoke-relay finalizer is retained and teardown is skipped, so the spokes keep running until the policy is changed.

## Troubleshooting

| Symptom | Check |
|---|---|
| `RelayPlacementResolved=False`, reason `WaitingForLoadBalancer` | the `vault` Service has no LoadBalancer ingress yet, or its type is not `LoadBalancer` — or, for non-cloud fleets, set the `kubevault.com/server-address` annotation to an external address instead |
| `RelayPlacementResolved=False`, placement errors | the Placement exists in the VaultServer namespace and a `ManagedClusterSetBinding` binds the cluster set to that namespace |
| `relayPlacementRef` silently ignored | OCM hub CRDs are not installed; the operator logs this at startup |
| ManifestWork `Degraded` with forbidden errors | the klusterlet work agent lacks permission for KubeVault CRs; the aggregation ClusterRole shipped in the ManifestWork requires OCM >= v0.12 |
| spoke Pod join init container failing | bootstrap token expired (check `status.relayPlacement.clusters[].tokenExpiry`) or the hub Vault API is unreachable from the spoke |
| VaultRelay `Connected` but SecretEngine fails | the engine type may not be supported through the spoke relay (MongoDB, Elasticsearch) |

## Cleanup

```bash
# hub
$ kubectl delete vaultserver vault -n demo     # cascades: ManifestWorks, hub SAs, policies, tokens
$ kubectl delete placement db-spokes -n demo
$ kubectl delete ns demo
```

## Next Steps

- [VaultRelay concept](/docs/concepts/vault-server-crds/vaultrelay.md)
- [VaultServer concept](/docs/concepts/vault-server-crds/vaultserver.md)
- [AppBinding concept](/docs/concepts/vault-server-crds/appbinding.md), in particular `spec.parameters.deploymentMode`
- [Secret engine guides](/docs/guides/secret-engines/postgres/overview.md)
- [Tenant Isolation with OpenBao Namespaces](/docs/guides/tenant-isolation/overview.md) — isolate each spoke's client-org databases into per-org OpenBao namespaces on the hub
