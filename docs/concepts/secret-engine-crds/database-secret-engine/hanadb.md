---
title: HanaDBRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: hanadbrole-database-crds
    name: HanaDBRole
    parent: database-crds-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# HanaDBRole

## What is HanaDBRole

A `HanaDBRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When a `HanaDBRole` is created, the KubeVault operator creates a [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to specification.
If the user deletes the `HanaDBRole` CRD, then the respective role will also be deleted from Vault.

## HanaDBRole CRD Specification

Like any official Kubernetes resource, a `HanaDBRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `HanaDBRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: HanaDBRole
metadata:
  name: hana-role
  namespace: demo
spec:
  secretEngineRef:
    name: vault-app
  creationStatements:
    - "CREATE USER \"{{name}}\" PASSWORD \"{{password}}\" NO FORCE_FIRST_PASSWORD_CHANGE VALID UNTIL '{{expiration}}';"
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `HanaDBRole` crd.

### HanaDBRole Spec

HanaDBRole `spec` contains information that necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <vault-appbinding-name>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
  creationStatements:
    - "statement-0"
    - "statement-1"
  revocationStatements:
    - "statement-0"
  rollbackStatements:
    - "statement-0"
  renewStatements:
    - "statement-0"
```

HanaDBRole spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: hana-secret-engine
```

#### spec.creationStatements

`spec.creationStatements` is a `required` field that specifies a list of database statements executed to create and configure a user.
The `{{name}}`, `{{password}}` and `{{expiration}}` values will be substituted by Vault.

```yaml
spec:
  creationStatements:
    - "CREATE USER \"{{name}}\" PASSWORD \"{{password}}\" NO FORCE_FIRST_PASSWORD_CHANGE VALID UNTIL '{{expiration}}';"
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

#### spec.revocationStatements

`spec.revocationStatements` is an `optional` field that specifies a list of database statements to be executed to revoke a user. The `{{name}}` value will be substituted. If not provided defaults to a generic drop user statement.

#### spec.rollbackStatements

`spec.rollbackStatements` is an `optional` field that specifies a list of database statements to be executed
rollback a create operation in the event of an error. Not every plugin type will support this functionality.

#### spec.renewStatements

`spec.renewStatements` is an `optional` field that specifies a list of database statements to be executed to renew a user. Not every plugin type will support this functionality.

### HanaDBRole Status

`status` shows the status of the HanaDBRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a HanaDBRole.

## Namespace inheritance (tenant isolation)

A `HanaDBRole` never sets or resolves an OpenBao namespace itself. It always inherits the
**effective namespace** of the `SecretEngine` it references via `spec.secretEngineRef`
(`SecretEngine.status.effectiveNamespace`) — empty for root, or the tenant's OpenBao
namespace once [tenant isolation](/docs/guides/tenant-isolation/overview.md) has placed
that engine in one. Every credential this role issues, and every revocation
(`SecretAccessRequest`), is scoped to that same namespace automatically — no
`HanaDBRole`-level configuration is needed.
