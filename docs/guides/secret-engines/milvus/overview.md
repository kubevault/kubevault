---
title: Manage Milvus credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-milvus
    name: Overview
    parent: milvus-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Milvus credentials using the KubeVault operator

OpenBao's [`milvus-database-plugin`](https://github.com/sigilr/openbao/pull/13) is a **dynamic-credentials** database plugin for [Milvus](https://milvus.io/). The plugin provisions credentials by talking to Milvus's [HTTP RESTful API v2 user-management endpoints](https://milvus.io/docs/users_and_roles.md): each issued credential becomes a Milvus user and is bound to one or more pre-existing roles on the target cluster. The plugin **does not create roles**; it only manages users and their role bindings, so every role referenced from `creationStatements` must already exist on the Milvus cluster.

The same CRD shape is used both for the in-process `milvus-database-plugin` and for the hub-spoke `remote-milvus-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-milvus-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [MilvusRole](/docs/concepts/secret-engine-crds/database-secret-engine/milvusrole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Run a Milvus cluster reachable over HTTP. The Milvus [standalone quickstart](https://milvus.io/docs/install_standalone-docker.md) exposes the HTTP endpoint at `19530`; managed Milvus on [Zilliz Cloud](https://zilliz.com/cloud) exposes an `https://...` URL with token-based auth.
- Pre-create the Milvus roles you want to bind credentials to (e.g. `dba`, `readonly`). The plugin only binds — it does not create roles. See [Users and Roles](https://milvus.io/docs/users_and_roles.md) for the Milvus role model.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Milvus

Create an `AppBinding` pointing at the Milvus HTTP endpoint. Unlike most database engines, the URL here is **not** a JDBC/connection URI — it is the HTTP(S) base URL of Milvus (e.g. `http://milvus.demo.svc:19530`). The referenced Secret carries either HTTP Basic Auth credentials (`username` + `password`) **or** a Bearer token (`token`) for Zilliz Cloud / API-token style auth. When the secret carries a `token` key, the operator forwards it as `token=` to the plugin **instead of** username/password.

### Self-hosted Milvus (HTTP Basic)

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: milvus
  namespace: demo
spec:
  clientConfig:
    url: http://milvus.demo.svc:19530
  secret:
    name: milvus-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: milvus-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: root
  password: Milvus
```

### Zilliz Cloud (Bearer token)

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: milvus
  namespace: demo
spec:
  clientConfig:
    url: https://in03-xxxxxxxxxxxxxxx.api.gcp-us-west1.zillizcloud.com
  secret:
    name: milvus-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: milvus-cred
  namespace: demo
type: Opaque
stringData:
  token: db_admin:zilliz-api-token-here
```

> If you front Milvus with a self-signed TLS cert (e.g. the standalone quickstart with TLS enabled), set `SecretEngine.spec.milvus.insecure: true` below. Drop the knob once you front Milvus with a real CA-issued certificate.

## Enable and Configure Milvus Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on a specified path and configure the secret engine with the given configuration.

A sample `SecretEngine` for Milvus:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: milvus-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  milvus:
    databaseRef:
      name: milvus
      namespace: demo
    pluginName: milvus-database-plugin   # optional; this is the default
    allowedRoles:
      - "*"
    dbName: default                      # optional; forwarded as the `dbName` request header on every API call
    insecure: false                      # set true only for self-signed dev clusters
```

Apply it and wait for `STATUS=Success`:

```bash
$ kubectl apply -f milvus-engine.yaml
secretengine.engine.kubevault.com/milvus-engine created

$ kubectl get secretengines -n demo
NAME            STATUS    AGE
milvus-engine   Success   10s
```

Use `kubectl describe secretengine -n demo milvus-engine` to inspect error events, if any.

## Create a MilvusRole

A [`MilvusRole`](/docs/concepts/secret-engine-crds/database-secret-engine/milvusrole.md) describes how the plugin should mint a dynamic credential. `creationStatements` is a single-element string slice holding a JSON role document of the form `{"roles":["role1","role2"]}`. The listed roles **must already exist** on the target Milvus cluster — the plugin only binds, it does not create roles.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: MilvusRole
metadata:
  name: milvus-readonly
  namespace: demo
spec:
  secretEngineRef:
    name: milvus-engine
  creationStatements:
    - '{"roles":["dba","readonly"]}'
  defaultTTL: 1h
  maxTTL: 24h
```

Apply and verify:

```bash
$ kubectl apply -f milvus-role.yaml
milvusrole.engine.kubevault.com/milvus-readonly created

$ kubectl get milvusrole -n demo
NAME              STATUS    AGE
milvus-readonly   Success   12s
```

The role name in Vault follows the format `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`, so you can verify directly with the Vault CLI:

```bash
$ vault read your-database-path/roles/k8s.-.demo.milvus-readonly
Key                      Value
---                      -----
creation_statements      [{"roles":["dba","readonly"]}]
db_name                  k8s.-.demo.milvus
default_ttl              1h
max_ttl                  24h
```

Deleting the `MilvusRole` removes the role from Vault.

## Issue Milvus credentials

Request a dynamic credential by creating a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: milvus-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: MilvusRole
    name: milvus-readonly
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Approve it through the KubeVault CLI:

```bash
$ kubectl vault approve secretaccessrequest milvus-cred-rqst -n demo
approved
```

Once approved, the operator issues the credential, stores it in a `Secret`, and binds the listed subjects via a `Role`/`RoleBinding`. The plugin creates a new Milvus user and grants the listed roles (`dba`, `readonly`) on the target Milvus cluster. The credential lives on the lease until you delete the `SecretAccessRequest` or it expires; on lease revocation the plugin removes the Milvus user.

```bash
$ kubectl get secretaccessrequest milvus-cred-rqst -n demo -o json | jq '.status'
{
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.milvus-readonly/abc...",
    "renewable": true
  },
  "secret": {
    "name": "milvus-cred-rqst-xxxxxx"
  }
}

$ kubectl get secret -n demo milvus-cred-rqst-xxxxxx -o jsonpath='{.data.username}' | base64 -d
v-kubernetes-demo-XXXXXXXX

$ kubectl get secret -n demo milvus-cred-rqst-xxxxxx -o jsonpath='{.data.password}' | base64 -d
xxxxxxxxxxxxxxxxxx
```

Use the issued `username` / `password` as HTTP Basic-Auth credentials (or as the `user:password` pair of a Bearer `token=` when talking to Zilliz Cloud) on the Milvus HTTP RESTful API v2. The credential is revoked when the `SecretAccessRequest` is deleted.

## Further reading

- Milvus users and roles: https://milvus.io/docs/users_and_roles.md
- OpenBao Milvus plugin: https://github.com/sigilr/openbao/pull/13
