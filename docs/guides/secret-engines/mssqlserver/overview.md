---
title: Manage Microsoft SQL Server credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-mssqlserver
    name: Overview
    parent: mssqlserver-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Microsoft SQL Server credentials using the KubeVault operator

OpenBao's [`mssql-database-plugin`](https://github.com/sigilr/openbao/pull/5) is a **dynamic-credentials** database plugin for [Microsoft SQL Server](https://learn.microsoft.com/en-us/sql/relational-databases/security/). The plugin was ported from pre-BUSL HashiCorp Vault using [`microsoft/go-mssqldb`](https://github.com/microsoft/go-mssqldb) and is T-SQL statement based (the same shape PostgreSQL uses): `MSSQLServerRole.spec.creationStatements` is a list of T-SQL statements with `{{name}}` and `{{password}}` placeholders, and credentials are issued dynamically through a `SecretAccessRequest`.

The same CRD shape is used both for the in-process `mssql-database-plugin` and for the hub-spoke `remote-mssql-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-mssql-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [MSSQLServerRole](/docs/concepts/secret-engine-crds/database-secret-engine/mssqlserverrole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Run a Microsoft SQL Server instance the operator can reach over the network. Refer to the SQL Server security docs at https://learn.microsoft.com/en-us/sql/relational-databases/security/ for deployment options.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Microsoft SQL Server

Create an `AppBinding` pointing at the SQL Server database. The URL is a SQL Server DSN (`sqlserver://<user>:<pass>@<host>:1433`); the referenced Secret carries the username and password the plugin uses to log in as the rotation principal.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: mssqlserver
  namespace: demo
spec:
  clientConfig:
    url: sqlserver://mssql.demo.svc:1433
  secret:
    name: mssqlserver-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: mssqlserver-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: sa
  password: change-me
```

## Enable and Configure Microsoft SQL Server Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on a specified path and configure the secret engine with the given configuration.

A sample `SecretEngine` for Microsoft SQL Server:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: mssqlserver-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  mssqlserver:
    databaseRef:
      name: mssqlserver
      namespace: demo
    pluginName: mssql-database-plugin   # optional; this is the default
    allowedRoles:
      - "*"
    maxOpenConnections: 4
    maxIdleConnections: 0
    maxConnectionLifetime: 0s
    containedDB: false
```

Set `containedDB: true` to switch the plugin into **contained-database authentication mode** (the plugin runs `CREATE USER ... WITH PASSWORD` against the user database directly). The default `false` keeps the standard behavior of creating logins on `master` and mapping users into the target database. See the SQL Server [contained databases](https://learn.microsoft.com/en-us/sql/relational-databases/databases/contained-databases) docs for the trade-offs.

Apply it and wait for `STATUS=Success`:

```bash
$ kubectl apply -f mssqlserver-engine.yaml
secretengine.engine.kubevault.com/mssqlserver-engine created

$ kubectl get secretengines -n demo
NAME                 STATUS    AGE
mssqlserver-engine   Success   10s
```

Use `kubectl describe secretengine -n demo mssqlserver-engine` to inspect error events, if any.

## Create a MSSQLServerRole

A [`MSSQLServerRole`](/docs/concepts/secret-engine-crds/database-secret-engine/mssqlserverrole.md) describes how the plugin should mint a dynamic credential. Each entry in `creationStatements` is a T-SQL statement, executed in sequence with the `{{name}}` and `{{password}}` placeholders substituted at credential-issue time.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: MSSQLServerRole
metadata:
  name: mssqlserver-readonly
  namespace: demo
spec:
  secretEngineRef:
    name: mssqlserver-engine
  creationStatements:
    - |
      CREATE LOGIN [{{name}}] WITH PASSWORD = '{{password}}';
      CREATE USER [{{name}}] FOR LOGIN [{{name}}];
      GRANT SELECT ON SCHEMA::dbo TO [{{name}}];
  defaultTTL: 1h
  maxTTL: 24h
```

Apply and verify:

```bash
$ kubectl apply -f mssqlserver-role.yaml
mssqlserverrole.engine.kubevault.com/mssqlserver-readonly created

$ kubectl get mssqlserverrole -n demo
NAME                   STATUS    AGE
mssqlserver-readonly   Success   12s
```

The role name in Vault follows the format `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`, so you can verify directly with the Vault CLI:

```bash
$ vault read your-database-path/roles/k8s.-.demo.mssqlserver-readonly
Key                      Value
---                      -----
creation_statements      [CREATE LOGIN [{{name}}] WITH PASSWORD = '{{password}}'; CREATE USER [{{name}}] FOR LOGIN [{{name}}]; GRANT SELECT ON SCHEMA::dbo TO [{{name}}];]
db_name                  k8s.-.demo.mssqlserver
default_ttl              1h
max_ttl                  24h
```

Deleting the `MSSQLServerRole` removes the role from Vault.

## Issue Microsoft SQL Server credentials

Request a dynamic credential by creating a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: mssqlserver-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: MSSQLServerRole
    name: mssqlserver-readonly
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Approve it through the KubeVault CLI:

```bash
$ kubectl vault approve secretaccessrequest mssqlserver-cred-rqst -n demo
approved
```

Once approved, the operator issues the credential, stores it in a `Secret`, and binds the listed subjects via a `Role`/`RoleBinding`. The credential lives on the lease until you delete the `SecretAccessRequest` or it expires.

```bash
$ kubectl get secretaccessrequest mssqlserver-cred-rqst -n demo -o json | jq '.status'
{
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.mssqlserver-readonly/abc...",
    "renewable": true
  },
  "secret": {
    "name": "mssqlserver-cred-rqst-xxxxxx"
  }
}

$ kubectl get secret -n demo mssqlserver-cred-rqst-xxxxxx -o jsonpath='{.data.username}' | base64 -d
v-kubernetes-demo-XXXXXXXX

$ kubectl get secret -n demo mssqlserver-cred-rqst-xxxxxx -o jsonpath='{.data.password}' | base64 -d
xxxxxxxxxxxxxxxxxx
```

Use the issued `username` / `password` to log in to the SQL Server database; the credential is revoked when the `SecretAccessRequest` is deleted (the plugin runs `revocationStatements`, or falls back to a sensible `DROP LOGIN` / `DROP USER` if you didn't set any).

## Further reading

- Microsoft SQL Server security docs: https://learn.microsoft.com/en-us/sql/relational-databases/security/
- OpenBao Microsoft SQL Server plugin: https://github.com/sigilr/openbao/pull/5
