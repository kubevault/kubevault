---
title: Overview
menu:
  docs_{{ .version }}:
    identifier: storage-overview
    name: Overview
    parent: storage-vault-server-crds
    weight: 1
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Storage Backend

## Configuring Storage Backend

```yaml
spec:
  backend:
    <backend-type>:
      ...
```

Here, we are going to describe the various attributes of the `spec.backend` field.

List of supported modes:

- [Aerospike](/docs/concepts/vault-server-crds/storage/aerospike.md)
- [Alicloud OSS](/docs/concepts/vault-server-crds/storage/alicloudoss.md)
- [Azure](/docs/concepts/vault-server-crds/storage/azure.md)
- [Cassandra](/docs/concepts/vault-server-crds/storage/cassandra.md)
- [CockroachDB](/docs/concepts/vault-server-crds/storage/cockroachdb.md)
- [Consul](/docs/concepts/vault-server-crds/storage/consul.md)
- [CouchDB](/docs/concepts/vault-server-crds/storage/couchdb.md)
- [DynamoDB](/docs/concepts/vault-server-crds/storage/dynamodb.md)
- [Etcd](/docs/concepts/vault-server-crds/storage/etcd.md)
- [Filesystem](/docs/concepts/vault-server-crds/storage/filesystem.md)
- [GCS](/docs/concepts/vault-server-crds/storage/gcs.md)
- [Google Cloud Spanner](/docs/concepts/vault-server-crds/storage/spanner.md)
- [Inmem](/docs/concepts/vault-server-crds/storage/inmem.md)
- [MSSQL](/docs/concepts/vault-server-crds/storage/mssql.md)
- [MySQL](/docs/concepts/vault-server-crds/storage/mysql.md)
- [OCI Object Storage](/docs/concepts/vault-server-crds/storage/oci.md)
- [PostgreSQL](/docs/concepts/vault-server-crds/storage/postgresql.md)
- [Raft](/docs/concepts/vault-server-crds/storage/raft.md)
- [S3](/docs/concepts/vault-server-crds/storage/s3.md)
- [Swift](/docs/concepts/vault-server-crds/storage/swift.md)
- [ZooKeeper](/docs/concepts/vault-server-crds/storage/zookeeper.md)
