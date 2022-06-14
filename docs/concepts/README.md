---
title: KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: concepts-readme
    name: Concepts
    parent: concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
url: /docs/{{ .version }}/concepts/
aliases:
  - /docs/{{ .version }}/concepts/README/
---

# Concepts

Concepts help you learn about the different parts of KubeVault and the abstractions it uses.

- What is KubeVault?
  - [Overview](/docs/concepts/overview.md). Provides an introduction to KubeVault operator, including the problems it solves and its use cases.
  - [Operator architecture](/docs/concepts/architecture.md). Provides a high-level illustration of the architecture of the KubeVault operator.

<ul class="nav nav-tabs" id="conceptsTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="vault-server-tab" data-toggle="tab" href="#vault-server" role="tab" aria-controls="vault-server" aria-selected="true">Vault Server</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="secret-engine-tab" data-toggle="tab" href="#secret-engine" role="tab" aria-controls="secret-engine" aria-selected="false">Secret Engines</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="vault-policy-tab" data-toggle="tab" href="#vault-policy" role="tab" aria-controls="vault-policy" aria-selected="false">Vault Policies</a>
  </li>
</ul>
<div class="tab-content" id="conceptsTabContent">
  <div class="tab-pane fade show active" id="vault-server" role="tabpanel" aria-labelledby="vault-server-tab">

## AppBinding

Introduces a way to specify `connection information`, `credential`, and `parameters` that are necessary for communicating with an app or service.

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)

## Vault Server Version

Introduces the concept of `VaultServerVersion` to specify the docker images of `HashiCorp Vault`, `Unsealer`, and `Exporter`.

- [VaultServerVersion](/docs/concepts/vault-server-crds/vaultserverversion.md)

## Vault Server

A `VaultServer` is a `Kubernetes CustomResourceDefinition (CRD)` which is used to deploy a `HashiCorp Vault` server on Kubernetes clusters in a Kubernetes native way. Introduces the concept of `VaultServer` for configuring a HashiCorp Vault server in a Kubernetes native way.

- [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md)

### Vault Unsealer Options
When a `Vault` server is started, it starts in a `sealed` state. In this state, Vault is configured to know where and how to access the physical storage, but doesn't know how to decrypt any of it.

`Unsealing` is the process of obtaining the plaintext root key necessary to read the decryption key to decrypt the data, allowing access to the Vault. Initializing & Unsealing Vault servers can be a tedious job. 
Introduces to various methods of automatically `Initialization` & `Unsealing` Vault Servers.

- [Overview](/docs/concepts/vault-server-crds/unsealer/overview.md)
- [AWS KMS and SSM](/docs/concepts/vault-server-crds/unsealer/aws_kms_ssm.md)
- [Azure Key Vault](/docs/concepts/vault-server-crds/unsealer/azure_key_vault.md)
- [Google KMS GCS](/docs/concepts/vault-server-crds/unsealer/google_kms_gcs.md)
- [Kubernetes Secret](/docs/concepts/vault-server-crds/unsealer/kubernetes_secret.md)
  
### Vault Server Storage
The `storage backend` represents the location for the durable storage of Vault's information. Each backend has pros, cons, advantages, and trade-offs. For example, some backends support `High Availability - HA` while others provide a more robust backup and restoration process. Introduces to various `Storage Backend` options supported by `KubeVault`.

- [Overview](/docs/concepts/vault-server-crds/storage/overview.md)
- [Azure](/docs/concepts/vault-server-crds/storage/azure.md)
- [DynamoDB](/docs/concepts/vault-server-crds/storage/dynamodb.md)
- [Etcd](/docs/concepts/vault-server-crds/storage/etcd.md)
- [GCS](/docs/concepts/vault-server-crds/storage/gcs.md)
- [In Memory](/docs/concepts/vault-server-crds/storage/inmem.md)
- [MySQL](/docs/concepts/vault-server-crds/storage/mysql.md)
- [PosgreSQL](/docs/concepts/vault-server-crds/storage/postgresql.md)
- [AWS S3](/docs/concepts/vault-server-crds/storage/s3.md)
- [Swift](/docs/concepts/vault-server-crds/storage/swift.md)
- [Consul](/docs/concepts/vault-server-crds/storage/consul.md)
- [Raft](/docs/concepts/vault-server-crds/storage/raft.md)

