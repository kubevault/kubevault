# MySQLRole CRD

Vault operator will create [database connection config](https://www.vaultproject.io/api/secret/databases/mysql-maria.html#configure-connection) and [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to `MySQLRole` CRD (CustomResourceDefinition) specification. If the user deletes the `MySQLRole` CRD, then respective role will also be deleted from Vault.

```yaml
apiVersion: authorization.kubedb.com/v1alpha1
kind: MySQLRole
metadata:
  name: <name>
  namespace: <namespace>
spec:
  ...
status:
  ...
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{spec.clusterName}.{spec.namespace}.{spec.name}`

## MySQLRole Spec

MySQLRole `spec` contains information that necessary for creating database config and role.

```yaml
apiVersion: authorization.kubedb.com/v1alpha1
kind: MySQLRole
metadata:
  name: mysql-test
  namespace: demo
spec:
  creationStatements:
    - "CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';"
    - "GRANT SELECT ON *.* TO '{{name}}'@'%';"
  defaultTTL: 300s
  maxTTL: 24h
  authManagerRef:
    namespace: demo
    name: vault-app
  databaseRef:
    name: mysql-app
```

MySQLRole Spec has following fields:

### spec.authManagerRef

`spec.authManagerRef` specifies the name and namespace of [AppBinding](https://github.com/kmodules/custom-resources/blob/10b24c8fd9028ab67a4b75cbf16d8f8e52cfe634/apis/appcatalog/v1alpha1/appbinding_types.go#L21) that contains information to communicate with Vault.

```yaml
spec:
  authManagerRef:
    namespace: demo
    name: vault-app
```

### spec.databaseRef

`spec.databaseRef` is a required field that specifies the name of [AppBinding](https://github.com/kmodules/custom-resources/blob/10b24c8fd9028ab67a4b75cbf16d8f8e52cfe634/apis/appcatalog/v1alpha1/appbinding_types.go#L21) that contains mysql database connection information. This should be in the same namespace of the `MySQLRole` CRD.

```yaml
spec:
  databaseRef:
    name: mysql-app
```

### spec.creationStatements

`spec.creationStatements` is a required field that specifies the database statements executed to create and configure a user. The `{{name}}` and `{{password}}` values will be substituted by Vault.

```yaml
spec:
  creationStatements:
    - "CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';"
    - "GRANT SELECT ON *.* TO '{{name}}'@'%';"
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

## MySQLRole Status

`status` shows the status of the MySQLRole. It is maintained by Vault operator. It contains following fields:

- `status` : Indicates whether the role successfully applied in vault or not or in progress or failed

- `conditions` : Represent observations of a MySQLRole.
