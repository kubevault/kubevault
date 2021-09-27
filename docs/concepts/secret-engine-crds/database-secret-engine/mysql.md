---
title: MySQLRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: mysql-database-crds
    name: MySQLRole
    parent: database-crds-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# MySQLRole

## What is MySQLRole

A `MySQLRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When a `MySQLRole` is created, the KubeVault operator creates a
[role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to the specification.
If the user deletes the `MySQLRole` CRD, then the respective role will also be deleted from Vault.

![MySQLRole CRD](/docs/images/concepts/mysql_role.svg)

## MySQLRole CRD Specification

Like any official Kubernetes resource, a `MySQLRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `MySQLRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: MySQLRole
metadata:
  name: mysql-role
  namespace: demo
spec:
  secretEngineRef:
    name: sql-secret-engine
  creationStatements:
    - "CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';"
    - "GRANT SELECT ON *.* TO '{{name}}'@'%';"
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `MySQLRole` crd.

### MySQLRole Spec

MySQLRole `spec` contains information that necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <secret-engine-name>
  path: <database-secret-engine-path>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
  creationStatements:
    - "statement-0"
    - "statement-1"
  revocationStatements:
    - "statement-0"
```

MySQLRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) reference which is used to connect with a Vault server. AppBinding must be on the same namespace with the MySQLRole object.

```yaml
spec:
  secretEngineRef:
    name: sql-secret-engine
```

#### spec.path

`spec.path` is an `optional` field that specifies the path where the secret engine is enabled. The default value is `database`.

```yaml
spec:
  path: my-mysql-path
```

#### spec.creationStatements

`spec.creationStatements` is a `required` field that specifies a list of database statements executed to create and configure a user.
The `{{name}}` and `{{password}}` values will be substituted by Vault.

```yaml
spec:
  creationStatements:
    - "CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';"
    - "GRANT SELECT ON *.* TO '{{name}}'@'%';"
```

#### spec.defaultTTL

`spec.defaultTTL` is an `optional` field that specifies the TTL for the leases associated with this role.
Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to system/engine default TTL time.

```yaml
spec:
  defaultTTL: "1h"
```

#### spec.maxTTL

`spec.maxTTL` is an `optional` field that specifies the maximum TTL for the leases associated with this role.
Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to system/engine default TTL time.

```yaml
spec:
  maxTTL: "1h"
```

#### spec.revocationStatements

`spec.revocationStatements` is an `optional` field that specifies a list of database statements to be executed to revoke a user. The `{{name}}` value will be substituted. If not provided defaults to a generic drop user statement.

### MySQLRole Status

`status` shows the status of the MySQLRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a MySQLRole.
