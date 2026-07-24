---
title: RabbitMQRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: rabbitmqrole-database-crds
    name: RabbitMQRole
    parent: database-crds-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# RabbitMQRole

## What is RabbitMQRole

A `RabbitMQRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When a `RabbitMQRole` is created, the KubeVault operator creates a [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to specification.
If the user deletes the `RabbitMQRole` CRD, then the respective role will also be deleted from Vault.

## RabbitMQRole CRD Specification

Like any official Kubernetes resource, a `RabbitMQRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `RabbitMQRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: RabbitMQRole
metadata:
  name: rabbitmq-role
  namespace: demo
spec:
  secretEngineRef:
    name: vault-app
  creationStatements:
    - '{"tags":"administrator","vhosts":{"/":{"configure":".*","write":".*","read":".*"}}}'
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `RabbitMQRole` crd.

### RabbitMQRole Spec

RabbitMQRole `spec` contains information that necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <vault-appbinding-name>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
  creationStatements:
    - "statement-0"
```

RabbitMQRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: rabbitmq-secret-engine
```

#### spec.creationStatements

`spec.creationStatements` is a `required` field that specifies a JSON role document used to provision the user. The exact schema depends on the plugin — see the example below.

```yaml
spec:
  creationStatements:
    - '{"tags":"administrator","vhosts":{"/":{"configure":".*","write":".*","read":".*"}}}'
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

### RabbitMQRole Status

`status` shows the status of the RabbitMQRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a RabbitMQRole.

## Namespace inheritance (tenant isolation)

A `RabbitMQRole` never sets or resolves an OpenBao namespace itself. It always inherits the
**effective namespace** of the `SecretEngine` it references via `spec.secretEngineRef`
(`SecretEngine.status.effectiveNamespace`) — empty for root, or the tenant's OpenBao
namespace once [tenant isolation](/docs/guides/tenant-isolation/overview.md) has placed
that engine in one. Every credential this role issues, and every revocation
(`SecretAccessRequest`), is scoped to that same namespace automatically — no
`RabbitMQRole`-level configuration is needed.
