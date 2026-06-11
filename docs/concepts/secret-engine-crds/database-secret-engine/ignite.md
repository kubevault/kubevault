---
title: IgniteRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: igniterole-database-crds
    name: IgniteRole
    parent: database-crds-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# IgniteRole

## What is IgniteRole

A `IgniteRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When a `IgniteRole` is created, the KubeVault operator creates a [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to specification.
If the user deletes the `IgniteRole` CRD, then the respective role will also be deleted from Vault.

## IgniteRole CRD Specification

Like any official Kubernetes resource, a `IgniteRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `IgniteRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: IgniteRole
metadata:
  name: ignite-role
  namespace: demo
spec:
  secretEngineRef:
    name: vault-app
  creationStatements:
    - "CREATE USER \"{{name}}\" WITH PASSWORD '{{password}}';"
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `IgniteRole` crd.

### IgniteRole Spec

IgniteRole `spec` contains information that necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <vault-appbinding-name>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
  creationStatements:
    - "statement-0"
    - "statement-1"
  revocationStatements:
    - "statement-0"
  rollbackStatements:
    - "statement-0"
```

IgniteRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: ignite-secret-engine
```

#### spec.creationStatements

`spec.creationStatements` is a `required` field that specifies a list of database statements executed to create and configure a user.
The `{{name}}` and `{{password}}` values will be substituted by Vault.

```yaml
spec:
  creationStatements:
    - "CREATE USER \"{{name}}\" WITH PASSWORD '{{password}}';"
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

#### spec.rollbackStatements

`spec.rollbackStatements` is an `optional` field that specifies a list of database statements to be executed
rollback a create operation in the event of an error. Not every plugin type will support this functionality.

### IgniteRole Status

`status` shows the status of the IgniteRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a IgniteRole.
