---
title: QdrantRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: qdrantrole-database-crds
    name: QdrantRole
    parent: database-crds-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# QdrantRole

## What is QdrantRole

A `QdrantRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When a `QdrantRole` is created, the KubeVault operator creates a [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to specification.
If the user deletes the `QdrantRole` CRD, then the respective role will also be deleted from Vault.

> Note: The `qdrant-database-plugin` shipped in [openbao/openbao#17](https://github.com/openbao/openbao/pull/17) is **static-credentials-only**. It does NOT implement dynamic credential issuance (`NewUser`) — Qdrant loads its API key from the `QDRANT__SERVICE__API_KEY` environment variable at server startup and exposes no runtime user-management API. The `QdrantRole` CRD therefore only attaches role metadata (`db_name`, `default_ttl`, `max_ttl`) to the pre-existing Qdrant API key. Operators wire up key rotation by calling `bao write database/static-roles/<role>` against this metadata; from then on OpenBao rotates the API key on the configured cadence and exposes the current value at `database/static-creds/<role>` under the `password` field.

## QdrantRole CRD Specification

Like any official Kubernetes resource, a `QdrantRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `QdrantRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: QdrantRole
metadata:
  name: qdrant-role
  namespace: demo
spec:
  secretEngineRef:
    name: vault-app
  defaultTTL: "1h"
  maxTTL: "24h"
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `QdrantRole` CRD.

### QdrantRole Spec

QdrantRole `spec` contains information that is necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <vault-appbinding-name>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
```

QdrantRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: qdrant-secret-engine
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

### QdrantRole Status

`status` shows the status of the QdrantRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a QdrantRole.
