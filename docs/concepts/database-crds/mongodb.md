---
title: MongoDBRole | Vault Secret Engine
menu:
  docs_0.2.0:
    identifier: mongodb-database-crds
    name: MongoDBRole
    parent: database-crds-concepts
    weight: 10
menu_name: docs_0.2.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# MongoDBRole CRD

Vault operator will create [database connection config](https://www.vaultproject.io/api/secret/databases/mongodb.html#configure-connection) and [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to `MongoDBRole` CRD (CustomResourceDefinition) specification. If the user deletes the `MongoDBRole` CRD, then respective role will also be deleted from Vault.

```yaml
apiVersion: authorization.kubedb.com/v1alpha1
kind: MongoDBRole
metadata:
  name: <name>
  namespace: <namespace>
spec:
  ...
status:
  ...
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{spec.clusterName}.{spec.namespace}.{spec.name}`

## MongoDBRole Spec

MongoDBRole `spec` contains information that necessary for creating database config and role.

```yaml
apiVersion: authorization.kubedb.com/v1alpha1
kind: MongoDBRole
metadata:
  name: mongo-test
  namespace: demo
spec:
  creationStatements:
    - "{ \"db\": \"admin\", \"roles\": [{ \"role\": \"readWrite\" }, {\"role\": \"read\", \"db\": \"foo\"}] }"
  defaultTTL: 300s
  maxTTL: 24h
  authManagerRef:
    namespace: demo
    name: vault-app
  databaseRef:
    name: mongo-app
```

MongoDBRole Spec has following fields:

### spec.authManagerRef

`spec.authManagerRef` specifies the name and namespace of [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains information to communicate with Vault.

```yaml
spec:
  authManagerRef:
    namespace: demo
    name: vault-app
```

### spec.databaseRef

`spec.databaseRef` is a required field that specifies the name of [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains mongodb database connection information. This should be in the same namespace of the `MongoDBRole` CRD.

```yaml
spec:
  databaseRef:
    name: mongo-app
```

### spec.creationStatements

`spec.creationStatements` is a required field that specifies the database statements executed to create and configure a user. See in [here](https://www.vaultproject.io/api/secret/databases/mongodb.html#creation_statements) for Vault documentation.

```yaml
spec:
  creationStatements:
    - "{ \"db\": \"admin\", \"roles\": [{ \"role\": \"readWrite\" }, {\"role\": \"read\", \"db\": \"foo\"}] }"
```

### spec.defaultTTL

`spec.defaultTTL` is an optional field that specifies the TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to system/engine default TTL time.

```yaml
spec:
  defaultTTL: "1h"
```

### spec.maxTTL

`spec.maxTTL` is an optional field that specifies the maximum TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to system/engine default TTL time.

```yaml
spec:
  maxTTL: "1h"
```

### spec.revocationStatements

`spec.revocationStatements` is an optional field that specifies the database statements to be executed to revoke a user. See in [here](https://www.vaultproject.io/api/secret/databases/mongodb.html#revocation_statements) for Vault documentation.

## MongoDBRole Status

`status` shows the status of the MongoDBRole. It is maintained by Vault operator. It contains following fields:

- `status` : Indicates whether the role successfully applied in vault or not or in progress or failed

- `conditions` : Represent observations of a MongoDBRole.
