---
title: MariaDBRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: mariadb-database-crds
    name: MariaDBRole
    parent: database-crds-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# MariaDBRole

## What is MariaDBRole

A `MariaDBRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When a `MariaDBRole` is created, the KubeVault operator creates a
[role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to the specification.
If the user deletes the `MariaDBRole` CRD, then the respective role will also be deleted from Vault.

![MariaDBRole CRD](/docs/images/concepts/mariadb_role.svg)

## MariaDBRole CRD Specification

Like any official Kubernetes resource, a `MariaDBRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `MariaDBRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: MariaDBRole
metadata:
  name: mariadb-role
  namespace: demo
spec:
  secretEngineRef:
    name: mariadb-secret-engine
  creationStatements:
    - "CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';"
    - "GRANT SELECT ON *.* TO '{{name}}'@'%';"
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `MariaDBRole` crd.

### MariaDBRole Spec

MariaDBRole `spec` contains information that necessary for creating a database role.

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

MariaDBRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: mariadb-secret-engine
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

### MariaDBRole Status

`status` shows the status of the MariaDBRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a MariaDBRole.
