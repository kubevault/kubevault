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

Dynamic credentials generated from a `QdrantRole` are issued as Qdrant Granular Access API Keys (JSON Web Tokens signed with HS256) with stateful token revocation using the `value_exists` validation claim.

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
    name: qdrant-engine
  defaultTTL: "1h"
  maxTTL: "24h"
  creationStatements:
    - |
      {
        "access": [
          {
            "collection": "my_collection",
            "access": "rw"
          }
        ]
      }
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
    name: <secret-engine-name>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
  creationStatements:
    - "statement-0"
```

QdrantRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: qdrant-secret-engine
```

#### spec.creationStatements

`spec.creationStatements` is a `required` field that specifies the permissions encoded into the dynamic Granular Access API Key (JWT).
You can specify collection-level granular access rules (with `r` for read-only or `rw` for read-write) or global permissions (`r` for read or `m` for manage):

```yaml
spec:
  creationStatements:
    - |
      {
        "access": [
          {
            "collection": "products",
            "access": "rw"
          },
          {
            "collection": "logs",
            "access": "r"
          }
        ]
      }
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
  maxTTL: "24h"
```

### QdrantRole Status

`status` shows the status of the QdrantRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions`: Represent observations of a QdrantRole.

## Namespace inheritance (tenant isolation)

A `QdrantRole` never sets or resolves an OpenBao namespace itself. It always inherits the
**effective namespace** of the `SecretEngine` it references via `spec.secretEngineRef`
(`SecretEngine.status.effectiveNamespace`) — empty for root, or the tenant's OpenBao
namespace once [tenant isolation](/docs/guides/tenant-isolation/overview.md) has placed
that engine in one. Every credential this role issues, and every revocation
(`SecretAccessRequest`), is scoped to that same namespace automatically — no
`QdrantRole`-level configuration is needed.
