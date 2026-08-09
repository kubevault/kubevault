---
title: CockroachDB | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: cockroachdb-storage
    name: CockroachDB
    parent: storage-vault-server-crds
    weight: 55
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# CockroachDB

In the CockroachDB storage backend, Vault data will be stored in [CockroachDB](https://www.cockroachlabs.com/). CockroachDB speaks the PostgreSQL wire protocol, so this backend is configured the same way as `postgresql`. This backend was removed from upstream OpenBao and is only available in this fork; Vault documentation for the CockroachDB storage backend can be found [here](https://developer.hashicorp.com/vault/docs/configuration/storage/cockroachdb).

> `spec.backend.cockroachdb` is only available on the `kubevault.com/v1alpha2` `VaultServer` API.

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault-with-cockroachdb
  namespace: demo
spec:
  version: "2.0.0"
  replicas: 1
  backend:
    cockroachdb:
      address: "cockroachdb-public.demo.svc:26257"
      credentialSecretRef:
        name: cockroachdb-cred
      sslMode: verify-full
```

## spec.backend.cockroachdb

To use CockroachDB as backend storage in Vault, specify `spec.backend.cockroachdb` in the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    cockroachdb:
      address: <host:port>
      credentialSecretRef:
        name: <secret_name>
      databaseRef:
        name: <appbinding_name>
        namespace: <appbinding_namespace>
      sslMode: <disable|require|verify-ca|verify-full>
      table: <table_name>
      maxParallel: <max_parallel>
      transactionMaxParallel: <transaction_max_parallel>
      skipCreateTable: <true/false>
      haEnabled: <true/false>
      haTable: <ha_table_name>
```

Here, we are going to describe the various attributes of the `spec.backend.cockroachdb` field.

### cockroachdb.address

`cockroachdb.address` specifies the address (`host:port`) of the CockroachDB cluster to connect to. This is required unless `databaseRef` is set, in which case the address is resolved from the referenced `AppBinding` instead.

```yaml
spec:
  backend:
    cockroachdb:
      address: "cockroachdb-public.demo.svc:26257"
```

### cockroachdb.credentialSecretRef

`cockroachdb.credentialSecretRef` is an optional field that specifies the name of the `Secret`, in the same namespace as the `VaultServer`, containing the username and password (and, optionally, a full `connection_url`) used to connect to CockroachDB. The secret contains the following keys:

- `username`
- `password`
- `connection_url` (optional; when present, it is used verbatim instead of being built from `address`/`username`/`password`/`sslMode`)

```yaml
spec:
  backend:
    cockroachdb:
      credentialSecretRef:
        name: cockroachdb-cred
```

### cockroachdb.databaseRef

`cockroachdb.databaseRef` is an optional field that references a KubeDB-managed `AppBinding` for the CockroachDB cluster. When set, the host (and, if the `AppBinding` carries one, the credential secret) is resolved from it instead of from `address`/`credentialSecretRef`.

```yaml
spec:
  backend:
    cockroachdb:
      databaseRef:
        name: cockroachdb-appbinding
        namespace: demo
```

### cockroachdb.sslMode

`cockroachdb.sslMode` is an optional field that specifies the SSL mode to use when connecting to CockroachDB. Must be one of `disable`, `require`, `verify-ca`, or `verify-full`. Defaults to `verify-full`.

```yaml
spec:
  backend:
    cockroachdb:
      sslMode: verify-full
```

### cockroachdb.table

`cockroachdb.table` is an optional field that specifies the name of the table in which to write Vault data. Defaults to `openbao_kv_store`.

```yaml
spec:
  backend:
    cockroachdb:
      table: "vault_kv_store"
```

### cockroachdb.maxParallel

`cockroachdb.maxParallel` is an optional field that specifies the maximum number of parallel operations to take place.

```yaml
spec:
  backend:
    cockroachdb:
      maxParallel: 128
```

### cockroachdb.transactionMaxParallel

`cockroachdb.transactionMaxParallel` is an optional field that specifies the maximum number of concurrent operations to take place within a single CockroachDB transaction.

```yaml
spec:
  backend:
    cockroachdb:
      transactionMaxParallel: 32
```

### cockroachdb.skipCreateTable

`cockroachdb.skipCreateTable` is an optional field that, when `true`, skips the CREATE TABLE (and, if `haEnabled` is set, CREATE TABLE for the HA table) statement Vault would otherwise run against the database. This field accepts a boolean value. The default is `false`.

```yaml
spec:
  backend:
    cockroachdb:
      skipCreateTable: true
```

### cockroachdb.haEnabled

`cockroachdb.haEnabled` is an optional field that specifies whether this backend should be used to run Vault in high availability mode. This field accepts a boolean value (as a string). The default value is `false`.

```yaml
spec:
  backend:
    cockroachdb:
      haEnabled: "true"
```

### cockroachdb.haTable

`cockroachdb.haTable` is an optional field that specifies the name of the table used for HA lock information, when `haEnabled` is `true`. Defaults to `openbao_ha_locks`.

```yaml
spec:
  backend:
    cockroachdb:
      haTable: "vault_ha_locks"
```
