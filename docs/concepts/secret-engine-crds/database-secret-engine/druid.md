---
title: DruidRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: druidrole-database-crds
    name: DruidRole
    parent: database-crds-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# DruidRole

## What is DruidRole

A `DruidRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When a `DruidRole` is created, the KubeVault operator creates a [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to specification.
If the user deletes the `DruidRole` CRD, then the respective role will also be deleted from Vault.

## DruidRole CRD Specification

Like any official Kubernetes resource, a `DruidRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `DruidRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: DruidRole
metadata:
  name: druid-role
  namespace: demo
spec:
  secretEngineRef:
    name: vault-app
  creationStatements:
    - '{"roles":["admin","read"]}'
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `DruidRole` crd.

### DruidRole Spec

DruidRole `spec` contains information that necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <vault-appbinding-name>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
  creationStatements:
    - "statement-0"
```

DruidRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: druid-secret-engine
```

#### spec.creationStatements

`spec.creationStatements` is a `required` field that specifies a JSON role document used to provision the user. The exact schema depends on the plugin — see the example below.

```yaml
spec:
  creationStatements:
    - '{"roles":["admin","read"]}'
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

### DruidRole Status

`status` shows the status of the DruidRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a DruidRole.
