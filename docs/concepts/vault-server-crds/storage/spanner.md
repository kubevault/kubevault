---
title: Google Cloud Spanner | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: spanner-storage
    name: Google Cloud Spanner
    parent: storage-vault-server-crds
    weight: 80
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Google Cloud Spanner

In the Spanner storage backend, Vault data will be stored in [Google Cloud Spanner](https://cloud.google.com/spanner). This backend was removed from upstream OpenBao and is only available in this fork; Vault documentation for the Spanner storage backend can be found [here](https://developer.hashicorp.com/vault/docs/configuration/storage/google-cloud-spanner).

> `spec.backend.spanner` is only available on the `kubevault.com/v1alpha2` `VaultServer` API.

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault-with-spanner
  namespace: demo
spec:
  version: "2.0.0"
  replicas: 1
  backend:
    spanner:
      database: "projects/my-project/instances/my-instance/databases/vault"
      credentialSecretRef:
        name: spanner-credential
```

## spec.backend.spanner

To use Spanner as backend storage in Vault, specify `spec.backend.spanner` in the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    spanner:
      database: <spanner_database_path>
      table: <table_name>
      haEnabled: <true/false>
      haTable: <ha_table_name>
      maxParallel: <max_parallel>
      credentialSecretRef:
        name: <secret_name>
```

Here, we are going to describe the various attributes of the `spec.backend.spanner` field.

### spanner.database

`spanner.database` is a required field that specifies the full name of the Spanner database, in the form `projects/<project>/instances/<instance>/databases/<database>`.

```yaml
spec:
  backend:
    spanner:
      database: "projects/my-project/instances/my-instance/databases/vault"
```

### spanner.table

`spanner.table` is an optional field that specifies the name of the table to use for Vault data. Defaults to `Vault`.

```yaml
spec:
  backend:
    spanner:
      table: "Vault"
```

### spanner.haEnabled

`spanner.haEnabled` is an optional field that specifies whether this backend should be used to run Vault in high availability mode. This field accepts a boolean value (as a string). The default value is `false`.

```yaml
spec:
  backend:
    spanner:
      haEnabled: "true"
```

### spanner.haTable

`spanner.haTable` is an optional field that specifies the name of the table used for HA leader election. Defaults to the value of `table` suffixed with `HA`.

```yaml
spec:
  backend:
    spanner:
      haTable: "VaultHA"
```

### spanner.maxParallel

`spanner.maxParallel` is an optional field that specifies the maximum number of concurrent requests to Spanner.

```yaml
spec:
  backend:
    spanner:
      maxParallel: 128
```

### spanner.credentialSecretRef

`spanner.credentialSecretRef` is an optional field that specifies the name of the `Secret`, in the same namespace as the `VaultServer`, containing the Google application credential used to reach Spanner. The secret contains the following key:

- `sa.json`

```yaml
spec:
  backend:
    spanner:
      credentialSecretRef:
        name: spanner-credential
```
