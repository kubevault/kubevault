---
title: What is KubeVault
menu:
  docs_{{ .version }}:
    identifier: what-is-kubevault-concepts
    name: Overview
    parent: concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Overview

## What is KubeVault

KubeVault operator is a Kubernetes controller for [HashiCorp Vault](https://www.vaultproject.io/). Vault is a tool for secrets management, encryption as a service, and privileged access management. Deploying, maintaining, and managing Vault in Kubernetes could be challenging. KubeVault operator eases these operational tasks so that developers can focus on solving business problems.

## Why use KubeVault

KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes. It covers automatic initialization and unsealing, and securely stores unseal keys and root tokens in a cloud KMS (Key Management Service) service. It provides the following features:

- Deploy TLS Secured [Vault Server](https://kubevault.com/docs/latest/concepts/vault-server-crds/vaultserver/)
- Manage Vault Server [TLS using Cert-manager](https://kubevault.com/docs/v2022.02.22/concepts/tls-encryption/overview/)
- Automate [Initialization & Unseal process of Vault Servers](https://kubevault.com/docs/v2022.02.22/concepts/vault-server-crds/unsealer/unsealer/)
- Add Durability to Vault's Data using [Storage Backend]()
- Enable & Configure [Secret Engines](https://kubevault.com/docs/v2022.02.22/concepts/secret-engine-crds/secretengine/)
- Create & Configure [Vault Roles](https://kubevault.com/docs/v2022.02.22/concepts/secret-engine-crds/gcp-secret-engine/gcprole/)
- Manage [Vault Policy](https://kubevault.com/docs/v2022.02.22/concepts/policy-crds/vaultpolicy/) & [Vault Policy Binding](https://kubevault.com/docs/v2022.02.22/concepts/policy-crds/vaultpolicybinding/)
- Manage user privileges using [SecretAccessRequest](/docs/concepts/secret-engine-crds/secret-access-request.md)
- Manage user privileges using [SecretRoleBinding](https://kubevault.com/docs/v2022.02.22/concepts/secret-engine-crds/secret-role-binding/)
- Inject Vault secrets into K8s resources
- Automate tedious operations using [KubeVault CLI](https://kubevault.com/docs/v2022.02.22/reference/cli/)
- Monitor Vault using Prometheus & Grafana Dashboard

## Core features

### Deploy TLS Secured Vault Server
A VaultServer is a Kubernetes CustomResourceDefinition (CRD) which is used to deploy a HashiCorp Vault server on Kubernetes clusters in a Kubernetes native way.

In production, Vault should always use TLS to provide secure communication between clients and the Vault server. You can deploy a TLS secure VaultServer using the KubeVault operator either with the self-signed certificate or with cert-manager to manage VaultServer TLS.

### Manage Vault Server TLS using Cert-manager
In production, Vault should always use TLS to provide secure communication between clients and the Vault server. KubeVault lets you use cert-manager to manage VaultServer TLS.

### Automate Initialization & Unseal process of Vault Servers

When a Vault server is started, it starts in a sealed state. In a sealed state, almost no operation is possible with a Vault server. So, you will need to unseal Vault. 

KubeVault operator provides automatic initialization and unsealing facility. When you deploy or scale up a Vault server, you don't have to worry about unsealing new Vault pods. The KubeVault operator will do it for you. Also, it provides various secure ways to store unseal keys and root token, e.g: Azure Key Vault, AWS KMS SSM, Google KMS GCS or Kubernetes Secret. 

### Enable & Configure Secret Engines
Secrets engines are components which store, generate, or encrypt data. Secrets engines are incredibly flexible, so it is easiest to think about them in terms of their function. Secrets engines are provided with some set of data, they take some action on that data, and they return a result.

KubeVault lets you enable & configure various Secret Engines e.g: AWS, Azure, Google Cloud KMS, MySQL, MariaDB, Elasticsearch, MongoDB, Postgresql, etc. in a Kubernetes native way.

### Create & Configure Vault Roles
In a Secret Engine, a role describes an identity with a set of permissions, groups, or policies you want to attach to a user of the Secret Engine.

KubeVault operator lets you create various roles e.g. AWSRole, AzureRole, GCPRole, MySQLRole, MariaDBRole, ElasticsearchRole, MongoDBRole, PostgresRole, etc. in a SecretEngine.

### Manage Vault Policy & Vault Policy Binding
Policies in Vault provide a declarative way to grant or forbid access to certain paths and operations in Vault. You can create, delete and update policy in Vault in a Kubernetes native way using KubeVault operator. KubeVault operator also provides a way to bind Vault policy with Kubernetes service accounts using the Vault Policy Binding. ServiceAccounts will have the permissions that are specified in the policy.

### Manage user privileges using SecretAccessRequest
A SecretAccessRequest is a Kubernetes CustomResourceDefinition (CRD) which allows a user to request a Vault server for credentials in a Kubernetes native way. A SecretAccessRequest can be created under various roles that can be enabled in a SecretEngine e.g: AWSRole, GCPRole, ElasticsearchRole, MongoDBRole, etc. This is a more human friendly way to manage DB privileges.

KubeVault operator lets you manage your DB user privileges with dynamic secrets rather than hard-coded credentials using SecretAccessRequest. This means that services that need to access a database no longer need to hardcode credentials: they can request them from Vault. Thus granting, revoking and monitoring user privileges is extremely easy with KubeVault.

### Manage user privileges using SecretRoleBinding
A SecretRoleBinding is a Kubernetes CustomResourceDefinition (CRD) which allows a user to bind a set of roles to a set of users. Using the SecretRoleBinding it’s possible to bind various roles e.g: AWSRole, GCPRole, ElasticsearchRole, MongoDBRole, etc. to Kubernetes ServiceAccounts. This way is more machine friendly and convenient for running your application with specific permissions.

Injecting Vault Secrets into Kubernetes resources requires specific permissions & using SecretRoleBinding it’s very easy to bind a set of policies to a set of Kubernetes Service Accounts.

### Inject Vault Secret into Kubernetes resources
Secrets Store CSI Driver for Kubernetes secrets - Integrates secrets stores with Kubernetes via a Container Storage Interface (CSI) volume. It allows Kubernetes to mount multiple secrets, keys, and certs stored in enterprise-grade external secrets stores into their pods as a volume. Once the Volume is attached, the data in it is mounted into the container’s file system.

KubeVault operator works seamlessly with Secrets Store CSI Driver. This is one of the recommended ways to mount Vault Secrets into Kubernetes resources along with Vault Agent Sidecar Injector.

Secrets Store CSI Driver requires a SecretProviderClass which is a namespaced resource that is used to provide driver configurations and provider-specific parameters to the CSI driver. Writing these SecretProviderClass can be a tedious job, but KubeVault CLI lets you generate SecretProviderClass using simple CLI commands.

### Automate tedious operations using KubeVault CLI
KubeVault CLI is an integral part of the KubeVault operator. It makes various tasks simple while working with the operator e.g. Approve/Deny/Revoke SecretAccessRequest, Generate SecretProviderClass, Get, Set, List, Sync Vault Unseal Keys and Vault Root Token, etc.

### Monitor Vault using Prometheus & Grafana Dashboard
You can monitor Vault servers using the Vault dashboard.


KubeVault operator has native support for monitoring via [Prometheus](https://prometheus.io/). You can use builtin [Prometheus](https://github.com/prometheus/prometheus) scraper or [Prometheus Operator](https://github.com/coreos/prometheus-operator) to monitor KubeVault operator.
