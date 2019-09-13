---
title: Azure | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: azure-storage
    name: Azure
    parent: storage-vault-server-crds
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Azure

In Azure storage backend, data will be stored in [Azure Storage Container](https://azure.microsoft.com/en-us/services/storage/). Vault documentation for azure storage can be found in [here](https://www.vaultproject.io/docs/configuration/storage/azure.html).

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault-with-azure
  namespace: demo
spec:
  nodes: 1
  version: "0.11.1"
  backend:
    azure:
      accountName: "vault-ac"
      accountKeySecret: "azure-cred"
      container: "my-vault-storage"
```

## spec.backend.azure

To use Azure as backend storage in Vault specify `spec.backend.azure` in [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    azure:
      accountName: <storage_account_name>
      accountKeySecret: <storage_account_key_secret_name>
      container: <container_name>
      maxParallel: <max_parallel>
```

`spec.backend.azure` has following fields:

#### azure.accountName

`azure.accountName` is a required field that specifies the Azure Storage account name.

```yaml
spec:
  backend:
    azure:
      accountName: "my-vault-storage"
```

#### azure.accountKeySecret

`azure.accountKeySecret` is a required field that specifies the name of the secret containing Azure Storage account key. The secret contains following key:

- `account_key`

```yaml
spec:
  backend:
    azure:
      accountKeySecret: "azure-storage-key"
```

#### azure.container

`azure.container` is an required field that specifies the Azure Storage Blob container name.

```yaml
spec:
  backend:
    azure:
      container: "my-vault-storage"
```

#### azure.maxParallel

`maxParallel` is an optional field that specifies the maximum number of parallel operations to take place. This field accepts integer value. If this field is not specified, then Vault will set value `128`.

```yaml
spec:
  backend:
    azure:
      maxParallel: 124
```
