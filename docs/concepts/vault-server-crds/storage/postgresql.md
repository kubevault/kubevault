---
title: PostgreSQL | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: postgresql-storage
    name: PostgreSQL
    parent: storage-vault-server-crds
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# PostgreSQL

In PostgreSQL storage backend, data will be stored in [PostgreSQL](https://www.postgresql.org/). Vault documentation for PostgreSQL storage can be found in [here](https://www.vaultproject.io/docs/configuration/storage/postgresql.html).

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault-with-postgreSQL
  namespace: demo
spec:
  nodes: 1
  version: "0.11.1"
  backend:
    postgreSQL:
      connectionUrlSecret: "my-postgres-conn"
```

## spec.backend.postgreSQL

To use postgreSQL as backend storage in Vault specify `spec.backend.postgreSQL` in [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    postgreSQL:
      connectionUrlSecret: <secret_name>
      table: <table_name>
      maxParallel: <max_parallel>
```

`spec.backend.postgreSQL` has following fields:

#### postgreSQL.connectionUrlSecret

`postgreSQL.connectionUrlSecret` is a required field that specifies the name of the secret containing the connection string to use to authenticate and connect to PostgreSQL. The secret contains the following key:

- `connection_url`

```yaml
spec:
  backend:
    postgreSQL:
      connectionUrlSecret: "my-postgres-conn"
```

#### postgreSQL.table

`postgreSQL.table` is an optional field that specifies the name of the table in which to write Vault data. If it is not specified, then Vault will set value `vault_kv_store`. Vault will not create table, so this table must exist in database.

```yaml
spec:
  backend:
    postgreSQL:
      table: "vault_data"
```

#### postgreSQL.maxParallel

`maxParallel` is an optional field that specifies the maximum number of parallel operations to take place. This field accepts integer value. If this field is not specified, then Vault will set value `128`.

```yaml
spec:
  backend:
    postgreSQL:
      maxParallel: 124
```
