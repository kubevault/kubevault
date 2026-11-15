---
title: Storage Backends Overview
menu:
  docs_{{ .version }}:
    identifier: storage-backends-overview
    name: Overview
    parent: storage-backends-guides
    weight: 1
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Storage Backends

A `VaultServer` keeps the data that Vault encrypts and stores in the storage backend set under `spec.backend`. Exactly one backend can be configured. The guides in this section deploy a `VaultServer` with each of the supported backends, in the same way as the [Raft](/docs/guides/storage-backends/raft.md) and [PostgreSQL](/docs/guides/storage-backends/postgresql.md) guides: prepare the storage, create the credentials, deploy the `VaultServer`, then write a secret to see that it works.

To learn what each field of a backend does, read the [storage concepts](/docs/concepts/vault-server-crds/storage/overview.md).

## Supported Storage Backends

| Backend | High availability | What you need |
| --- | --- | --- |
| [Aerospike](/docs/guides/storage-backends/aerospike.md) | No | An Aerospike server with an existing namespace |
| [Alibaba Cloud OSS](/docs/guides/storage-backends/alicloudoss.md) | No | An OSS bucket and an access key pair |
| [Amazon DynamoDB](/docs/guides/storage-backends/dynamodb.md) | Yes (`haEnabled`) | AWS credentials (the table is created by Vault) |
| [Amazon S3](/docs/guides/storage-backends/s3.md) | No | An S3 bucket and an access key pair |
| [Azure Blob Storage](/docs/guides/storage-backends/azure.md) | No | An Azure storage account and its access key |
| [Cassandra](/docs/guides/storage-backends/cassandra.md) | No | Cassandra with a keyspace and a table (KubeDB `Cassandra`) |
| [CockroachDB](/docs/guides/storage-backends/cockroachdb.md) | Yes (`haEnabled`) | A CockroachDB cluster and a user |
| [Consul](/docs/guides/storage-backends/consul.md) | Yes | A Consul server |
| [CouchDB](/docs/guides/storage-backends/couchdb.md) | No | A CouchDB server with a database |
| [etcd](/docs/guides/storage-backends/etcd.md) | Yes (`haEnabled`) | An etcd v3 cluster (KubeDB `Etcd`) |
| [Filesystem](/docs/guides/storage-backends/filesystem.md) | No | A `PersistentVolumeClaim` (created from `volumeClaimTemplate`) |
| [Google Cloud Spanner](/docs/guides/storage-backends/spanner.md) | Yes (`haEnabled`) | A Spanner database with the `Vault` tables and a service account key |
| [Google Cloud Storage](/docs/guides/storage-backends/gcs.md) | Yes (`haEnabled`) | A GCS bucket and a service account key |
| [In-Memory](/docs/guides/storage-backends/inmem.md) | No | Nothing. Data is lost on restart |
| [Microsoft SQL Server](/docs/guides/storage-backends/mssql.md) | No | A SQL Server (KubeDB `MSSQLServer`) |
| [MySQL](/docs/guides/storage-backends/mysql.md) | Yes (`haEnabled`) | A MySQL server (KubeDB `MySQL`) |
| [OCI Object Storage](/docs/guides/storage-backends/oci.md) | Yes (`haEnabled`) | Two OCI buckets and an API signing key |
| [OpenStack Swift](/docs/guides/storage-backends/swift.md) | No | A Swift container and OpenStack credentials |
| [PostgreSQL](/docs/guides/storage-backends/postgresql.md) | Yes (`haEnabled`) | A PostgreSQL server (KubeDB `Postgres`) |
| [Raft](/docs/guides/storage-backends/raft.md) | Yes | A `StorageClass` for the integrated storage volumes |
| [ZooKeeper](/docs/guides/storage-backends/zookeeper.md) | Yes | A ZooKeeper ensemble (KubeDB `ZooKeeper`) |

## Things to Know

- **Namespace and names.** Every guide uses the `demo` namespace and a `VaultServer` named `vault`. Follow the `Cleanup` section of a guide before you start another one in the same namespace.
- **High availability.** Backends without high availability have no leader election, so run them with a single replica. Backends marked `haEnabled` run in HA mode only when that field is set. Raft, Consul and ZooKeeper are always HA-capable.
- **Credentials.** Credentials are read from a `Secret` in the namespace of the `VaultServer`, referenced by `credentialSecretRef`. For the databases that KubeDB manages, `databaseRef` points to the `AppBinding` of the database instead, and the operator takes the address, the credentials and the TLS settings from it.
- **Cleanup order.** With `terminationPolicy: WipeOut`, the operator connects to the storage backend to delete the data Vault stored there. Always delete the `VaultServer` before the storage backend it uses.
