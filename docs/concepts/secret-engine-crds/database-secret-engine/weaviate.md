---
title: WeaviateRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: weaviaterole-database-crds
    name: WeaviateRole
    parent: database-crds-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# WeaviateRole

## What is WeaviateRole

A `WeaviateRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When a `WeaviateRole` is created, the KubeVault operator creates a [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to specification.
If the user deletes the `WeaviateRole` CRD, then the respective role will also be deleted from Vault.

Dynamic credentials generated from a `WeaviateRole` create a dynamic user in Weaviate, assign the configured RBAC roles, and return a dynamic API key.

## WeaviateRole CRD Specification

Like any official Kubernetes resource, a `WeaviateRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `WeaviateRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: WeaviateRole
metadata:
  name: weaviate-role
  namespace: demo
spec:
  secretEngineRef:
    name: weaviate-engine
  defaultTTL: "1h"
  maxTTL: "24h"
  creationStatements:
    - |
      {
        "roles": [
          "viewer"
        ]
      }
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `WeaviateRole` CRD.

### WeaviateRole Spec

WeaviateRole `spec` contains information that is necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <secret-engine-name>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
  creationStatements:
    - "statement-0"
```

WeaviateRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: weaviate-secret-engine
```

#### spec.creationStatements

`spec.creationStatements` is a `required` field that specifies the roles assigned to the dynamically-created user in Weaviate:

```yaml
spec:
  creationStatements:
    - |
      {
        "roles": [
          "viewer"
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

### WeaviateRole Status

`status` shows the status of the WeaviateRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions`: Represent observations of a WeaviateRole.

## Namespace inheritance (tenant isolation)

A `WeaviateRole` never sets or resolves an OpenBao namespace itself. It always inherits the
**effective namespace** of the `SecretEngine` it references via `spec.secretEngineRef`
(`SecretEngine.status.effectiveNamespace`) — empty for root, or the tenant's OpenBao
namespace once [tenant isolation](/docs/guides/tenant-isolation/overview.md) has placed
that engine in one. Every credential this role issues, and every revocation
(`SecretAccessRequest`), is scoped to that same namespace automatically — no
`WeaviateRole`-level configuration is needed.
