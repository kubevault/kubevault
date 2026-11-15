---
title: Microsoft SQL Server | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: mssql-storage
    name: Microsoft SQL Server
    parent: storage-vault-server-crds
    weight: 150
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Microsoft SQL Server

In the MSSQL storage backend, Vault data will be stored in [Microsoft SQL Server](https://www.microsoft.com/en-us/sql-server). Vault documentation for the MSSQL storage backend can be found [here](https://developer.hashicorp.com/vault/docs/configuration/storage/mssql).

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
      databaseRef:
        name: <appbinding_name>
        namespace: <appbinding_namespace>
      encrypt: <true|false|strict|disable>
      tlsSecretRef:
        name: <secret_name>
      trustServerCertificate: <true|false>
      tlsServerName: <server_name>
      tlsMinVersion: <tls10|tls11|tls12|tls13>
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

`mssql.server` specifies the address of the MSSQL host. Either `server` or `databaseRef` must be set.

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

### mssql.databaseRef

`mssql.databaseRef` is an optional field that points to the `AppBinding` of a KubeDB managed `MSSQLServer`. When it is set, the operator takes the server, port and credentials from the `AppBinding`, so `server`, `port` and `credentialSecretRef` are not needed. If the `AppBinding` has a CA bundle, the operator also uses it to verify the server certificate (see [mssql.tlsSecretRef](#mssqltlssecretref)).

```yaml
spec:
  backend:
    mssql:
      databaseRef:
        name: mssql
        namespace: demo
```

### mssql.encrypt

`mssql.encrypt` is an optional field that specifies how the connection to SQL Server is encrypted. It is one of:

- `"true"`: encrypt the whole connection, and fail if the server doesn't support encryption.
- `"false"`: encrypt only the login packet, unless the server forces encryption.
- `strict`: use TDS 8.0, where TLS is set up before any TDS traffic. Requires SQL Server 2022 or later.
- `disable`: don't encrypt anything.

Unless it is `disable`, the server certificate is verified, unless `trustServerCertificate` is `true`. When `encrypt` is unset, it is `"true"` if a CA is available (from `tlsSecretRef` or the `databaseRef` `AppBinding`) or any other TLS field is set. Otherwise it is left to the driver, which encrypts only the login packet and does **not** verify the server certificate.

```yaml
spec:
  backend:
    mssql:
      encrypt: "true"
```

### mssql.tlsSecretRef

`mssql.tlsSecretRef` is an optional field that specifies the name of the `Secret`, in the same namespace as the `VaultServer`, containing the CA certificate used to verify the SQL Server certificate. The secret contains the following key:

- `ca.crt`

The operator mounts it into the Vault pods and sets `tls_ca_file` in the Vault configuration. If `databaseRef` is set and `tlsSecretRef` is not, the operator copies the CA bundle of the `AppBinding` into a secret named `<vaultserver-name>-mssql-tls` and uses that. It can't be used with `encrypt: disable`. SQL Server has no client certificate authentication, so only a CA is used.

```yaml
spec:
  backend:
    mssql:
      tlsSecretRef:
        name: mssql-ca
```

### mssql.trustServerCertificate

`mssql.trustServerCertificate` is an optional field that specifies whether to skip verifying the server certificate. Not recommended in production. Defaults to `false`.

```yaml
spec:
  backend:
    mssql:
      trustServerCertificate: true
```

### mssql.tlsServerName

`mssql.tlsServerName` is an optional field that specifies the name the server certificate must be valid for. Defaults to the server address.

```yaml
spec:
  backend:
    mssql:
      tlsServerName: "mssql.demo.svc"
```

### mssql.tlsMinVersion

`mssql.tlsMinVersion` is an optional field that specifies the minimum TLS version to use. It is one of `tls10`, `tls11`, `tls12` and `tls13`.

```yaml
spec:
  backend:
    mssql:
      tlsMinVersion: tls12
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
