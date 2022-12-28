---
title: RedisRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: redis-database-crds
    name: RedisRole
    parent: database-crds-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# RedisRole

## What is RedisRole

A `RedisRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a Redis database secret engine role in a Kubernetes native way.

When a `RedisRole` is created, the KubeVault operator creates a Vault [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to the specification.
If the user deletes the `RedisRole` CRD, then the respective role will also be deleted from Vault.

![RedisRole CRD](/docs/images/concepts/redis_role.svg)

## RedisRole CRD Specification

Like any official Kubernetes resource, a `RedisRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `RedisRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: RedisRole
metadata:
  name: rd-role
  namespace: demo
spec:
  secretEngineRef:
    name: redis-secret-engine
  creationStatements:
    - "statement-0"
    - "statement-1"
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `RedisRole` crd.

### RedisRole Spec

RedisRole `spec` contains information that necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <secret-engine-name>
  creationStatements:
    - "statement-0"
    - "statement-1"
  revocationStatements:
    - "statement-0"
```

RedisRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: redis-secret-engine
```

#### spec.creationStatements

`spec.creationStatements` is a `required` field that specifies a list of database statements executed to create and configure a user.
See in [here](https://developer.hashicorp.com/vault/api-docs/secret/databases/redis#creation_statements) for Vault documentation.

```yaml
spec:
  creationStatements:
    - '["~*", "+@read","+@write"]'
```

#### spec.defaultTTL

`spec.defaultTTL` is an `optional` field that specifies the TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds.
 Defaults to system/engine default TTL time.

```yaml
spec:
  defaultTTL: "1h"
```

#### spec.maxTTL

`spec.maxTTL` is an `optional` field that specifies the maximum TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds.
Defaults to system/engine default TTL time.

```yaml
spec:
  maxTTL: "1h"
```

#### spec.revocationStatements

`spec.revocationStatements` is an `optional` field that specifies
a list of database statements to be executed to revoke a user.
See [here](https://www.vaultproject.io/api/secret/databases/redis.html#revocation_statements) for Vault documentation.

### RedisRole Status

`status` shows the status of the RedisRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a RedisRole.
