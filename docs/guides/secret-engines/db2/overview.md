---
title: Manage IBM Db2 credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-db2
    name: Overview
    parent: db2-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage IBM Db2 credentials using the KubeVault operator

OpenBao's [`db2-database-plugin`](https://github.com/sigilr/openbao/pull/19) is a **static-credentials-only** database plugin. There is no pure-Go Db2 driver and most production deployments delegate user management to the OS or LDAP, so the plugin does not support dynamic `bao read database/creds/<role>` calls. Instead, KubeVault treats Db2 like a static-credentials engine: you provision the Db2 principal out of band, then use [`DB2Role`](/docs/concepts/secret-engine-crds/database-secret-engine/db2role.md) to attach rotation metadata to that pre-existing principal.

The same CRD shape is used both for the in-process `db2-database-plugin` and for the hub-spoke `remote-db2-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-db2-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [DB2Role](/docs/concepts/secret-engine-crds/database-secret-engine/db2role.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Provision a Db2 deployment that exposes the [Db2 REST API](https://www.ibm.com/docs/en/db2/11.5?topic=apis-rest) (`/dbapi/v4/host_status`). The plugin only uses this endpoint for an optional reachability check.
- Provision a Db2 principal (via `AUTH_NATIVE`, OS, or LDAP) for each role you want to rotate. KubeVault will not create the principal.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Db2

Create an `AppBinding` pointing at the Db2 REST endpoint. The username/password Secret (if present) feeds the plugin's optional Basic Auth healthcheck.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: db2
  namespace: demo
spec:
  clientConfig:
    url: http://db2-dbapi.demo.svc:50000
  secret:
    name: db2-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: db2-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: bao
  password: <strong-password>
```

## Enable and configure the Db2 secret engine

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: db2-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  db2:
    databaseRef:
      name: db2
      namespace: demo
    pluginName: db2-database-plugin
    allowedRoles:
    - "*"
```

Apply and wait for the engine to land:

```bash
$ kubectl apply -f db2-secret-engine.yaml
secretengine.engine.kubevault.com/db2-engine created
$ kubectl get secretengines -n demo
NAME         STATUS    AGE
db2-engine   Success   10s
```

Behind the scenes the KubeVault operator writes:

```
bao write database/config/k8s.<cluster>.demo.db2 \
    plugin_name=db2-database-plugin \
    url=http://db2-dbapi.demo.svc:50000 \
    username=bao \
    password=<strong-password> \
    allowed_roles="*"
```

When the referenced AppBinding is `deploymentMode: RemoteAgent`, the operator substitutes `plugin_name=remote-db2-plugin` and adds `spoke_name=<spoke>` so the hub forwards the call to the matching `bao agent run` daemon. See the [remote-db-plugin DESIGN](https://github.com/sigilr/openbao/blob/db-plugin-db2/plugins/database/remote-db-plugin/DESIGN.md) for the hub-spoke flow.

## Create a DB2Role

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: DB2Role
metadata:
  name: app
  namespace: demo
spec:
  secretEngineRef:
    name: db2-engine
  defaultTTL: 24h
  maxTTL: 168h
```

```bash
$ kubectl apply -f db2-role.yaml
db2role.engine.kubevault.com/app created
$ kubectl get db2role -n demo
NAME   STATUS    AGE
app    Success   8s
```

The KubeVault operator writes the role metadata as `database/roles/k8s.<cluster>.demo.app`. Db2's dynamic creds endpoint will still return the documented "dynamic credentials are not supported" error — that's the plugin contract; pair the `DB2Role` with `bao write database/static-roles/<name>` and a pre-existing Db2 principal.

## Rotate static credentials

Configure the static-roles binding directly against OpenBao (the static-role API is not currently exposed via a KubeVault CRD):

```bash
$ bao write database/static-roles/k8s.<cluster>.demo.app \
    db_name=k8s.<cluster>.demo.db2 \
    username=APP \
    rotation_period=24h
```

`bao read database/static-creds/k8s.<cluster>.demo.app` returns the latest rotated password. KubeVault revokes the role on `DB2Role` deletion via the standard finalizer path.

## Cleanup

```bash
$ kubectl delete db2role -n demo app
$ kubectl delete secretengine -n demo db2-engine
```

## Caveats

- **No dynamic credentials.** `bao read database/creds/<role>` always returns "dynamic credentials are not supported".
- **Out-of-band principal management.** Provision Db2 principals via your configuration management; the plugin rotates passwords via `bao` audit but cannot apply them to the Db2 auth source.
- **Insecure flag.** `spec.db2.insecure=true` disables TLS verification against the Db2 REST endpoint. Use only in dev.
