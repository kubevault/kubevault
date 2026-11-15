---
title: Vault Server Overview
menu:
  docs_{{ .version }}:
    identifier: overview-vault-server
    name: Overview
    parent: vault-server-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Overview

The KubeVault operator makes it easy to deploy, maintain and manage Vault servers in Kubernetes clusters. It covers automatic initialization and unsealing and also stores unseal keys and root token in a secure way. The KubeVault operator can manage policies and secret engines of Vault servers which are not provisioned by the KubeVault operator. It has the following features:

- **Vault Policy Management**: Provides a Kubernetes native way to manage Vault policies and bind those policies to the users or the auth method roles.

  - [Vault Policy](/docs/guides/policy-management/overview.md#vaultpolicy)
  - [Vault Policy Binding](/docs/guides/policy-management/overview.md#vaultpolicybinding)

- **Vault Secret Engine Management**: Provides a Kubernetes native way to manage Vault secret engines.

  - [AWS Secret Engine](/docs/guides/secret-engines/aws/overview.md)
  - [Azure Secret Engine](/docs/guides/secret-engines/azure/overview.md)
  - Database Secret Engine
    - [Druid Secret Engine](/docs/guides/secret-engines/database/druid/overview.md)
    - [Elasticsearch Secret Engine](/docs/guides/secret-engines/database/elasticsearch/overview.md)
    - [etcd Secret Engine](/docs/guides/secret-engines/database/etcd/overview.md)
    - [Ignite Secret Engine](/docs/guides/secret-engines/database/ignite/overview.md)
    - [Kafka Secret Engine](/docs/guides/secret-engines/database/kafka/overview.md)
    - [MariaDB Secret Engine](/docs/guides/secret-engines/database/mariadb/overview.md)
    - [Microsoft SQL Server Secret Engine](/docs/guides/secret-engines/database/mssqlserver/overview.md)
    - [Milvus Secret Engine](/docs/guides/secret-engines/database/milvus/overview.md)
    - [MongoDB Secret Engine](/docs/guides/secret-engines/database/mongodb/overview.md)
    - [MySQL Secret Engine](/docs/guides/secret-engines/database/mysql/overview.md)
    - [Neo4j Secret Engine](/docs/guides/secret-engines/database/neo4j/overview.md)
    - [Oracle Secret Engine](/docs/guides/secret-engines/database/oracle/overview.md)
    - [PostgreSQL Secret Engine](/docs/guides/secret-engines/database/postgres/overview.md)
    - [Qdrant Secret Engine](/docs/guides/secret-engines/database/qdrant/overview.md)
    - [RabbitMQ Secret Engine](/docs/guides/secret-engines/database/rabbitmq/overview.md)
    - [Redis Secret Engine](/docs/guides/secret-engines/database/redis/overview.md)
    - [SAP HanaDB Secret Engine](/docs/guides/secret-engines/database/hanadb/overview.md)
    - [Solr Secret Engine](/docs/guides/secret-engines/database/solr/overview.md)
    - [Weaviate Secret Engine](/docs/guides/secret-engines/database/weaviate/overview.md)
  - [GCP Secret Engine](/docs/guides/secret-engines/gcp/overview.md)
  - [Key/Value Secret Engine](/docs/guides/secret-engines/kv/overview.md)
  - [PKI Secret Engine](/docs/guides/secret-engines/pki/overview.md)

## Setup Vault Server

![Overview](/docs/images/guides/vault-server/overview_vault_server_guide.svg)

Deploy Vault server using the KubeVault operator:

- [Deploy Vault Server](/docs/guides/vault-server/vault-server.md)
- [Enable Vault CLI](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

 Configure external Vault server so that the  KubeVault operator can communicate with it:

- [Configure Cluster and External Vault Server](/docs/guides/vault-server/external-vault-sever.md)
