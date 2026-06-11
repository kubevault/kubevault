---
title: Manage Weaviate credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-weaviate
    name: Overview
    parent: weaviate-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Weaviate credentials using the KubeVault operator

OpenBao's [`weaviate-database-plugin`](https://github.com/sigilr/openbao/pull/18) is a **static-credentials-only** database plugin. [Weaviate](https://weaviate.io/developers/weaviate/configuration/authentication) loads its API keys from the `AUTHENTICATION_APIKEY_ALLOWED_KEYS` environment variable at server startup and exposes no runtime user-management API, so the plugin cannot create or delete users on demand. Instead, it probes the Weaviate HTTP `/v1/.well-known/ready` endpoint with the configured key sent as a Bearer token to verify reachability and returns "dynamic credentials are not supported" for `bao read database/creds/<role>`. KubeVault treats Weaviate like a static-credentials engine: you provision the Weaviate API key out of band (via the server environment variable), then use [`WeaviateRole`](/docs/concepts/secret-engine-crds/database-secret-engine/weaviaterole.md) to attach rotation metadata to that pre-existing key.

The same CRD shape is used both for the in-process `weaviate-database-plugin` and for the hub-spoke `remote-weaviate-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-weaviate-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [WeaviateRole](/docs/concepts/secret-engine-crds/database-secret-engine/weaviaterole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Provision a Weaviate deployment with [API-key authentication enabled](https://weaviate.io/developers/weaviate/configuration/authentication) by setting `AUTHENTICATION_APIKEY_ENABLED=true` and `AUTHENTICATION_APIKEY_ALLOWED_KEYS=<your-api-key>` on the server. The plugin only uses the HTTP endpoint and the `/v1/.well-known/ready` probe (key sent as a Bearer token) for a reachability check.
- Decide on the API key in advance; KubeVault does not generate or rotate the Weaviate-side key — it only writes the OpenBao-side static-roles metadata for an already-existing key.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Weaviate

Create an `AppBinding` pointing at the Weaviate HTTP endpoint. The secret's `password` field carries the Weaviate API key — KubeVault forwards it to the plugin as `api_key=`, which the plugin then sends as a Bearer token against `/v1/.well-known/ready` (Weaviate does not use HTTP Basic Auth).

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: weaviate
  namespace: demo
spec:
  clientConfig:
    url: http://weaviate.demo.svc:8080
  secret:
    name: weaviate-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: weaviate-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: bao
  password: <weaviate-api-key>
```

> The `username` field is required by the `kubernetes.io/basic-auth` Secret type but is ignored by the Weaviate plugin; only `password` (forwarded as `api_key`) is consumed.

## Enable and configure the Weaviate secret engine

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: weaviate-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  weaviate:
    databaseRef:
      name: weaviate
      namespace: demo
    pluginName: weaviate-database-plugin
    allowedRoles:
    - "*"
```

Apply and wait for the engine to land:

```bash
$ kubectl apply -f weaviate-secret-engine.yaml
secretengine.engine.kubevault.com/weaviate-engine created
$ kubectl get secretengines -n demo
NAME              STATUS    AGE
weaviate-engine   Success   10s
```

Behind the scenes the KubeVault operator writes:

```
bao write database/config/k8s.<cluster>.demo.weaviate \
    plugin_name=weaviate-database-plugin \
    url=http://weaviate.demo.svc:8080 \
    api_key=<weaviate-api-key> \
    allowed_roles="*"
```

When the referenced AppBinding is `deploymentMode: RemoteAgent`, the operator substitutes `plugin_name=remote-weaviate-plugin` and adds `spoke_name=<spoke>` so the hub forwards the call to the matching `bao agent run` daemon. See the [remote-db-plugin DESIGN](https://github.com/sigilr/openbao/blob/db-plugin-weaviate/plugins/database/remote-db-plugin/DESIGN.md) for the hub-spoke flow.

## Create a WeaviateRole

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: WeaviateRole
metadata:
  name: app
  namespace: demo
spec:
  secretEngineRef:
    name: weaviate-engine
  defaultTTL: 24h
  maxTTL: 168h
```

```bash
$ kubectl apply -f weaviate-role.yaml
weaviaterole.engine.kubevault.com/app created
$ kubectl get weaviaterole -n demo
NAME   STATUS    AGE
app    Success   8s
```

The KubeVault operator writes the role metadata as `database/roles/k8s.<cluster>.demo.app`. Weaviate's dynamic creds endpoint will still return the documented "dynamic credentials are not supported" error — that's the plugin contract; pair the `WeaviateRole` with `bao write database/static-roles/<name>` and a pre-existing Weaviate API key.

## Rotate static credentials

Configure the static-roles binding directly against OpenBao (the static-role API is not currently exposed via a KubeVault CRD):

```bash
$ bao write database/static-roles/k8s.<cluster>.demo.app \
    db_name=k8s.<cluster>.demo.weaviate \
    username=APP \
    rotation_period=24h
```

`bao read database/static-creds/k8s.<cluster>.demo.app` returns the latest rotated API key. KubeVault revokes the role on `WeaviateRole` deletion via the standard finalizer path.

## Cleanup

```bash
$ kubectl delete weaviaterole -n demo app
$ kubectl delete secretengine -n demo weaviate-engine
```

## Caveats

- **No dynamic credentials.** `bao read database/creds/<role>` always returns "dynamic credentials are not supported" — Weaviate has no runtime user-management API.
- **Out-of-band key management.** Weaviate loads its API keys from `AUTHENTICATION_APIKEY_ALLOWED_KEYS` at server startup; the plugin rotates the key via `bao` audit but cannot apply the new value to the Weaviate server without an environment reload (e.g. restart with the rotated key wired into the Pod env).
- **Insecure flag.** `spec.weaviate.insecure=true` disables TLS verification when probing the Weaviate HTTP endpoint. Use only in dev.
