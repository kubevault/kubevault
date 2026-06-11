---
title: Manage Oracle Database credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-oracle
    name: Overview
    parent: oracle-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Oracle Database credentials using the KubeVault operator

OpenBao's [`oracle-database-plugin`](https://github.com/sigilr/openbao/pull/6) is a **dynamic-credentials** database plugin for [Oracle Database](https://docs.oracle.com/en/database/). The plugin uses the pure-Go [`sijms/go-ora/v2`](https://github.com/sijms/go-ora) driver, so no Oracle client libraries need to be installed inside the OpenBao container. It is SQL-statement based (the same shape PostgreSQL uses): `OracleRole.spec.creationStatements` is a list of Oracle DDL/DML statements with `{{name}}` and `{{password}}` placeholders, and credentials are issued dynamically through a `SecretAccessRequest`.

> **Expiration is a no-op.** Oracle has no native `VALID UNTIL` clause on users, so the plugin's `UpdateUser` path is intentionally a no-op. Lease expiration still works as normal (the operator issues a new credential when the previous one expires), but you do **not** need to template an `{{expiration}}` placeholder into your `creationStatements` and any `renewStatements` you might be tempted to write would have no effect. To revoke a user before its lease expires, use `revocationStatements` (defaults to `DROP USER {{name}} CASCADE;` if you don't set any).

The same CRD shape is used both for the in-process `oracle-database-plugin` and for the hub-spoke `remote-oracle-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-oracle-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [OracleRole](/docs/concepts/secret-engine-crds/database-secret-engine/oraclerole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Run an Oracle database the operator can reach over the network. Refer to the Oracle Database docs at https://docs.oracle.com/en/database/ for deployment options.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Oracle

Create an `AppBinding` pointing at the Oracle database. The URL is an Oracle DSN that the pure-Go driver understands. Both connect-descriptor and easy-connect forms work, for example:

- `oracle://<user>:<pass>@<host>:1521/<service-name>` — easy-connect with service name (`ORCLCDB`, `XEPDB1`, …).
- `oracle://<user>:<pass>@<host>:1521/?SID=<sid>` — easy-connect with SID.

The referenced Secret carries the username and password the plugin uses to log in as the rotation principal (it must have privileges to `CREATE USER`, `GRANT`, and `DROP USER`).

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: oracle
  namespace: demo
spec:
  clientConfig:
    url: oracle://oracle.demo.svc:1521/ORCLPDB1
  secret:
    name: oracle-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: oracle-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: SYSTEM
  password: change-me
```

## Enable and Configure Oracle Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on a specified path and configure the secret engine with the given configuration.

A sample `SecretEngine` for Oracle:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: oracle-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  oracle:
    databaseRef:
      name: oracle
      namespace: demo
    pluginName: oracle-database-plugin   # optional; this is the default
    allowedRoles:
      - "*"
    maxOpenConnections: 4
    maxIdleConnections: 0
    maxConnectionLifetime: 0s
```

Apply it and wait for `STATUS=Success`:

```bash
$ kubectl apply -f oracle-engine.yaml
secretengine.engine.kubevault.com/oracle-engine created

$ kubectl get secretengines -n demo
NAME            STATUS    AGE
oracle-engine   Success   10s
```

Use `kubectl describe secretengine -n demo oracle-engine` to inspect error events, if any.

## Create an OracleRole

An [`OracleRole`](/docs/concepts/secret-engine-crds/database-secret-engine/oraclerole.md) describes how the plugin should mint a dynamic credential. Each entry in `creationStatements` is an Oracle DDL/DML statement, executed in sequence with the `{{name}}` and `{{password}}` placeholders substituted at credential-issue time. Note that Oracle identifiers are case-sensitive when double-quoted; the plugin issues the user name in upper-case form by default, so leaving `{{name}}` un-quoted is the simplest approach.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: OracleRole
metadata:
  name: oracle-readonly
  namespace: demo
spec:
  secretEngineRef:
    name: oracle-engine
  creationStatements:
    - |
      CREATE USER {{name}} IDENTIFIED BY "{{password}}";
      GRANT CONNECT, RESOURCE TO {{name}};
  revocationStatements:
    - DROP USER {{name}} CASCADE;
  defaultTTL: 1h
  maxTTL: 24h
```

Apply and verify:

```bash
$ kubectl apply -f oracle-role.yaml
oraclerole.engine.kubevault.com/oracle-readonly created

$ kubectl get oraclerole -n demo
NAME              STATUS    AGE
oracle-readonly   Success   12s
```

The role name in Vault follows the format `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`, so you can verify directly with the Vault CLI:

```bash
$ vault read your-database-path/roles/k8s.-.demo.oracle-readonly
Key                      Value
---                      -----
creation_statements      [CREATE USER {{name}} IDENTIFIED BY "{{password}}"; GRANT CONNECT, RESOURCE TO {{name}};]
db_name                  k8s.-.demo.oracle
default_ttl              1h
max_ttl                  24h
revocation_statements    [DROP USER {{name}} CASCADE;]
```

Deleting the `OracleRole` removes the role from Vault.

## Issue Oracle credentials

Request a dynamic credential by creating a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: oracle-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: OracleRole
    name: oracle-readonly
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Approve it through the KubeVault CLI:

```bash
$ kubectl vault approve secretaccessrequest oracle-cred-rqst -n demo
approved
```

Once approved, the operator issues the credential, stores it in a `Secret`, and binds the listed subjects via a `Role`/`RoleBinding`. The credential lives on the lease until you delete the `SecretAccessRequest` or it expires.

```bash
$ kubectl get secretaccessrequest oracle-cred-rqst -n demo -o json | jq '.status'
{
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.oracle-readonly/abc...",
    "renewable": true
  },
  "secret": {
    "name": "oracle-cred-rqst-xxxxxx"
  }
}

$ kubectl get secret -n demo oracle-cred-rqst-xxxxxx -o jsonpath='{.data.username}' | base64 -d
V_KUBERNETES_DEMO_XXXXXXXX

$ kubectl get secret -n demo oracle-cred-rqst-xxxxxx -o jsonpath='{.data.password}' | base64 -d
xxxxxxxxxxxxxxxxxx
```

Use the issued `username` / `password` to log in to the Oracle database; the credential is revoked when the `SecretAccessRequest` is deleted (the plugin runs `revocationStatements`, or falls back to a `DROP USER {{name}} CASCADE;` if you didn't set any).

## Further reading

- Oracle Database documentation: https://docs.oracle.com/en/database/
- OpenBao Oracle Database plugin: https://github.com/sigilr/openbao/pull/6
- `sijms/go-ora` pure-Go Oracle driver: https://github.com/sijms/go-ora
