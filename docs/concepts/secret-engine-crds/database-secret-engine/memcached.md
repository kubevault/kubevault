---
title: MemcachedRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: memcachedrole-database-crds
    name: MemcachedRole
    parent: database-crds-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# MemcachedRole

## What is MemcachedRole

A `MemcachedRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When a `MemcachedRole` is created, the KubeVault operator creates a [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to specification.
If the user deletes the `MemcachedRole` CRD, then the respective role will also be deleted from Vault.

> Note: The `memcached-database-plugin` shipped in [openbao/openbao#16](https://github.com/openbao/openbao/pull/16) is **static-credentials-only**. It does NOT implement dynamic credential issuance (`NewUser`) — Memcached SASL accounts live in the server's SASL DB (`sasldb2`) and are provisioned with `saslpasswd2` out-of-band; there is no runtime user-management API. The `MemcachedRole` CRD therefore only attaches role metadata (`db_name`, `default_ttl`, `max_ttl`) to a pre-existing Memcached SASL principal. Operators wire up password rotation by calling `bao write database/static-roles/<role>` against this metadata; from then on OpenBao rotates the principal's password on the configured cadence and exposes the current credential at `database/static-creds/<role>`.

## MemcachedRole CRD Specification

Like any official Kubernetes resource, a `MemcachedRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `MemcachedRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: MemcachedRole
metadata:
  name: memcached-role
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

Here, we are going to describe the various sections of the `MemcachedRole` CRD.

### MemcachedRole Spec

MemcachedRole `spec` contains information that is necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <vault-appbinding-name>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
```

MemcachedRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: memcached-secret-engine
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

### MemcachedRole Status

`status` shows the status of the MemcachedRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a MemcachedRole.
