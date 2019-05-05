---
title: KubeVault Concepts
menu:
  docs_0.2.0:
    identifier: concepts-readme
    name: Overview
    parent: concepts
    weight: 10
menu_name: docs_0.2.0
section_menu_id: concepts
url: /docs/0.2.0/concepts/
aliases:
  - /docs/0.2.0/concepts/README/
---

# Concepts

Concepts help you learn about the different parts of the KubeVault and the abstractions it uses.

<ul class="nav nav-tabs" id="conceptsTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="operator-tab" data-toggle="tab" href="#operator" role="tab" aria-controls="operator" aria-selected="true">Vault Operator</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="csi-driver-tab" data-toggle="tab" href="#csi-driver" role="tab" aria-controls="csi-driver" aria-selected="false">Secret Engines</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="policy-mgr-tab" data-toggle="tab" href="#policy-mgr" role="tab" aria-controls="policy-mgr" aria-selected="false">Policy Management</a>
  </li>
</ul>
<div class="tab-content" id="conceptsTabContent">
  <div class="tab-pane fade show active" id="operator" role="tabpanel" aria-labelledby="operator-tab">

- What is KubeVault?
  - [Overview](/docs/concepts/what-is-kubevault.md). Provides a conceptual introduction to KubeVault operator, including the problems it solves and its high-level architecture.
- Custom Resource Definitions
  - [Vault Server](/docs/concepts/vault-server-crds/vaultserver.md). Introduces the concept of `VaultServer` for configuring a HashiCorp Vault server in a Kubernetes native way.
  - [Vault Server Version](/docs/concepts/vault-server-crds/vaultserverversion.md). Introduces the concept of `VaultServerVersion` to specify the docker images of HashiCorp Vault, Unsealer and Exporter.
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
- Authentication Methods for Vault Server
  - [AWS IAM Auth Method](/docs/concepts/vault-server-crds/auth-methods/aws-iam.md)
  - [Kubernetes Auth Method](/docs/concepts/vault-server-crds/auth-methods/kubernetes.md)
  - [TLS Certificates Auth Method](/docs/concepts/vault-server-crds/auth-methods/tls.md)
  - [Token Auth Method](/docs/concepts/vault-server-crds/auth-methods/token.md)
  - [Userpass Auth Method](/docs/concepts/vault-server-crds/auth-methods/userpass.md)

</div>
<div class="tab-pane fade" id="csi-driver" role="tabpanel" aria-labelledby="csi-driver-tab">

- AWS IAM Secret Engines
  - [AWSRole](/docs/concepts/secret-engine-crds/awsrole.md)
  - [AWSAccessKeyRequest](/docs/concepts/secret-engine-crds/awsaccesskeyrequest.md)
- Database Secret Engines
  - [MongoDBRole](/docs/concepts/database-crds/mongodb.md)
  - [MySQLRole](/docs/concepts/database-crds/mysql.md)
  - [PostgresRole](/docs/concepts/database-crds/postgresrole.md)
  - [DatabaseAccessRequest](/docs/concepts/database-crds/databaseaccessrequest.md)

</div>
<div class="tab-pane fade" id="policy-mgr" role="tabpanel" aria-labelledby="policy-mgr-tab">

- Vault Policy Management
  - [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md)
  - [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md)

</div>
</div>
