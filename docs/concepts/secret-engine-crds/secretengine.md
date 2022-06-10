---
title: Secret Engine
menu:
  docs_{{ .version }}:
    identifier: secret-engine-crds
    name: SecretEngine
    parent: secret-engine-crds-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# SecretEngine

## What is a SecretEngine

A `SecretEngine` is a Kubernetes `CustomResourceDefinition` (CRD) which is designed to automate the process of enabling and configuring secret engines in Vault in a Kubernetes native way.

Secrets engines are components that store, generate,
or encrypt data. Secrets engines are provided some set of data, they take some action on that data, and they return a result.
Secrets engines are enabled at a "path" in Vault.
In this way, each secrets engine defines its paths and properties. To the user, secrets engines behave similar to a virtual filesystem, supporting operations like read, write, and delete.

When a `SecretEngine` CRD is created, the KubeVault operator will perform the following operations:

- **Creates** vault policy for the secret engine. The vault policy name follows the naming format:`k8s.{clusterName}.{metadata.namespace}.{metadata.name}`. For example, the policy for GCP secret engine is below:

```hcl
  path "<path>/config" {
        capabilities = ["create", "update", "read", "delete"]
  }
  
  path "<path>/roleset/*" {
      capabilities = ["create", "update", "read", "delete"]
  }
  
  path "<path>/token/*" {
      capabilities = ["create", "update", "read"]
  }
  
  path "<path>/key/*" {
      capabilities = ["create", "update", "read"]
  }
```

- **Updates** the Kubernetes auth role of the default k8s service account created with `VaultServer` with a new policy. The new policy will be merged with previous policies.

- **Enables** the secrets engine at a given path. By default, they are enabled at their "type" (e.g. "aws" is enabled at "aws/").

- **Configures** the secret engine with the given configuration.

## SecretEngine CRD Specification

Like any official Kubernetes resource, a `SecretEngine` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `SecretEngine` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: first-secret-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  gcp:
    credentialSecret: "gcp-cred"
status:
  observedGeneration: 1
  phase: Success
```

Here, we are going to describe the various sections of the `SecretEngine` crd.

### SecretEngine Spec

SecretEngine `.spec` contains information about the secret engine configuration and an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) reference which is used to connect with a Vault server.

SecretEngine `.spec` has the following fields:

#### spec.vaultRef

`spec.vaultRef` is a `required` field that specifies an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
reference which is used to connect with a Vault server. AppBinding must be on the same namespace of the secret engine crd.

```yaml
spec:
  vaultRef:
    name: vault
```

#### spec.path

`spec.path` will be set by the KubeVault operator.

Secret engines are enabled at a "path" in Vault. When a request comes to Vault, the router automatically routes anything with the route prefix to the secret engine. Since operator configures a secret engine to a specified path with SecretEngine resource, you can provide **only one secret engine configuration** out of the following ones:

- `spec.aws` : Specifies aws secret engine configuration

- `spec.azure`: Specifies azure secret engine configuration

- `spec.gcp`: Specifies gcp secret engine configuration

- `spec.postgres`: Specifies database(postgres) secret engine    configuration

- `spec.mongodb`: Specifies database(mongodb) secret engine      configuration

- `spec.mysql`: Specifies database(mysql) secret engine configuration

- `spec.elasticsearch`: Specifies database(elasticsearch) secret engine configuration


#### spec.aws

`spec.aws` specifies the configuration required to configure
AWS secret engine. [See more](https://www.vaultproject.io/api/secret/aws/index.html#parameters)

```yaml
spec:
  aws:
    credentialSecret: aws-cred
    region: us-east-1
    leaseConfig:
      lease: 1h
      leaseMax: 1h
