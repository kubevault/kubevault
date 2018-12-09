
## Setup

1. Enable the Datbase secret engine:

```bash
$ vault secrets enable database
Success! Enabled the database secrets engine at: database/
```

2. Configure Vault with the proper plugin and connection information:

```bash
$ vault write database/config/my-postgresql-database \
  plugin_name=postgresql-database-plugin \
  allowed_roles="my-role" \
  connection_url="postgresql://{{username}}:{{password}}@159.89.41.120:30595/postgresdb?sslmode=disable" \
  username="postgresadmin" \
  password="admin123"
```

3. Configure a role that maps a name in Vault to an SQL statement to execute to create the database credential:

```bash
$ vault write database/roles/my-role \
  db_name=my-postgresql-database \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
  GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
  default_ttl="1h" \
  max_ttl="24h"
Success! Data written to: database/roles/my-role

```

4. Create Storageclass with followings:

```bash
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: vault-pg-storage
  annotations:
    storageclass.kubernetes.io/is-default-class: "false"
provisioner: com.vault.csi.vaultdbs
parameters:
  ref: default/vaultapp
  engine: DATABASE
  role: my-role
  path: database
```