---
title: DB2Role | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: db2role-database-crds
    name: DB2Role
    parent: database-crds-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# DB2Role

## What is DB2Role

A `DB2Role` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create a database secret engine role in a Kubernetes native way.

When a `DB2Role` is created, the KubeVault operator creates a [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to specification.
If the user deletes the `DB2Role` CRD, then the respective role will also be deleted from Vault.

> Note: The `db2-database-plugin` shipped in [openbao/openbao#19](https://github.com/openbao/openbao/pull/19) is **static-credentials-only**. It does NOT implement dynamic credential issuance (`NewUser`) — IBM Db2 user accounts must be provisioned out-of-band against the underlying OS / LDAP realm that Db2 authenticates against. The `DB2Role` CRD therefore only attaches role metadata (`db_name`, `default_ttl`, `max_ttl`) to a pre-existing Db2 principal. Operators wire up password rotation by calling `bao write database/static-roles/<role>` against this metadata; from then on OpenBao rotates the principal's password on the configured cadence and exposes the current credential at `database/static-creds/<role>`.

## DB2Role CRD Specification

Like any official Kubernetes resource, a `DB2Role` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `DB2Role` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: DB2Role
metadata:
  name: db2-role
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

Here, we are going to describe the various sections of the `DB2Role` CRD.

### DB2Role Spec

DB2Role `spec` contains information that is necessary for creating a database role.

```yaml
spec:
  secretEngineRef:
    name: <vault-appbinding-name>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
```

DB2Role spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: db2-secret-engine
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

### DB2Role Status

`status` shows the status of the DB2Role. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
    which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a DB2Role.