```

- `aws.credentialSecret` : `Required`. Specifies the k8s secret name that contains AWS access key ID and secret access key
  
  ```yaml
    spec:
      aws:
        credentialSecret: <secret-name>
  ```

  The `data` field of the secret must contain the following key-value pairs:
  
  ```yaml
    data:
      access_kay: <access key>
      secret_key: <secret key>
  ```

- `aws.region` : `Required`. Specifies the AWS region.

- `aws.iamEndpoint` : `Optional`. Specifies a custom HTTP IAM endpoint to use.

- `aws.stsEndpoint` : `Optional`. Specifies a custom HTTP STS endpoint to use.

- `config.maxRetries` : `Optional`. Specifies the number of max retries the client should use for recoverable errors.

- `aws.leaseConfig` : `Optional`. Specifies the lease configuration.

  ```yaml
    config:
      leaseConfig:
        lease: 1h
        leaseMax: 1h
  ```

It has the following fields:

  - `leaseConfig.lease` : `Optional`. Specifies the lease value. Accepts time suffixed strings (eg, "1h").

  - `leaseConfig.leaseMax` : `Optional`. Specifies the maximum lease value. Accepts time suffixed strings (eg, "1h").

#### spec.azure

`spec.azure` specifies the configuration required to configure
Azure secret engine. [See more](https://www.vaultproject.io/api/secret/azure/index.html#configure-access)

```yaml
spec:
  azure:
    credentialSecret: azure-cred
    environment: AzurePublicCloud
```

- `credentialSecret` : `Required`. Specifies the k8s secret name containing azure credentials. The `data` field of the mentioned k8s secret can have the following key-value pairs.

  - `subscription-id` : `Required`. Specifies the subscription id for the Azure Active Directory.

  - `tenant-id` : `Required`. Specifies the tenant id for the Azure Active Directory.

  - `client-id` : `Optional`. Specifies the OAuth2 client id to connect to Azure.

  - `client-secret` : `Optional`. Specifies the OAuth2 client secret to connect to Azure.

  ```yaml
    data:
      subscription-id: <value>
      tenant-id: <value>
      client-id: <value>
      client-secret: <value>
  ```

- `environment` : `Optional`. Specifies the Azure environment. If not specified, Vault will use Azure Public Cloud.

#### spec.gcp

`spec.gcp` specifies the configuration required to configure GCP
secret engine. [See more](https://www.vaultproject.io/api/secret/gcp/index.html#write-config)

```yaml
  spec:
    gcp:
      credentialSecret: gcp-cred
      ttl: 0s
      maxTTL: 0s
