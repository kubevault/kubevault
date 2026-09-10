---
title: MSSQL | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: mssql-storage
    name: MSSQL
    parent: storage-vault-server-crds
    weight: 75
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# MSSQL

In the MSSQL storage backend, Vault data will be stored in [Microsoft SQL Server](https://www.microsoft.com/en-us/sql-server). This backend was removed from upstream OpenBao and is only available in this fork; Vault documentation for the MSSQL storage backend can be found [here](https://developer.hashicorp.com/vault/docs/configuration/storage/mssql).

> `spec.backend.mssql` is only available on the `kubevault.com/v1alpha2` `VaultServer` API. MSSQL has no `haEnabled` field and cannot be used to run Vault in high availability mode.

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault-with-mssql
  namespace: demo
spec:
  version: "2.0.0"
  replicas: 1
  backend:
    mssql:
      server: "mssql.demo.svc"
      credentialSecretRef:
        name: mssql-cred
```

## spec.backend.mssql

To use MSSQL as backend storage in Vault, specify `spec.backend.mssql` in the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    mssql:
      server: <host>
      port: <port>
      credentialSecretRef:
        name: <secret_name>
      database: <database_name>
      table: <table_name>
      schema: <schema_name>
      appName: <app_name>
      connectionTimeout: <seconds>
      logLevel: <log_level>
      maxParallel: <max_parallel>
```

Here, we are going to describe the various attributes of the `spec.backend.mssql` field.

### mssql.server

`mssql.server` is a required field that specifies the address of the MSSQL host.

```yaml
spec:
  backend:
    mssql:
      server: "mssql.demo.svc"
```

### mssql.port

`mssql.port` is an optional field that specifies the port of the MSSQL host. Defaults to the driver's standard port (`1433`) when unset.

```yaml
spec:
  backend:
    mssql:
      port: "1433"
```

### mssql.credentialSecretRef

`mssql.credentialSecretRef` is an optional field that specifies the name of the `Secret`, in the same namespace as the `VaultServer`, containing the MSSQL username and password to connect with. The secret contains the following keys:

- `username`
- `password`

```yaml
spec:
  backend:
    mssql:
      credentialSecretRef:
        name: mssql-cred
```

### mssql.database

`mssql.database` is an optional field that specifies the name of the database to use. Vault will attempt to create it if it does not already exist. Defaults to `openbao`.

```yaml
spec:
  backend:
    mssql:
      database: "vault"
```

### mssql.table

`mssql.table` is an optional field that specifies the name of the table in which to write Vault data. Vault will attempt to create it if missing. Defaults to `openbao`.

```yaml
spec:
  backend:
    mssql:
      table: "vault_kv_store"
```

### mssql.schema

`mssql.schema` is an optional field that specifies the schema, within the database, that the table lives in. Vault will attempt to create it if missing (requires permission to run `CREATE SCHEMA`). Defaults to `dbo`.

```yaml
spec:
  backend:
    mssql:
      schema: "dbo"
```

### mssql.appName

`mssql.appName` is an optional field that specifies the application name to report to the server. Defaults to `openbao`.

```yaml
spec:
  backend:
    mssql:
      appName: "vault"
```

### mssql.connectionTimeout

`mssql.connectionTimeout` is an optional field that specifies the connection timeout, in seconds. Defaults to `30`.

```yaml
spec:
  backend:
    mssql:
      connectionTimeout: "30"
```

### mssql.logLevel

`mssql.logLevel` is an optional field that specifies the driver's internal log level bitmask.

```yaml
spec:
  backend:
    mssql:
      logLevel: "1"
```

### mssql.maxParallel

`mssql.maxParallel` is an optional field that specifies the maximum number of concurrent requests to MSSQL.

```yaml
spec:
  backend:
    mssql:
      maxParallel: 128
```
