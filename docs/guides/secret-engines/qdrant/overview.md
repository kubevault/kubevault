---
title: Manage Qdrant credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-qdrant
    name: Overview
    parent: qdrant-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Qdrant credentials using the KubeVault operator

OpenBao's [`qdrant-database-plugin`](https://github.com/sigilr/openbao/pull/17) is a **static-credentials-only** database plugin. [Qdrant](https://qdrant.tech/documentation/guides/security/) loads its API key from the `QDRANT__SERVICE__API_KEY` environment variable at server startup and exposes no runtime user-management API, so the plugin cannot create or delete users on demand. Instead, it probes the Qdrant HTTP `/readyz` endpoint with the configured key sent in the `api-key` header to verify reachability and returns "dynamic credentials are not supported" for `bao read database/creds/<role>`. KubeVault treats Qdrant like a static-credentials engine: you provision the Qdrant API key out of band (via the server environment variable), then use [`QdrantRole`](/docs/concepts/secret-engine-crds/database-secret-engine/qdrantrole.md) to attach rotation metadata to that pre-existing key.

The same CRD shape is used both for the in-process `qdrant-database-plugin` and for the hub-spoke `remote-qdrant-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-qdrant-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [QdrantRole](/docs/concepts/secret-engine-crds/database-secret-engine/qdrantrole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Provision a Qdrant deployment with [API-key authentication enabled](https://qdrant.tech/documentation/guides/security/) by setting `QDRANT__SERVICE__API_KEY` on the server. The plugin only uses the HTTP endpoint and the `/readyz` probe (key sent in the `api-key` header) for a reachability check.
- Decide on the API key in advance; KubeVault does not generate or rotate the Qdrant-side key — it only writes the OpenBao-side static-roles metadata for an already-existing key.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Qdrant

Create an `AppBinding` pointing at the Qdrant HTTP endpoint. The secret's `password` field carries the Qdrant API key — KubeVault forwards it to the plugin as `api_key=`, which the plugin then sends in the `api-key` HTTP header against `/readyz` (Qdrant does not use HTTP Basic Auth).

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: qdrant
  namespace: demo
spec:
  clientConfig:
    url: http://qdrant.demo.svc:6333
  secret:
    name: qdrant-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: qdrant-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: bao
  password: <qdrant-api-key>
```

> The `username` field is required by the `kubernetes.io/basic-auth` Secret type but is ignored by the Qdrant plugin; only `password` (forwarded as `api_key`) is consumed.

## Enable and configure the Qdrant secret engine

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: qdrant-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  qdrant:
    databaseRef:
      name: qdrant
      namespace: demo
    pluginName: qdrant-database-plugin
    allowedRoles:
    - "*"
```

Apply and wait for the engine to land:

```bash
$ kubectl apply -f qdrant-secret-engine.yaml
secretengine.engine.kubevault.com/qdrant-engine created
$ kubectl get secretengines -n demo
NAME            STATUS    AGE
qdrant-engine   Success   10s
```

Behind the scenes the KubeVault operator writes:

```
bao write database/config/k8s.<cluster>.demo.qdrant \
    plugin_name=qdrant-database-plugin \
    url=http://qdrant.demo.svc:6333 \
    api_key=<qdrant-api-key> \
    allowed_roles="*"
```

When the referenced AppBinding is `deploymentMode: RemoteAgent`, the operator substitutes `plugin_name=remote-qdrant-plugin` and adds `spoke_name=<spoke>` so the hub forwards the call to the matching `bao agent run` daemon. See the [remote-db-plugin DESIGN](https://github.com/sigilr/openbao/blob/db-plugin-qdrant/plugins/database/remote-db-plugin/DESIGN.md) for the hub-spoke flow.

## Create a QdrantRole

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: QdrantRole
metadata:
  name: app
  namespace: demo
spec:
  secretEngineRef:
    name: qdrant-engine
  defaultTTL: 24h
  maxTTL: 168h
```

```bash
$ kubectl apply -f qdrant-role.yaml
qdrantrole.engine.kubevault.com/app created
$ kubectl get qdrantrole -n demo
NAME   STATUS    AGE
app    Success   8s
```

The KubeVault operator writes the role metadata as `database/roles/k8s.<cluster>.demo.app`. Qdrant's dynamic creds endpoint will still return the documented "dynamic credentials are not supported" error — that's the plugin contract; pair the `QdrantRole` with `bao write database/static-roles/<name>` and a pre-existing Qdrant API key.

## Rotate static credentials

Configure the static-roles binding directly against OpenBao (the static-role API is not currently exposed via a KubeVault CRD):

```bash
$ bao write database/static-roles/k8s.<cluster>.demo.app \
    db_name=k8s.<cluster>.demo.qdrant \
    username=APP \
    rotation_period=24h
```

`bao read database/static-creds/k8s.<cluster>.demo.app` returns the latest rotated API key. KubeVault revokes the role on `QdrantRole` deletion via the standard finalizer path.

## Cleanup

```bash
$ kubectl delete qdrantrole -n demo app
$ kubectl delete secretengine -n demo qdrant-engine
```

## Caveats

- **No dynamic credentials.** `bao read database/creds/<role>` always returns "dynamic credentials are not supported" — Qdrant has no runtime user-management API.
- **Out-of-band key management.** Qdrant loads its API key from `QDRANT__SERVICE__API_KEY` at server startup; the plugin rotates the key via `bao` audit but cannot apply the new value to the Qdrant server without an environment reload (e.g. restart with the rotated key wired into the Pod env).
- **Insecure flag.** `spec.qdrant.insecure=true` disables TLS verification when probing the Qdrant HTTP endpoint. Use only in dev.
