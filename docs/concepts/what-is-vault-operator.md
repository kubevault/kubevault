# What is Vault operator

Vault operator is a Kubernetes controller for [HashiCorp Vault](https://www.vaultproject.io/). Vault is a tool for secrets management, encryption as a service, and privileged access management. Deploying, maintaining, and managing Vault in Kubernetes could be challenging. Vault operator eases these operational tasks so that developers can focus on solving business problems.

# Why use Vault operator

Vault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes. It covers automatic initialization and unsealing and also stores unseal keys and root token in a secure way using cloud KMS (Key Management Service) service. It provides the following features:

- Automatic initializing and unsealing Vault
- Manage Vault [Policy](https://www.vaultproject.io/docs/concepts/policies.html)
- Manage Vault [AWS secret engine](https://www.vaultproject.io/docs/secrets/aws/index.html#aws-secrets-engine)
- Manage Vault [MongoDB Database secret engine](https://www.vaultproject.io/api/secret/databases/mongodb.html)
- Manage Vault [MySQL Database secret engine](https://www.vaultproject.io/api/secret/databases/mysql-maria.html)
- Manage Vault [PostgreSQL Database secret engine](https://www.vaultproject.io/api/secret/databases/postgresql.html)
- Monitor Vault using Prometheus

# Core features

## Automatic Initialization & Unsealing of Vault Servers

When a Vault server is started, it starts in a sealed state. In a sealed state, almost no operation is possible with a Vault server. So, you will need to unseal Vault. Vault operator provides automatic initialization and unsealing facility. When you deploy or scale up a Vault server, you don't have worry about unsealing new Vault pods. Vault operator will do it for you. Also, it provides a secure way to store unseal keys and root token.

## Manage Vault Policy

Policies in Vault provide a declarative way to grant or forbid access to certain paths and operations in Vault. You can create, delete and update policy in Vault in a Kubernetes native way using Vault operator. Vault operator also provides a way to bind Vault policy with Kubernetes service accounts. ServiceAccounts will have the permissions that are specified in the policy.

## Manage Vault AWS Secret Engine

AWS secret engine in Vault generates AWS access credentials dynamically based on IAM policies. This makes AWS IAM user management easier. Using Vault operator, you can configure AWS secret engine and issue AWS access credential via Vault. A User can request AWS credential and after it's been approved Vault operator will create a Kubernetes Secret containing the AWS credential and also creates RBAC Role and RoleBinding so that the user can access the secret.

# Manage Vault MongoDB Database Secret Engine

MongoDB database secret engine in Vault generates MongoDB database credentials dynamically based on configured roles. Using Vault operator, you can configure secret engine, create role and issue credential from Vault. A User can request credential and after it's been approved Vault operator will create a Kubernetes Secret containing the credential and also creates RBAC Role and RoleBinding so that the user can access the Secret.

# Manage Vault MySQL Database Secret Engine

MySQL database secret engine in Vault generates MySQL database credentials dynamically based on configured roles. Using Vault operator, you can configure secret engine, create role and issue credential from Vault. A User can request credential and after it's been approved Vault operator will create a Kubernetes Secret containing the credential and also creates RBAC Role and RoleBinding so that the user can access the Secret.

# Manage Vault Postgres Database Secret Engine

Postgres database secret engine in Vault generates Postgres database credentials dynamically based on configured roles. Using Vault operator, you can configure secret engine, create role and issue credential from Vault. A User can request credential and after it's been approved Vault operator will create a Kubernetes Secret containing the credential and also creates RBAC Role and RoleBinding so that the user can access the Secret.

# Monitor Vault using Prometheus

Vault operator has native support for monitoring via [Prometheus](https://prometheus.io/). You can use builtin [Prometheus](https://github.com/prometheus/prometheus) scrapper or [CoreOS Prometheus Operator](https://github.com/coreos/prometheus-operator) to monitor Vault operator.

# Architecture

Vault operator is composed of following controllers:

- A **Vault Server controller** that deploys Vault in Kubernetes cluster. It also injects unsealer and stastd export as sidecar to perform unsealing and monitoring.

- A **Auth controller** that enables auth method in Vault.

- A **Policy controller** that manages Vault policy and also bind the policy with Kubernetes ServiceAccount.

- A **AWS secret engine controller** that configures secret engine, manages role and credential.

- A **Database secret engine controller** that configures database secret engine, manages role and credential.

![vault operator architecture](/docs/images/concepts/vault_operator_architecture.svg)