### Authentication Methods for Vault Server
`Auth methods` are the components in Vault that perform authentication and are responsible for assigning identity and a set of policies to a user. In all cases, Vault will enforce authentication as part of the request processing. In most cases, Vault will delegate the authentication administration and decision to the relevant configured external auth method (e.g., Amazon Web Services, GitHub, Google Cloud Platform, Kubernetes, Microsoft Azure, Okta ...).

Having multiple auth methods enables you to use an auth method that makes the most sense for your use case of `Vault` and your organization.
Introduces to various `Authentication methods` supported by `KubeVault`.

- [Overview](/docs/concepts/vault-server-crds/auth-methods/overview.md)
- [AWS IAM Auth Method](/docs/concepts/vault-server-crds/auth-methods/aws-iam.md)
- [Kubernetes Auth Method](/docs/concepts/vault-server-crds/auth-methods/kubernetes.md)
- [TLS Certificates Auth Method](/docs/concepts/vault-server-crds/auth-methods/tls.md)
- [Token Auth Method](/docs/concepts/vault-server-crds/auth-methods/token.md)
- [Userpass Auth Method](/docs/concepts/vault-server-crds/auth-methods/userpass.md)
- [GCP IAM Auth Method](/docs/concepts/vault-server-crds/auth-methods/gcp-iam.md)
- [Azure Auth Method](/docs/concepts/vault-server-crds/auth-methods/azure.md)
- [JWT/OIDC Auth Method](/docs/concepts/vault-server-crds/auth-methods/jwt-oidc.md)

</div>
<div class="tab-pane fade" id="secret-engine" role="tabpanel" aria-labelledby="secret-engine-tab">

## Secret Engine

`SecretEngine` is a Kubernetes `Custom Resource Definition`(CRD). It provides a way to enable and configure a Vault secret engine. Introduces to `SecretEngine` CRD, fields, & it's various use cases.

- [Secret Engine](/docs/concepts/secret-engine-crds/secretengine.md)

### Secret Engine Roles
In a `Secret Engine`, a `role` describes an identity with a set of `permissions`, `groups`, or `policies` you want to attach a user of the Secret Engine. Introduces to various roles supported by `KubeVault`.

- [AWSRole](/docs/concepts/secret-engine-crds/aws-secret-engine/awsrole.md)
- [GCPRole](/docs/concepts/secret-engine-crds/gcp-secret-engine/gcprole.md)
- [AzureRole](/docs/concepts/secret-engine-crds/azure-secret-engine/azurerole.md)
- [MongoDBRole](/docs/concepts/secret-engine-crds/database-secret-engine/mongodb.md)
- [MySQLRole](/docs/concepts/secret-engine-crds/database-secret-engine/mysql.md)
- [PostgresRole](/docs/concepts/secret-engine-crds/database-secret-engine/postgresrole.md)
- [ElasticsearchRole](/docs/concepts/secret-engine-crds/database-secret-engine/elasticsearch.md)
- [MariaDBRole](/docs/concepts/secret-engine-crds/database-secret-engine/mariadb.md)
  
### Secret Access Request
A `SecretAccessRequest` is a `Kubernetes CustomResourceDefinition (CRD)` which allows a user to request a Vault server for `credentials` in a Kubernetes native way. A `SecretAccessRequest` can be created under various roleRef e.g: `AWSRole`, `GCPRole`, `ElasticsearchRole`, `MongoDBRole`, etc. Introduces to `SecretAccessRequest` CRD, fields & it's various use cases.

- [SecretAccessRequest](/docs/concepts/secret-engine-crds/secret-access-request.md)

### Secret Role Binding
A `SecretRoleBinding` is a `Kubernetes CustomResourceDefinition (CRD)` which allows a user to bind a set of `roles` to a set of `users`. Using the `SecretRoleBinding` it’s possible to bind various roles e.g: `AWSRole`, `GCPRole`, `ElasticsearchRole`, `MongoDBRole`, etc. to Kubernetes `ServiceAccounts`.

- [SecretRoleBinding](/docs/concepts/secret-engine-crds/secret-role-binding.md)

</div>
<div class="tab-pane fade" id="vault-policy" role="tabpanel" aria-labelledby="vault-policy-tab">

## Vault Policy

Everything in the Vault is path-based, and policies are no exception. Policies provide a declarative way to grant or forbid access to certain operations in Vault. Policies are `deny` by default, so an empty policy grants no permission in the system.

- [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md): is used to create, update or delete Vault policies.
- [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md): is used to create Vault auth roles associated with an authentication type/entity and a set of Vault policies.

</div>
</div>
