---
title: MySQL | Vault Server Storage
menu:
  docs_0.1.0:
    identifier: mysql-storage
    name: MySQL
    parent: storage-vault-server-crds
    weight: 35
menu_name: docs_0.1.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# MySQL

In MySQL storage backend, data will be stored in [MySQL](https://www.mysql.org/). Vault documentation for MySQL storage can be found in [here](https://www.vaultproject.io/docs/configuration/storage/mysql.html).

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault-with-mysql
  namespace: demo
spec:
  nodes: 1
  version: "0.11.1"
  backend:
    mySQL:
      address: "my.mysql.com:3306"
      userCredentialSecret: "mysql-cred"
```

## spec.backend.mySQL

To use MySQL as backend storage in Vault specify `spec.backend.mySQL` in [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    mySQL:
      address: <address>
      database: <database_name>
      table: <table_name>
      userCredentialSecret: <secret_name>
      tlsCASecret: <secret_name>
      maxParallel: <max_parallel>
```

`spec.backend.mySQL` has following fields:

#### mySQL.address

`mySQL.address` is a required field that specifies the address of the MySQL host.

```yaml
spec:
  backend:
    mySQL:
      address: "my.mysql.com:3306"
```

#### mySQL.userCredentialSecret

`mySQL.userCredentialSecret` is a required field that specifies the name of the secret containing MySQL username and password to connect with the database. The secret contains the following fields:

- `username`
- `password`

```yaml
spec:
  backend:
    mySQL:
      userCredentialSecret: "mysql-cred"
```

### mySQL.databse

`mySQL.database` is an optional field that specifies the name of the database. If the database does not exist, Vault will attempt to create it. If it is not specified, then Vault will set vault `vault`.

```yaml
spec:
  backend:
    mySQL:
      database: "my_vault"
```

#### mySQL.table

`mySQL.table` is an optional field that specifies the name of the table. If the table does not exist, Vault will attempt to create it. If it is not specified, then Vault will set value `vault`.

```yaml
spec:
  backend:
    mySQL:
      table: "vault_data"
```

#### mySQL.tlsCASecret

`mySQL.tlsCASecret` is an optional field that specifies the name of the secret containing the CA certificate to connect using TLS. The secret contains following fields:

- `tls_ca_file`

```yaml
spec:
  backend:
    mySQL:
      tlsCASecret: "mysql-ca"
```

#### mySQL.maxParallel

`maxParallel` is an optional field that specifies the maximum number of parallel operations to take place. This field accepts integer value. If this field is not specified, then Vault will set value `128`.

```yaml
spec:
  backend:
    mySQL:
      maxParallel: 124
```