```

- `credentialSecret` : `Required`. Specifies the k8s secret name that contains google application credentials.

  ```yaml
    spec:
      gcp:
        credentialSecret: <secret-name>  
  ```

  The `data` field of the mentioned k8s secret must contain the following key-value pair:
  
  ```yaml
    data:
      sa.json: <google-application-credential>
  ```

- `ttl` : `Optional`. Specifies default config TTL for long-lived credentials (i.e. service account keys). Default value is 0s.

- `maxTTL` : `Optional`. Specifies the maximum config TTL for long-lived credentials (i.e. service account keys). The default value is 0s.

#### spec.postgres

`spec.postgres` specifies the configuration required to configure PostgreSQL database secret engine. [See more](https://www.vaultproject.io/api/secret/databases/postgresql.html#configure-connection)

  ```yaml
    spec:
      postgres:
        databaseRef:
          name: <appbinding-name>
          namespace: <appbinding-namespace>
        pluginName: <plugin-name>
        allowedRoles:
          - "rule1"
          - "rule2"
        maxOpenConnections: <max-open-connection>
        maxIdleConnections: <max-idle-connection>
        maxConnectionLifetime: <max-connection-lifetime>
  ```

- `databaseRef` : `Required`. Specifies an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) reference that is required to connect to a PostgreSQL database. It is also used to generate `db_name` (i.e. `/v1/path/config/db_name`) where the database secret engine will be configured at. The name of the `db_name` follows this pattern: `k8s.{clusterName}.{namespace}.{name}`.
  
  - `name` : `Required`. Specifies the AppBinding name.

  - `namespace` : `Required`. Specifies the AppBinding namespace.

  ```yaml
    postgres:
      databaseRef:
        name: db-app
        namespace: demo
  ```

  The generated `db_name` for the above example will be: `k8s.-.demo.db-app`. If the cluster name is empty, it is replaced by "`-`".

- `pluginName` : `Optional`. Specifies the name of the plugin to use for this connection.
    Default plugin name is `postgres-database-plugin`.

  ```yaml
    postgres:
      pluginName: postgres-database-plugin
  ```

- `allowedRoles` : `Optional`. Specifies a list of roles allowed to use this connection.
    Default to `"*"` (i.e. any role can use this connection).

  ```yaml
    postgres:
      allowedRoles:
        - "readonly"
  ```

- `maxOpenConnections` : `Optional`. Specifies the maximum number of open connections to
    the database. Default value 2.

  ```yaml
    postgres:
      maxOpenConnections: 3
  ```

- `maxIdleConnections` : `Optional`.  Specifies the maximum number of idle connections to the database. Zero uses the value of max_open_connections and a negative value disables idle connections. If larger than max_open_connections it will be reduced to be equal. Default value 0.

  ```yaml
    postgres:
      maxIdleConnections: 1
  ```

- `maxConnectionLifetime` : `Optional`. Specifies the maximum amount of time a connection may be reused. If <= 0s connections are reused forever. Default value 0s.

  ```yaml
    postgres:
      maxConnectionLifetime: 5s
  ```

#### spec.mongodb

`spec.mongodb` specifies the configuration required to configure MongoDB database secret engine. [See more](https://www.vaultproject.io/api/secret/databases/mongodb.html#configure-connection)

```yaml
  spec:
    mongodb:
      databaseRef:
        name: <appbinding-name>
        namespace: <namespace>
      pluginName: <plugin-name>
      allowedRoles:
        - "role1"
        - "role2"
      writeConcern: <write-concern>
