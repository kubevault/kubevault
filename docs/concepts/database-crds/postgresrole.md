---
title: PostgresRole | Vault Secret Engine
menu:
  docs_0.1.0:
    identifier: postgresrole-database-crds
    name: PostgresRole
    parent: database-crds-concepts
    weight: 20
menu_name: docs_0.1.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# PostgresRole CRD

Vault operator will create [database connection config](https://www.vaultproject.io/api/secret/databases/postgresql.html#configure-connection) and [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to `PostgresRole` CRD (CustomResourceDefinition) specification. If the user deletes the `PostgresRole` CRD, then respective role will also be deleted from Vault.

```yaml
apiVersion: authorization.kubedb.com/v1alpha1
kind: PostgresRole
metadata:
  name: <name>
  namespace: <namespace>
spec:
  ...
status:
  ...
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{spec.clusterName}.{spec.namespace}.{spec.name}`

## PostgresRole Spec

PostgresRole `spec` contains information that necessary for creating database config and role.

```yaml
apiVersion: authorization.kubedb.com/v1alpha1
kind: PostgresRole
metadata:
  name: postgres-test
  namespace: demo
spec:
  creationStatements:
    - "CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';"
    - "GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";"
  defaultTTL: 300s
  maxTTL: 24h
  authManagerRef:
    namespace: demo
    name: vault-app
  databaseRef:
    name: postgres-app
```

PostgresRole Spec has following fields:

### spec.authManagerRef

`spec.authManagerRef` specifies the name and namespace of [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains information to communicate with Vault.

```yaml
spec:
  authManagerRef:
    namespace: demo
    name: vault-app
```

### spec.databaseRef

`spec.databaseRef` is a required field that specifies the name of [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains Postgres database connection information. This should be in the same namespace of the `PostgresRole` CRD.

```yaml
spec:
  databaseRef:
    name: postgres-app
```

### spec.creationStatements

`spec.creationStatements` is a required field that specifies the database statements executed to create and configure a user. The `{{name}}`, `{{password}}` and `{{expiration}}` values will be substituted by Vault.

```yaml
spec:
  creationStatements:
    - "CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';"
    - "GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";"
```

### spec.defaultTTL

`spec.defaultTTL` is an optional field that specifies the TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to system/engine default TTL time.

```yaml
spec:
  defaultTTL: "1h"
```

### spec.maxTTL

`spec.maxTTL` is an optional field that specifies the maximum TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to system/engine default TTL time.

```yaml
spec:
  maxTTL: "1h"
```

### spec.revocationStatements

`spec.revocationStatements` is an optional field that specifies the database statements to be executed to revoke a user. The `{{name}}` value will be substituted. If not provided defaults to a generic drop user statement.

### spec.rollbackStatements

`spec.rollbackStatements` is an optional field that specifies the database statements to be executed rollback a create operation in the event of an error. Not every plugin type will support this functionality.

### spec.renewStatements

`spec.renewStatements` is an optional field that specifies the database statements to be executed to renew a user. Not every plugin type will support this functionality.

## PostgresRole Status

`status` shows the status of the PostgresRole. It is maintained by Vault operator. It contains following fields:

- `status` : Indicates whether the role successfully applied in vault or not or in progress or failed

- `conditions` : Represent observations of a PostgresRole.
