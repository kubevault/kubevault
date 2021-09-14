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

Introduces a way to specify connection information, credential, and parameters that are necessary for communicating with an app or service.

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)

## Vault Server Version

Introduces the concept of `VaultServerVersion` to specify the docker images of HashiCorp Vault, Unsealer, and Exporter.

- [VaultServerVersion](/docs/concepts/vault-server-crds/vaultserverversion.md)

## Vault Server

Introduces the concept of `VaultServer` for configuring a HashiCorp Vault server in a Kubernetes native way.

- [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md)

  - Vault Unsealer Options
    - [AWS KMS and SSM](/docs/concepts/vault-server-crds/unsealer/aws_kms_ssm.md)
    - [Azure Key Vault](/docs/concepts/vault-server-crds/unsealer/azure_key_vault.md)
    - [Google KMS GCS](/docs/concepts/vault-server-crds/unsealer/google_kms_gcs.md)
    - [Kubernetes Secret](/docs/concepts/vault-server-crds/unsealer/kubernetes_secret.md)

  - Vault Server Storage
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

  - Authentication Methods for Vault Server
    - [AWS IAM Auth Method](/docs/concepts/vault-server-crds/auth-methods/aws-iam.md)
    - [Kubernetes Auth Method](/docs/concepts/vault-server-crds/auth-methods/kubernetes.md)
    - [TLS Certificates Auth Method](/docs/concepts/vault-server-crds/auth-methods/tls.md)
    - [Token Auth Method](/docs/concepts/vault-server-crds/auth-methods/token.md)
    - [Userpass Auth Method](/docs/concepts/vault-server-crds/auth-methods/userpass.md)
    - [GCP IAM Auth Method](/docs/concepts/vault-server-crds/auth-methods/gcp-iam.md)
    - [Azure Auth Method](/docs/concepts/vault-server-crds/auth-methods/azure.md)

</div>
<div class="tab-pane fade" id="secret-engine" role="tabpanel" aria-labelledby="secret-engine-tab">

## Secret Engine

`SecretEngine` is a Kubernetes `Custom Resource Definition`(CRD). It provides a way to enable and configure a Vault secret engine.

- [Secret Engine](/docs/concepts/secret-engine-crds/secretengine.md)

  - AWS IAM Secret Engine
    - [AWSRole](/docs/concepts/secret-engine-crds/aws-secret-engine/awsrole.md)
    - [AWSAccessKeyRequest](/docs/concepts/secret-engine-crds/aws-secret-engine/awsaccesskeyrequest.md)

  - GCP Secret Engine
    - [GCPRole](/docs/concepts/secret-engine-crds/gcp-secret-engine/gcprole.md)
    - [GCPAccessKeyRequest](/docs/concepts/secret-engine-crds/gcp-secret-engine/gcpaccesskeyrequest.md)

  - Azure Secret Engine
    - [AzureRole](/docs/concepts/secret-engine-crds/azure-secret-engine/azurerole.md)
    - [AzureAccessKeyRequest](/docs/concepts/secret-engine-crds/azure-secret-engine/azureaccesskeyrequest.md)

  - Database Secret Engines
    - [MongoDBRole](/docs/concepts/secret-engine-crds/database-secret-engine/mongodb.md)
    - [MySQLRole](/docs/concepts/secret-engine-crds/database-secret-engine/mysql.md)
    - [PostgresRole](/docs/concepts/secret-engine-crds/database-secret-engine/postgresrole.md)
    - [ElasticsearchRole](/docs/concepts/secret-engine-crds/database-secret-engine/elasticsearch.md)
    - [DatabaseAccessRequest](/docs/concepts/secret-engine-crds/database-secret-engine/databaseaccessrequest.md)

</div>
<div class="tab-pane fade" id="vault-policy" role="tabpanel" aria-labelledby="vault-policy-tab">

## Vault Policy

Everything in the Vault is path-based, and policies are no exception. Policies provide a declarative way to grant or forbid access to certain operations in Vault. Policies are `deny` by default, so an empty policy grants no permission in the system.

- [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md): is used to create, update or delete Vault policies.
- [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md): is used to create Vault auth roles associated with an authentication type/entity and a set of Vault policies.

</div>
</div>
