---
title: CouchDB | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: couchdb-storage
    name: CouchDB
    parent: storage-vault-server-crds
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# CouchDB

In the CouchDB storage backend, Vault data will be stored in [Apache CouchDB](https://couchdb.apache.org/). This backend was removed from upstream OpenBao and is only available in this fork; Vault documentation for the CouchDB storage backend can be found [here](https://developer.hashicorp.com/vault/docs/configuration/storage/couchdb).

> `spec.backend.couchdb` is only available on the `kubevault.com/v1alpha2` `VaultServer` API. CouchDB has no `haEnabled` field and cannot be used to run Vault in high availability mode.

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault-with-couchdb
  namespace: demo
spec:
  version: "2.0.0"
  replicas: 1
  backend:
    couchdb:
      endpoint: "https://couchdb.demo.svc:5984/vault"
      credentialSecretRef:
        name: couchdb-cred
```

## spec.backend.couchdb

To use CouchDB as backend storage in Vault, specify `spec.backend.couchdb` in the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    couchdb:
      endpoint: <database_url>
      credentialSecretRef:
        name: <secret_name>
      maxParallel: <max_parallel>
```

Here, we are going to describe the various attributes of the `spec.backend.couchdb` field.

### couchdb.endpoint

`couchdb.endpoint` is a required field that specifies the full URL to the CouchDB database to use, including the database name. The database must already exist; Vault does not create it.

```yaml
spec:
  backend:
    couchdb:
      endpoint: "https://couchdb.demo.svc:5984/vault"
```

### couchdb.credentialSecretRef

`couchdb.credentialSecretRef` is an optional field that specifies the name of the `Secret`, in the same namespace as the `VaultServer`, containing the CouchDB username and password to connect with. The secret contains the following keys:

- `username`
- `password`

```yaml
spec:
  backend:
    couchdb:
      credentialSecretRef:
        name: couchdb-cred
```

### couchdb.maxParallel

`couchdb.maxParallel` is an optional field that specifies the maximum number of concurrent requests to CouchDB.

```yaml
spec:
  backend:
    couchdb:
      maxParallel: 128
```
