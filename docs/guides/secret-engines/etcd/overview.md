---
title: Manage etcd credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-etcd
    name: Overview
    parent: etcd-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage etcd credentials using the KubeVault operator

OpenBao's [`etcd-database-plugin`](https://github.com/sigilr/openbao/pull/50) is a **dynamic-credentials** database plugin for [etcd](https://etcd.io/), built on etcd's built-in [v3 Auth API](https://etcd.io/docs/latest/op-guide/authentication/) via the official `go.etcd.io/etcd/client/v3` client. Each issued credential becomes a native etcd user (created via `UserAdd`) and is granted one or more pre-existing roles via `UserGrantRole`. The plugin **does not create roles**; it only manages users and their role grants, so every role referenced from `creationStatements` must already exist on the cluster (e.g. via `etcdctl role add`).

The same CRD shape is used both for the in-process `etcd-database-plugin` and for the hub-spoke `remote-etcd-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-etcd-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [EtcdRole](/docs/concepts/secret-engine-crds/database-secret-engine/etcd.md)
- [SecretAccessRequest](/docs/concepts/secret-engine-crds/secret-access-request.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Run an etcd cluster with [authentication enabled](https://etcd.io/docs/latest/op-guide/authentication/) (`etcdctl auth enable`). The root user configured on the `AppBinding` must have permission to add users and grant roles.
- Pre-create the etcd roles you want to bind credentials to (e.g. `reader`, `writer` via `etcdctl role add`). The plugin only grants — it does not create roles.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for etcd

Create an `AppBinding` pointing at the etcd client endpoint. The referenced Secret carries the root username and password used to authenticate against etcd's v3 Auth API when the plugin calls `UserAdd`/`UserGrantRole`/`UserChangePassword`/`UserDelete`.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: etcd
  namespace: demo
spec:
  clientConfig:
    url: http://etcd.demo.svc:2379
  secret:
    kind: Secret
    name: etcd-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: etcd-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: root
  password: change-me
```

> **TLS Configuration via AppBinding:**
> If the etcd cluster serves client traffic over TLS, configure it in the etcd `AppBinding` via `spec.clientConfig.caBundle` or `spec.tlsSecret` (and optionally `spec.clientConfig.insecureSkipTLSVerify` for self-signed certificates). The KubeVault operator automatically detects TLS from the referenced `AppBinding` and configures TLS (`use_tls`, CA certificate, and verification flags) for the etcd plugin. Do not specify `useTLS` or `insecure` on the `SecretEngine` object — etcd has no such fields.

## Enable and Configure etcd Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on a specified path and configure the secret engine with the given configuration.

A sample `SecretEngine` for etcd:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: etcd-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  etcd:
    databaseRef:
      name: etcd
      namespace: demo
    pluginName: etcd-database-plugin   # optional; this is the default
    allowedRoles:
      - "*"
```

Apply it and wait for `STATUS=Success`:

```bash
$ kubectl apply -f etcd-engine.yaml
secretengine.engine.kubevault.com/etcd-engine created

$ kubectl get secretengines -n demo
NAME          STATUS    AGE
etcd-engine   Success   10s
```

Use `kubectl describe secretengine -n demo etcd-engine` to inspect error events, if any.

Behind the scenes the KubeVault operator writes:

```
bao write database/config/k8s.<cluster>.demo.etcd \
    plugin_name=etcd-database-plugin \
    endpoints=http://etcd.demo.svc:2379 \
    username=root \
    password=change-me \
    allowed_roles="*"
```

## Create an EtcdRole

An [`EtcdRole`](/docs/concepts/secret-engine-crds/database-secret-engine/etcd.md) describes how the plugin should mint a dynamic credential. `creationStatements` is a single-element string slice holding a JSON role document of the form `'{"roles":["reader","writer"]}'`. The listed roles **must already exist** on the etcd cluster — the plugin only grants via `UserGrantRole`, it does not create roles.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: EtcdRole
metadata:
  name: etcd-reader
  namespace: demo
spec:
  secretEngineRef:
    name: etcd-engine
  creationStatements:
    - '{"roles":["reader"]}'
  defaultTTL: 1h
  maxTTL: 24h
```

Apply and verify:

```bash
$ kubectl apply -f etcd-role.yaml
etcdrole.engine.kubevault.com/etcd-reader created

$ kubectl get etcdrole -n demo
NAME          STATUS    AGE
etcd-reader   Success   12s
```

The role name in Vault follows the format `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`, so you can verify directly with the Vault CLI:

```bash
$ vault read your-database-path/roles/k8s.-.demo.etcd-reader
Key                      Value
---                      -----
creation_statements      [{"roles":["reader"]}]
db_name                  k8s.-.demo.etcd
default_ttl              1h
max_ttl                  24h
```

Deleting the `EtcdRole` removes the role from Vault.

## Issue etcd credentials

Request a dynamic credential by creating a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: etcd-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: EtcdRole
    name: etcd-reader
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Approve it through the KubeVault CLI:

```bash
$ kubectl vault approve secretaccessrequest etcd-cred-rqst -n demo
approved
```

Once approved, the operator issues the credential, stores it in a `Secret`, and binds the listed subjects via a `Role`/`RoleBinding`. The plugin creates a new etcd user via `UserAdd` and grants it the `reader` role via `UserGrantRole`. The credential lives on the lease until you delete the `SecretAccessRequest` or it expires; on lease revocation the plugin deletes the etcd user via `UserDelete` (treating a not-found user as success, so revocation is idempotent).

```bash
$ kubectl get secretaccessrequest etcd-cred-rqst -n demo -o json | jq '.status'
{
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.etcd-reader/abc...",
    "renewable": true
  },
  "secret": {
    "name": "etcd-cred-rqst-xxxxxx"
  }
}

$ kubectl get secret -n demo etcd-cred-rqst-xxxxxx -o jsonpath='{.data.username}' | base64 -d
v-kubernetes-demo-XXXXXXXX

$ kubectl get secret -n demo etcd-cred-rqst-xxxxxx -o jsonpath='{.data.password}' | base64 -d
xxxxxxxxxxxxxxxxxx
```

Use the issued `username` / `password` as your etcd client's auth credentials (e.g. `etcdctl --user username:password`); the credential is revoked when the `SecretAccessRequest` is deleted.

## Further reading

- etcd authentication: https://etcd.io/docs/latest/op-guide/authentication/
- OpenBao etcd plugin: https://github.com/sigilr/openbao/pull/50
