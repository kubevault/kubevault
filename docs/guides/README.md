---
title: Table of Contents | Guides
description: Table of Contents | Guides
menu:
  docs_{{ .version }}:
    identifier: guides-readme
    name: Readme
    parent: guides
    weight: -1
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/
aliases:
  - /docs/{{ .version }}/guides/README/
---

# Guides

Guides show you how to perform tasks with KubeVault operator and Vault CSI driver.

## Using KubeVault operator

- To deploy Vault with a specific storage backend (Aerospike, Amazon S3, Azure Blob Storage, Cassandra, CockroachDB, Consul, etcd, Google Cloud Storage, MySQL, PostgreSQL, Raft and more) see [here](/docs/guides/storage-backends/overview.md).
- To deploy Vault in AWS EKS see [here](/docs/guides/platforms/eks.md).
- To deploy Vault in Google GKE see [here](/docs/guides/platforms/gke.md).
- To deploy Vault in Azure AKS see [here](/docs/guides/platforms/aks.md).
- To manage Vault policy see [here](/docs/guides/policy-management/overview.md).
- To manage Vault AWS secret engine see [here](/docs/guides/secret-engines/aws/overview.md).
- To manage Vault Azure secret engine see [here](/docs/guides/secret-engines/azure/overview.md).
- To manage Vault Druid Database secret engine see [here](/docs/guides/secret-engines/database/druid/overview.md).
- To manage Vault Elasticsearch Database secret engine see [here](/docs/guides/secret-engines/database/elasticsearch/overview.md).
- To manage Vault etcd Database secret engine see [here](/docs/guides/secret-engines/database/etcd/overview.md).
- To manage Vault GCP secret engine see [here](/docs/guides/secret-engines/gcp/overview.md).
- To manage Vault Ignite Database secret engine see [here](/docs/guides/secret-engines/database/ignite/overview.md).
- To manage Vault Kafka Database secret engine see [here](/docs/guides/secret-engines/database/kafka/overview.md).
- To manage Vault Key/Value secret engine see [here](/docs/guides/secret-engines/kv/overview.md).
- To manage Vault MariaDB Database secret engine see [here](/docs/guides/secret-engines/database/mariadb/overview.md).
- To manage Vault Microsoft SQL Server Database secret engine see [here](/docs/guides/secret-engines/database/mssqlserver/overview.md).
- To manage Vault Milvus Database secret engine see [here](/docs/guides/secret-engines/database/milvus/overview.md).
- To manage Vault MongoDB Database secret engine see [here](/docs/guides/secret-engines/database/mongodb/overview.md).
- To manage Vault MySQL Database secret engine see [here](/docs/guides/secret-engines/database/mysql/overview.md).
- To manage Vault Neo4j Database secret engine see [here](/docs/guides/secret-engines/database/neo4j/overview.md).
- To manage Vault Oracle Database secret engine see [here](/docs/guides/secret-engines/database/oracle/overview.md).
- To manage Vault PKI secret engine see [here](/docs/guides/secret-engines/pki/overview.md).
- To manage Vault PostgreSQL Database secret engine see [here](/docs/guides/secret-engines/database/postgres/overview.md).
- To manage Vault Qdrant Database secret engine see [here](/docs/guides/secret-engines/database/qdrant/overview.md).
- To manage Vault RabbitMQ Database secret engine see [here](/docs/guides/secret-engines/database/rabbitmq/overview.md).
- To manage Vault Redis Database secret engine see [here](/docs/guides/secret-engines/database/redis/overview.md).
- To manage Vault SAP HanaDB Database secret engine see [here](/docs/guides/secret-engines/database/hanadb/overview.md).
- To manage Vault Solr Database secret engine see [here](/docs/guides/secret-engines/database/solr/overview.md).
- To manage Vault Weaviate Database secret engine see [here](/docs/guides/secret-engines/database/weaviate/overview.md).
- To set up monitoring see [here](/docs/guides/monitoring/overview.md).
- To use KubeVault operator for external Vault see [here](/docs/guides/platforms/external-vault.md).
- To use KubeVault operator in multiple cluster for same Vault see [here](/docs/guides/platforms/multi-cluster-vault.md).
- To deploy Vault in a hub-spoke model across OCM managed clusters see [here](/docs/guides/hub-spoke/deploy-hub-spoke.md).

## Using Vault CSI driver

- To mount AWS secret engine secret in pod see [here](/docs/guides/secret-engines/aws/csi-driver.md).
- To mount Azure secret engine secret in pod see [here](/docs/guides/secret-engines/azure/csi-driver.md).
- To mount Druid Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/druid/csi-driver.md).
- To mount Elasticsearch Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/elasticsearch/csi-driver.md).
- To mount etcd Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/etcd/csi-driver.md).
- To mount GCP secret engine secret in pod see [here](/docs/guides/secret-engines/gcp/csi-driver.md).
- To mount Ignite Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/ignite/csi-driver.md).
- To mount Kafka Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/kafka/csi-driver.md).
- To mount Key/Value secret engine secret in pod see [here](/docs/guides/secret-engines/kv/csi-driver.md).
- To mount MariaDB Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/mariadb/csi-driver.md).
- To mount Microsoft SQL Server Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/mssqlserver/csi-driver.md).
- To mount Milvus Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/milvus/csi-driver.md).
- To mount MongoDB Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/mongodb/csi-driver.md).
- To mount MySQL Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/mysql/csi-driver.md).
- To mount Neo4j Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/neo4j/csi-driver.md).
- To mount Oracle Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/oracle/csi-driver.md).
- To mount PKI secret engine secret in pod see [here](/docs/guides/secret-engines/pki/csi-driver.md).
- To mount PostgreSQL Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/postgres/csi-driver.md).
- To mount Qdrant Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/qdrant/csi-driver.md).
- To mount RabbitMQ Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/rabbitmq/csi-driver.md).
- To mount Redis Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/redis/csi-driver.md).
- To mount SAP HanaDB Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/hanadb/csi-driver.md).
- To mount Solr Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/solr/csi-driver.md).
- To mount Weaviate Database secret engine secret in pod see [here](/docs/guides/secret-engines/database/weaviate/csi-driver.md).
- To set up monitoring see [here](/docs/guides/monitoring/overview.md).
