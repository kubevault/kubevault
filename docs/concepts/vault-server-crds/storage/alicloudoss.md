---
title: Alicloud OSS | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: alicloudoss-storage
    name: Alicloud OSS
    parent: storage-vault-server-crds
    weight: 95
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Alicloud OSS

In the Alicloud OSS storage backend, Vault data will be stored in an [Alibaba Cloud Object Storage Service](https://www.alibabacloud.com/product/object-storage-service) bucket. This backend was removed from upstream OpenBao and is only available in this fork; Vault documentation for the Alicloud OSS storage backend can be found [here](https://developer.hashicorp.com/vault/docs/configuration/storage/alicloudoss).

> `spec.backend.alicloudoss` is only available on the `kubevault.com/v1alpha2` `VaultServer` API. Alicloud OSS has no `haEnabled` field and cannot be used to run Vault in high availability mode.

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault-with-alicloudoss
  namespace: demo
spec:
  version: "2.0.0"
  replicas: 1
  backend:
    alicloudoss:
      endpoint: "http://oss-cn-hangzhou.aliyuncs.com"
      bucket: "vault-bucket"
      credentialSecretRef:
        name: alicloud-cred
```

## spec.backend.alicloudoss

To use Alicloud OSS as backend storage in Vault, specify `spec.backend.alicloudoss` in the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    alicloudoss:
      endpoint: <oss_endpoint>
      bucket: <bucket_name>
      credentialSecretRef:
        name: <secret_name>
      maxParallel: <max_parallel>
```

Here, we are going to describe the various attributes of the `spec.backend.alicloudoss` field.

### alicloudoss.endpoint

`alicloudoss.endpoint` is a required field that specifies the OSS endpoint to connect to, e.g. `http://oss-cn-hangzhou.aliyuncs.com`.

```yaml
spec:
  backend:
    alicloudoss:
      endpoint: "http://oss-cn-hangzhou.aliyuncs.com"
```

### alicloudoss.bucket

`alicloudoss.bucket` is a required field that specifies the name of the OSS bucket to store data in. It must already exist; Vault does not create it.

```yaml
spec:
  backend:
    alicloudoss:
      bucket: "vault-bucket"
```

### alicloudoss.credentialSecretRef

`alicloudoss.credentialSecretRef` is an optional field that specifies the name of the `Secret`, in the same namespace as the `VaultServer`, containing the Alicloud access key ID and secret to connect with. The secret contains the following keys:

- `access_key`
- `secret_key`

```yaml
spec:
  backend:
    alicloudoss:
      credentialSecretRef:
        name: alicloud-cred
```

### alicloudoss.maxParallel

`alicloudoss.maxParallel` is an optional field that specifies the maximum number of concurrent requests to OSS.

```yaml
spec:
  backend:
    alicloudoss:
      maxParallel: 128
```
