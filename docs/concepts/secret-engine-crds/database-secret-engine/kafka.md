---
title: KafkaRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: kafkarole-database-crds
    name: KafkaRole
    parent: database-crds-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# KafkaRole

## What is KafkaRole

A `KafkaRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When a `KafkaRole` is created, the KubeVault operator creates a [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to specification.
If the user deletes the `KafkaRole` CRD, then the respective role will also be deleted from Vault.

## KafkaRole CRD Specification

Like any official Kubernetes resource, a `KafkaRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `KafkaRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: KafkaRole
metadata:
  name: kafka-role
  namespace: demo
spec:
  secretEngineRef:
    name: vault-app
  creationStatements:
    - '{"mechanism":"SCRAM-SHA-256","acls":[]}'
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `KafkaRole` crd.

### KafkaRole Spec

KafkaRole `spec` contains information that necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <vault-appbinding-name>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
  creationStatements:
    - "statement-0"
```

KafkaRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: kafka-secret-engine
```

#### spec.creationStatements

`spec.creationStatements` is a `required` field that specifies a JSON role document used to provision the user. The exact schema depends on the plugin — see the example below.

```yaml
spec:
  creationStatements:
    - '{"mechanism":"SCRAM-SHA-256","acls":[]}'
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

### KafkaRole Status

`status` shows the status of the KafkaRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a KafkaRole.
