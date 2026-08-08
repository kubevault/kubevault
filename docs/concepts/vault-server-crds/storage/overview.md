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

- [Filesystem](/docs/concepts/vault-server-crds/storage/filesystem.md)
- [Inmem](/docs/concepts/vault-server-crds/storage/inmem.md)
- [PostgreSQL](/docs/concepts/vault-server-crds/storage/postgresql.md)
- [Raft](/docs/concepts/vault-server-crds/storage/raft.md)

> Only the storage backends supported by [OpenBao](https://openbao.org/docs/configuration/storage/)
> are documented here. Legacy HashiCorp Vault-only backends (Azure, Consul, DynamoDB, Etcd,
> GCS, MySQL, S3, Swift) have been removed.
