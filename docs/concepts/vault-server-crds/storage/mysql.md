---
title: MySQL | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: mysql-storage
    name: MySQL
    parent: storage-vault-server-crds
    weight: 35
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# MySQL

In MySQL storage backend, Vault data will be stored in [MySQL](https://www.mysql.com/). Vault documentation for MySQL storage can be found in [here](https://www.vaultproject.io/docs/configuration/storage/mysql.html).

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault-with-mysql
  namespace: demo
spec:
  replicas: 1
  version: "1.2.0"
  backend:
    mysql:
      address: "my.mysql.com:3306"
      userCredentialSecret: "mysql-cred"
```

## spec.backend.mysql

To use MySQL as backend storage in Vault, specify `spec.backend.mysql` in [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    mysql:
      address: <address>
      database: <database_name>
      table: <table_name>
      userCredentialSecret: <secret_name>
      tlsCASecret: <secret_name>
      maxParallel: <max_parallel>
```

Here, we are going to describe the various attributes of the `spec.backend.mysql` field.

### mysql.address

`mysql.address` is a required field that specifies the address of the MySQL host.

```yaml
spec:
  backend:
    mysql:
      address: "my.mysql.com:3306"
```

### mysql.userCredentialSecret

`mysql.userCredentialSecret` is a required field that specifies the name of the secret containing MySQL username and password to connect with the database. The secret contains the following fields:

- `username`
- `password`

```yaml
spec:
  backend:
    mysql:
      userCredentialSecret: "mysql-cred"
```

### mysql.database

`mysql.database` is an optional field that specifies the name of the database. If the database does not exist, Vault will attempt to create it. If it is not specified, then Vault will set vault `vault`.

```yaml
spec:
  backend:
    mysql:
      database: "my_vault"
```

### mysql.table

`mysql.table` is an optional field that specifies the name of the table. If the table does not exist, Vault will attempt to create it. If it is not specified, then Vault will set value to `vault`.

```yaml
spec:
  backend:
    mysql:
      table: "vault_data"
```

### mysql.tlsCASecret

`mysql.tlsCASecret` is an optional field that specifies the name of the secret containing the CA certificate to connect using TLS. The secret contains the following fields:

- `tls_ca_file`

```yaml
spec:
  backend:
    mysql:
      tlsCASecret: "mysql-ca"
```

### mysql.maxParallel

`maxParallel` is an optional field that specifies the maximum number of parallel operations to take place. This field accepts integer value. If this field is not specified, then Vault will set value to `128`.

```yaml
spec:
  backend:
    mysql:
      maxParallel: 124
```
