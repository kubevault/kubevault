---
title: ElasticsearchRole | Vault Secret Engine
menu:
docs_{{ .version }}:
identifier: elasticsearch-database-crds
name: ElasticsearchRole
parent: database-crds-concepts
weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# ElasticsearchRole

## What is ElasticsearchRole

A `ElasticsearchRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create an Elasticsearch database secret engine role in a Kubernetes native way.

When a `ElasticsearchRole` is created, the KubeVault operator creates a Vault [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) according to the specification.
If the user deletes the `ElasticsearchRole` CRD, then the respective role will also be deleted from Vault.

![ElasticsearchRole CRD](/docs/images/concepts/mongodb_role.svg)

## ElasticsearchRole CRD Specification

Like any official Kubernetes resource, a `ElasticsearchRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `ElasticsearchRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: ElasticsearchRole
metadata:
  name: es-role
  namespace: demo
spec:
  vaultRef:
    name: vault-app
  databaseRef:
    name: es-app
    namespace: demo
  creationStatements:
    - "statement-0"
    - "statement-1"
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `ElasticsearchRole` crd.

### ElasticsearchRole Spec

ElasticsearchRole `spec` contains information that necessary for creating a database role.

```yaml
spec:
  vaultRef:
    name: <vault-appbinding-name>
  databaseRef:
    name: <database-appbinding-name>
    namespace: <database-appbinding-namespace>
  databaseName: <database-name>
  path: <secret-engine-path>
  defaultTTL: <default-ttl>
  maxTTL: <max-ttl>
  creationStatements:
    - "statement-0"
    - "statement-1"
  revocationStatements:
    - "statement-0"
```

ElasticsearchRole spec has the following fields:

#### spec.vaultRef

`spec.vaultRef` is a `required` field that specifies the name of an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) reference which is used to connect with a Vault server. AppBinding must be in the same namespace with the ElasticsearchRole object.

```yaml
spec:
  vaultRef:
    name: vault-app
```

#### spec.databaseRef

`spec.databaseRef` is an `optional` field that specifies the reference to an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains elasticsearch database connection information. It is used to generate the `db_name`. The naming format for `db_name` is: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

```yaml
spec:
  databaseRef:
    name: es-app
    namespace: demo
```

#### spec.databaseName

`spec.databaseName` is an `optional` field that specifies the `db_name`. It is used when `spec.databaseRef` is empty otherwise ignored.
Both `spec.databaseRef` and `spec.databaseName` cannot be empty at the same time.

```yaml
spec:
  databaseName: k8s.-.demo.es-app
```

#### spec.path

`spec.path` is an `optional` field that specifies the path where the secret engine is enabled. The default value is `database`.

```yaml
spec:
  path: my-es-path
```

#### spec.creationStatements

`spec.creationStatements` is a `required` field that specifies a list of database statements executed to create and configure a user.
See in [here](https://www.vaultproject.io/api/secret/databases/mongodb.html#creation_statements) for Vault documentation.

```yaml
spec:
  creationStatements:
    - "{"elasticsearch_roles": ["superuser"]}"
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

### ElasticsearchRole Status

`status` shows the status of the ElasticsearchRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation,
  which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of a ElasticsearchRole.
