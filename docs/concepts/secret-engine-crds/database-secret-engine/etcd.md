---
title: EtcdRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: etcdrole-database-crds
    name: EtcdRole
    parent: database-crds-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# EtcdRole

## What is EtcdRole

An `EtcdRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When an `EtcdRole` is created, the KubeVault operator creates a [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to specification.
If the user deletes the `EtcdRole` CRD, then the respective role will also be deleted from Vault.

## EtcdRole CRD Specification

Like any official Kubernetes resource, an `EtcdRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `EtcdRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: EtcdRole
metadata:
  name: etcd-role
  namespace: demo
spec:
  secretEngineRef:
    name: vault-app
  creationStatements:
    - '{"roles":["reader"]}'
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `EtcdRole` crd.

### EtcdRole Spec

EtcdRole `spec` contains information that necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <vault-appbinding-name>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
  creationStatements:
    - "statement-0"
```

EtcdRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: etcd-secret-engine
```

#### spec.creationStatements

`spec.creationStatements` is a `required` field that specifies a single-element list holding a JSON role document naming pre-existing etcd roles to grant the new user (e.g. via `etcdctl role add`). The plugin only grants roles via `UserGrantRole`; it does not create them.

```yaml
spec:
  creationStatements:
    - '{"roles":["reader","writer"]}'
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

### EtcdRole Status

`status` shows the status of the EtcdRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of an EtcdRole.

## Namespace inheritance (tenant isolation)

An `EtcdRole` never sets or resolves an OpenBao namespace itself. It always inherits the
**effective namespace** of the `SecretEngine` it references via `spec.secretEngineRef`
(`SecretEngine.status.effectiveNamespace`) — empty for root, or the tenant's OpenBao
namespace once [tenant isolation](/docs/guides/tenant-isolation/overview.md) has placed
that engine in one. Every credential this role issues, and every revocation
(`SecretAccessRequest`), is scoped to that same namespace automatically — no
`EtcdRole`-level configuration is needed.
