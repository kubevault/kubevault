---
title: Manage Apache Ignite credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-ignite
    name: Overview
    parent: ignite-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Apache Ignite credentials using the KubeVault operator

OpenBao's [`ignite-database-plugin`](https://github.com/sigilr/openbao/pull/14) is a **dynamic-credentials** database plugin for [Apache Ignite](https://ignite.apache.org/docs/latest/security/authentication). It drives Ignite's REST API (`cmd=qryfldexe`) to execute `CREATE USER` / `ALTER USER` / `DROP USER` SQL DDL statements, so `IgniteRole.spec.creationStatements` is a list of Ignite SQL DDL statements with `{{name}}` and `{{password}}` placeholders, and credentials are issued dynamically through a `SecretAccessRequest`.

The same CRD shape is used both for the in-process `ignite-database-plugin` and for the hub-spoke `remote-ignite-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-ignite-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [IgniteRole](/docs/concepts/secret-engine-crds/database-secret-engine/igniterole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Run an Apache Ignite cluster the operator can reach over the network. The plugin requires **Ignite 2.5+** with **persistence enabled** and `authenticationEnabled=true` in the cluster configuration (see https://ignite.apache.org/docs/latest/security/authentication). Without persistence, Ignite refuses to enable authentication.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Apache Ignite

Create an `AppBinding` pointing at the Ignite REST endpoint. The URL is the HTTP(S) base of the Ignite REST API (e.g. `http://ignite.demo.svc:8080`); the referenced Secret carries the username and password the plugin uses to authenticate Basic Auth against the REST endpoint as the rotation principal.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: ignite
  namespace: demo
spec:
  clientConfig:
    url: http://ignite.demo.svc:8080
  secret:
    name: ignite-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: ignite-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: ignite
  password: ignite
```

## Enable and Configure Apache Ignite Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on a specified path and configure the secret engine with the given configuration.

A sample `SecretEngine` for Apache Ignite:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: ignite-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  ignite:
    databaseRef:
      name: ignite
      namespace: demo
    pluginName: ignite-database-plugin   # optional; this is the default
    allowedRoles:
      - "*"
    insecure: false
```

Apply it and wait for `STATUS=Success`:

```bash
$ kubectl apply -f ignite-engine.yaml
secretengine.engine.kubevault.com/ignite-engine created

$ kubectl get secretengines -n demo
NAME            STATUS    AGE
ignite-engine   Success   10s
```

Use `kubectl describe secretengine -n demo ignite-engine` to inspect error events, if any.

Unlike SQL-driver based engines (Postgres, MySQL, HanaDB), the Ignite configuration intentionally omits `maxOpenConnections` / `maxIdleConnections` / `maxConnectionLifetime` because the REST API does not expose connection-pool tuning knobs.

## Create an IgniteRole

An [`IgniteRole`](/docs/concepts/secret-engine-crds/database-secret-engine/igniterole.md) describes how the plugin should mint a dynamic credential. Each entry in `creationStatements` is an Ignite SQL DDL statement, executed in sequence with the `{{name}}` and `{{password}}` placeholders substituted at credential-issue time.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: IgniteRole
metadata:
  name: ignite-app
  namespace: demo
spec:
  secretEngineRef:
    name: ignite-engine
  creationStatements:
    - CREATE USER "{{name}}" WITH PASSWORD '{{password}}';
  revocationStatements:
    - DROP USER "{{name}}";
  defaultTTL: 1h
  maxTTL: 24h
```

Apply and verify:

```bash
$ kubectl apply -f ignite-role.yaml
igniterole.engine.kubevault.com/ignite-app created

$ kubectl get igniterole -n demo
NAME         STATUS    AGE
ignite-app   Success   12s
```

The role name in Vault follows the format `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`, so you can verify directly with the Vault CLI:

```bash
$ vault read your-database-path/roles/k8s.-.demo.ignite-app
Key                      Value
---                      -----
creation_statements      [CREATE USER "{{name}}" WITH PASSWORD '{{password}}';]
db_name                  k8s.-.demo.ignite
default_ttl              1h
max_ttl                  24h
revocation_statements    [DROP USER "{{name}}";]
```

Deleting the `IgniteRole` removes the role from Vault.

## Issue Apache Ignite credentials

Request a dynamic credential by creating a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: ignite-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: IgniteRole
    name: ignite-app
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Approve it through the KubeVault CLI:

```bash
$ kubectl vault approve secretaccessrequest ignite-cred-rqst -n demo
approved
```

Once approved, the operator issues the credential, stores it in a `Secret`, and binds the listed subjects via a `Role`/`RoleBinding`. The credential lives on the lease until you delete the `SecretAccessRequest` or it expires.

```bash
$ kubectl get secretaccessrequest ignite-cred-rqst -n demo -o json | jq '.status'
{
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.ignite-app/abc...",
    "renewable": true
  },
  "secret": {
    "name": "ignite-cred-rqst-xxxxxx"
  }
}

$ kubectl get secret -n demo ignite-cred-rqst-xxxxxx -o jsonpath='{.data.username}' | base64 -d
v-kubernetes-demo-XXXXXXXX

$ kubectl get secret -n demo ignite-cred-rqst-xxxxxx -o jsonpath='{.data.password}' | base64 -d
xxxxxxxxxxxxxxxxxx
```

Use the issued `username` / `password` to authenticate to the Ignite cluster (REST API, thin client, or JDBC); the credential is revoked when the `SecretAccessRequest` is deleted (the plugin runs `revocationStatements`, or falls back to a sensible `DROP USER` if you didn't set any).

## Further reading

- Apache Ignite authentication docs: https://ignite.apache.org/docs/latest/security/authentication
- OpenBao Apache Ignite plugin: https://github.com/sigilr/openbao/pull/14
