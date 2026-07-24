---
title: ZooKeeperRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: zookeeperrole-database-crds
    name: ZooKeeperRole
    parent: database-crds-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# ZooKeeperRole

## What is ZooKeeperRole

A `ZooKeeperRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When a `ZooKeeperRole` is created, the KubeVault operator creates a [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to specification.
If the user deletes the `ZooKeeperRole` CRD, then the respective role will also be deleted from Vault.

> Note: The `zookeeper-database-plugin` shipped in [openbao/openbao#21](https://github.com/openbao/openbao/pull/21) is **static-credentials-only**. It does NOT implement dynamic credential issuance (`NewUser`) — Apache ZooKeeper SASL/DIGEST principals are loaded from the JAAS config file at server startup, so user accounts must be provisioned out-of-band in the ensemble's JAAS configuration. The `ZooKeeperRole` CRD therefore only attaches role metadata (`db_name`, `default_ttl`, `max_ttl`) to a pre-existing ZooKeeper principal. Operators wire up password rotation by calling `bao write database/static-roles/<role>` against this metadata; from then on OpenBao rotates the principal's password on the configured cadence and exposes the current credential at `database/static-creds/<role>`.

## ZooKeeperRole CRD Specification

Like any official Kubernetes resource, a `ZooKeeperRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `ZooKeeperRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: ZooKeeperRole
metadata:
  name: zookeeper-role
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

Here, we are going to describe the various sections of the `ZooKeeperRole` CRD.

### ZooKeeperRole Spec

ZooKeeperRole `spec` contains information that is necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <vault-appbinding-name>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
```

ZooKeeperRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: zookeeper-secret-engine
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

### ZooKeeperRole Status

`status` shows the status of the ZooKeeperRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a ZooKeeperRole.

## Namespace inheritance (tenant isolation)

A `ZooKeeperRole` never sets or resolves an OpenBao namespace itself. It always inherits the
**effective namespace** of the `SecretEngine` it references via `spec.secretEngineRef`
(`SecretEngine.status.effectiveNamespace`) — empty for root, or the tenant's OpenBao
namespace once [tenant isolation](/docs/guides/tenant-isolation/overview.md) has placed
that engine in one. Every credential this role issues, and every revocation
(`SecretAccessRequest`), is scoped to that same namespace automatically — no
`ZooKeeperRole`-level configuration is needed.
