---
title: DatabaseAccessRequest | Vault Secret Engine
menu:
  docs_0.2.0:
    identifier: databaseaccessrequest-database-crds
    name: DatabaseAccessRequest
    parent: database-crds-concepts
    weight: 100
menu_name: docs_0.2.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# DatabaseAccessRequest CRD

`DatabaseAccessRequest` CRD is to request database credential from vault. If `DatabaseAccessRequest` is approved, then Vault operator will issue credential from vault and create Kubernetes secret containing credential. The secret name will be specified in `status.secret.name` field.

```yaml
apiVersion: authorization.kubedb.com/v1alpha1
kind: DatabaseAccessRequest
metadata:
  name: <name>
  namespace: <namespace>
spec:
  ...
status:
  ...
```

## DatabaseAccessRequest Spec

DatabaseAccessRequest `spec` contains information about database role and subject.

```yaml
apiVersion: authorization.kubedb.com/v1alpha1
kind: DatabaseAccessRequest
metadata:
  name: postgres-cred
  namespace: demo
spec:
  roleRef:
    kind: PostgresRole
    name: postgres-test
    namespace: default
  subjects:
    - kind: ServiceAccount
      name: pgdb-sa
      namespace: demo
```

DatabaseAccessRequest Spec has following fields:

### spec.roleRef

`spec.roleRef` is a required field that specifies the database role against which credential will be issued.

```yaml
spec:
  roleRef:
    kind: PostgresRole
    name: postgres-test
    namespace: demo
```

It has following field:

- `roleRef.kind` :  `Required`. Specifies the kind of object being referenced. Values are `MongoDBRole`, `MySQLRole`, and `PostgresRole`.
- `roleRef.name` : `Required`. Specifies the name of the object being referenced.
- `roleRef.namespace` : `Required`. Specifies the namespace of the referenced object.

### spec.subjects

`spec.subjects` is a required field that contains a reference to the object or user identities a role binding applies to. It will have read access of the credential secret. This can either hold a direct API object reference, or a value for non-objects such as user and group names.

```yaml
spec:
  subjects:
    - kind: ServiceAccount
      name: pgdb-sa
      namespace: demo
```

### spec.ttl

`spec.ttl` is an optional field that specifies the TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to roles default TTL time.

```yaml
spec:
  ttl: "1h"
```

## DatabaseAccessRequest Status

`status` shows the status of the DatabaseAccessRequest. It is maintained by Vault operator. It contains following fields:

- `secret` : Specifies the name of the secret containing database credential.

- `lease` : Contains lease information of the issued credential.

- `conditions` : Represent observations of a DatabaseAccessRequest.

    ```yaml
    status:
      conditions:
        - type: Approved
    ```

  It has following field:
  - `conditions[].type` : `Required`. Specifies request approval state. Supported type: `Approved` and `Denied`.
  - `conditions[].reason` : `Optional`. Specifies brief reason for the request state.
  - `conditions[].message` : `Optional`. Specifies human readable message with details about the request state.

> Note: Database credential will be issued if `conditions[].type` is `Approved`. Otherwise, Vault operator will not issue any credential.