```

- `databaseRef` : `Required`. Specifies an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) reference that is required to connect to a MongoDB database. It is also used to generate `db_name` (i.e. `/v1/path/config/db_name`) where the database secret engine will be configured at. The naming of `db_name` follows: `k8s.{clusterName}.{namespace}.{name}`.

  - `name` : `Required`. Specifies the AppBinding name.

  - `namespace` : `Required`. Specifies the AppBinding namespace.

  ```yaml
    mongodb:
      databaseRef:
        name: db-app
        namespace: demo
  ```

  The generated `db_name` for the above example will be: `k8s.-.demo.db-app`. If the cluster name is empty, it is replaced by "`-`".

- `pluginName` : `Optional`. Specifies the name of the plugin to use for this connection.
Default plugin name is `mongodb-database-plugin`.

  ```yaml
    mongodb:
      pluginName: mongodb-database-plugin
  ```

- `allowedRoles` : `Optional`. Specifies a list of roles allowed to use this connection.
    Default to `"*"` (i.e. any role can use this connection).

  ```yaml
    mongodb:
      allowedRoles:
        - "readonly"
  ```

- `writeConcern` : `Optional`. Specifies the MongoDB write concern.
  This is set for the entirety of the session, maintained for the life cycle of the plugin process. Must be a serialized JSON object,
  or a base64-encoded serialized JSON object. The JSON payload values map to the values in the Safe struct from the mongo driver.

  ```yaml
    mongodb:
      writeConcern: `{ \"wmode\": \"majority\", \"wtimeout\": 5000 }`
  ```

#### spec.mysql

`spec.mysql` specifies the configuration required to configure MySQL database secret engine. [See more](https://www.vaultproject.io/api/secret/databases/mysql-maria.html#configure-connection)

  ```yaml
    spec:
      mysql:
        databaseRef:
          name: <appbinding-name>
          namespace: <appbinding-namespace>
        pluginName: <plugin-name>
        allowedRoles:
          - "role1"
          - "role2"
          - ... ...
        maxOpenConnections: <max-open-connections>
        maxIdleConnections: <max-idle-connections>
        maxConnectionLifetime: <max-connection-lifetime>
  ```

- `databaseRef` : `Required`. Specifies an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) reference that is required to connect to a MySQL database. It is also used to generate `db_name` (i.e. `/v1/path/config/db_name`) where the database secret engine will be configured at. The naming of `db_name` follows: `k8s.{clusterName}.{namespace}.{name}`.

  - `name` : `Required`. Specifies the AppBinding name.

  - `namespace` : `Required`. Specifies the AppBinding namespace.

  ```yaml
    mysql:
      databaseRef:
        name: db-app
        namespace: demo
  ```

  The generated `db_name` for the above example will be: `k8s.-.demo.db-app`. If the cluster name is empty, it is replaced by "`-`".

- `pluginName` : `Optional`. Specifies the name of the plugin to use for this connection.
    The default plugin name is `mysql-database-plugin`.

  ```yaml
    mysql:
      pluginName: mysql-database-plugin
  ```

- `allowedRoles` : `Optional`. Specifies a list of roles allowed to use this connection.
    Default to `"*"` (i.e. any role can use this connection).

  ```yaml
    mysql:
      allowedRoles:
        - "readonly"
  ```

- `maxOpenConnections` : `Optional`. Specifies the maximum number of open connections to the database. Default value 2.

  ```yaml
    mysql:
      maxOpenConnections: 3
  ```

- `maxIdleConnections` : `Optional`.  Specifies the maximum number of idle connections to the database. Zero uses the value of max_open_connections and a negative value disables idle connections. If larger than max_open_connections it will be reduced to be equal. Default value 0.

  ```yaml
    mysql:
      maxIdleConnections: 1
  ```

- `maxConnectionLifetime` : `Optional`. Specifies the maximum amount of time a connection may be reused. If <= 0s connections are reused forever. Default value 0s.

  ```yaml
    mysql:
      maxConnectionLifetime: 5s
  ```

#### spec.elasticsearch

`spec.elasticsearch` specifies the configuration required to configure Elasticsearch database secret engine. [See more](https://www.vaultproject.io/api/secret/databases/elasticdb.html#configure-connection)

  ```yaml
    spec:
      elasticsearch:
        databaseRef:
          name: <appbinding-name>
          namespace: <appbinding-namespace>
        pluginName: <plugin-name>
        allowedRoles:
          - "role1"
          - "role2"
          - ... ...
  ```

- `databaseRef` : `Required`. Specifies an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) reference that is required to connect to an Elasticsearch database. It is also used to generate `db_name` (i.e. `/v1/path/config/db_name`) where the database secret engine will be configured at. The naming of `db_name` follows: `k8s.{clusterName}.{namespace}.{name}`.

  - `name` : `Required`. Specifies the AppBinding name.

  - `namespace` : `Required`. Specifies the AppBinding namespace.

  ```yaml
    elasticsearch:
      databaseRef:
        name: db-app
        namespace: demo
  ```

  The generated `db_name` for the above example will be: `k8s.-.demo.db-app`. If the cluster name is empty, it is replaced by "`-`".

- `pluginName` : `Optional`. Specifies the name of the plugin to use for this connection.
  The default plugin name is `elasticsearch-database-plugin`.

  ```yaml
    elasticsearch:
      pluginName: elasticsearch-database-plugin
  ```

- `allowedRoles` : `Optional`. Specifies a list of roles allowed to use this connection.
  Default to `"*"` (i.e. any role can use this connection).

  ```yaml
    elasticsearch:
      allowedRoles:
        - "readonly"
  ```

### SecretEngine Status

`status` shows the status of the SecretEngine. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation, which is updated on mutation by the API Server.

- `phase`: Indicates whether the secret engine successfully configured in the Vault or not.

- `conditions` : Represent observations of a SecretEngine.
