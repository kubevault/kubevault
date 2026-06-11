---
title: Manage Memcached credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-memcached
    name: Overview
    parent: memcached-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Memcached credentials using the KubeVault operator

OpenBao's [`memcached-database-plugin`](https://github.com/sigilr/openbao/pull/16) is a **static-credentials-only** database plugin. [Memcached](https://docs.memcached.org/) loads SASL credentials from a static auth file at server startup and exposes no runtime user-management API, so the plugin cannot create or delete users on demand. Instead, it pings the Memcached TCP endpoint (and optionally completes a TLS handshake) to verify reachability and returns "dynamic credentials are not supported" for `bao read database/creds/<role>`. KubeVault treats Memcached like a static-credentials engine: you provision the Memcached principal out of band (in the SASL auth file), then use [`MemcachedRole`](/docs/concepts/secret-engine-crds/database-secret-engine/memcachedrole.md) to attach rotation metadata to that pre-existing principal.

The same CRD shape is used both for the in-process `memcached-database-plugin` and for the hub-spoke `remote-memcached-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-memcached-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [MemcachedRole](/docs/concepts/secret-engine-crds/database-secret-engine/memcachedrole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Provision a Memcached deployment with [SASL authentication enabled](https://docs.memcached.org/features/authentication/) and the SASL auth file populated at server startup. The plugin only uses the TCP endpoint (and an optional TLS handshake) for a reachability check.
- Configure each Memcached principal in the SASL auth file for every role you want to rotate. KubeVault will not create the principal.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Memcached

Create an `AppBinding` pointing at the Memcached TCP endpoint. The username/password Secret (if present) feeds the plugin's Basic Auth check against the SASL endpoint.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: memcached
  namespace: demo
spec:
  clientConfig:
    url: tcp://memcached.demo.svc:11211
  secret:
    name: memcached-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: memcached-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: bao
  password: <strong-password>
```

## Enable and configure the Memcached secret engine

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: memcached-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  memcached:
    databaseRef:
      name: memcached
      namespace: demo
    pluginName: memcached-database-plugin
    allowedRoles:
    - "*"
```

Apply and wait for the engine to land:

```bash
$ kubectl apply -f memcached-secret-engine.yaml
secretengine.engine.kubevault.com/memcached-engine created
$ kubectl get secretengines -n demo
NAME               STATUS    AGE
memcached-engine   Success   10s
```

Behind the scenes the KubeVault operator writes:

```
bao write database/config/k8s.<cluster>.demo.memcached \
    plugin_name=memcached-database-plugin \
    url=tcp://memcached.demo.svc:11211 \
    username=bao \
    password=<strong-password> \
    allowed_roles="*"
```

When the referenced AppBinding is `deploymentMode: RemoteAgent`, the operator substitutes `plugin_name=remote-memcached-plugin` and adds `spoke_name=<spoke>` so the hub forwards the call to the matching `bao agent run` daemon. See the [remote-db-plugin DESIGN](https://github.com/sigilr/openbao/blob/db-plugin-memcached/plugins/database/remote-db-plugin/DESIGN.md) for the hub-spoke flow.

## Create a MemcachedRole

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: MemcachedRole
metadata:
  name: app
  namespace: demo
spec:
  secretEngineRef:
    name: memcached-engine
  defaultTTL: 24h
  maxTTL: 168h
```

```bash
$ kubectl apply -f memcached-role.yaml
memcachedrole.engine.kubevault.com/app created
$ kubectl get memcachedrole -n demo
NAME   STATUS    AGE
app    Success   8s
```

The KubeVault operator writes the role metadata as `database/roles/k8s.<cluster>.demo.app`. Memcached's dynamic creds endpoint will still return the documented "dynamic credentials are not supported" error — that's the plugin contract; pair the `MemcachedRole` with `bao write database/static-roles/<name>` and a pre-existing Memcached principal.

## Rotate static credentials

Configure the static-roles binding directly against OpenBao (the static-role API is not currently exposed via a KubeVault CRD):

```bash
$ bao write database/static-roles/k8s.<cluster>.demo.app \
    db_name=k8s.<cluster>.demo.memcached \
    username=APP \
    rotation_period=24h
```

`bao read database/static-creds/k8s.<cluster>.demo.app` returns the latest rotated password. KubeVault revokes the role on `MemcachedRole` deletion via the standard finalizer path.

## Cleanup

```bash
$ kubectl delete memcachedrole -n demo app
$ kubectl delete secretengine -n demo memcached-engine
```

## Caveats

- **No dynamic credentials.** `bao read database/creds/<role>` always returns "dynamic credentials are not supported" — Memcached has no runtime user-management API.
- **Out-of-band principal management.** Memcached principals live in the SASL auth file at server startup; the plugin rotates passwords via `bao` audit but cannot apply them to the Memcached server without a config reload.
- **Insecure flag.** `spec.memcached.insecure=true` disables TLS verification when probing the Memcached TCP endpoint. Use only in dev.
