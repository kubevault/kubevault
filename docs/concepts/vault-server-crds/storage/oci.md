---
title: OCI Object Storage | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: oci-storage
    name: OCI Object Storage
    parent: storage-vault-server-crds
    weight: 170
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# OCI Object Storage

In the OCI storage backend, Vault data will be stored in an [Oracle Cloud Infrastructure Object Storage](https://www.oracle.com/cloud/storage/object-storage/) bucket. Vault documentation for the OCI storage backend can be found [here](https://developer.hashicorp.com/vault/docs/configuration/storage/oci-object-storage).

> `spec.backend.oci` is only available on the `kubevault.com/v1alpha2` `VaultServer` API.

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault-with-oci
  namespace: demo
spec:
  version: "2.0.0"
  replicas: 1
  backend:
    oci:
      bucketName: "vault-bucket"
      namespaceName: "my-tenancy-namespace"
      region: "us-phoenix-1"
      authTypeAPIKey: true
      credentialSecretRef:
        name: oci-cred
```

## spec.backend.oci

To use OCI Object Storage as backend storage in Vault, specify `spec.backend.oci` in the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    oci:
      bucketName: <bucket_name>
      namespaceName: <object_storage_namespace>
      region: <region>
      authTypeAPIKey: <true/false>
      credentialSecretRef:
        name: <secret_name>
      haEnabled: <true/false>
      lockBucketName: <lock_bucket_name>
```

Here, we are going to describe the various attributes of the `spec.backend.oci` field.

### oci.bucketName

`oci.bucketName` is a required field that specifies the name of the OCI Object Storage bucket to store data in. It must already exist; Vault does not create it.

```yaml
spec:
  backend:
    oci:
      bucketName: "vault-bucket"
```

### oci.namespaceName

`oci.namespaceName` is a required field that specifies the Object Storage namespace the bucket belongs to.

```yaml
spec:
  backend:
    oci:
      namespaceName: "my-tenancy-namespace"
```

### oci.region

`oci.region` is an optional field that specifies the OCI region to use. Defaults to the region configured in the resolved OCI configuration provider.

```yaml
spec:
  backend:
    oci:
      region: "us-phoenix-1"
```

### oci.authTypeAPIKey

`oci.authTypeAPIKey` is an optional field that, when `true`, authenticates using an OCI API-key configuration file (see `credentialSecretRef`). When `false` (the default), authenticates using instance principal credentials, intended for Vault instances running on an OCI compute instance — in that mode, no `credentialSecretRef` is needed.

```yaml
spec:
  backend:
    oci:
      authTypeAPIKey: true
```

### oci.credentialSecretRef

`oci.credentialSecretRef` specifies the name of the `Secret`, in the same namespace as the `VaultServer`, containing an OCI API-key configuration file (in the same format as the default `~/.oci/config` file) and the private key it references. It is required when `authTypeAPIKey` is `true`, and only consulted then. The secret contains the following keys:

- `config`
- the private key, under any key name, for example `oci_api_key.pem`

Every key of the secret is mounted in `/etc/vault/storage/oci/.oci`, so `key_file` in the config must point there by its full path; `~` is not expanded to that directory.

```ini
[DEFAULT]
user=ocid1.user.oc1..<unique_id>
fingerprint=<key_fingerprint>
tenancy=ocid1.tenancy.oc1..<unique_id>
region=us-phoenix-1
key_file=/etc/vault/storage/oci/.oci/oci_api_key.pem
```

```bash
$ kubectl create secret generic oci-cred -n demo \
    --from-file=config=./config \
    --from-file=oci_api_key.pem=./oci_api_key.pem
```

```yaml
spec:
  backend:
    oci:
      credentialSecretRef:
        name: oci-cred
```

### oci.haEnabled

`oci.haEnabled` is an optional field that enables High Availability support. Requires `lockBucketName` to also be set. This field accepts a boolean value. The default value is `false`.

```yaml
spec:
  backend:
    oci:
      haEnabled: true
      lockBucketName: "vault-lock-bucket"
```

### oci.lockBucketName

`oci.lockBucketName` is an optional field that specifies the name of a second bucket used to store HA lock records. Required when `haEnabled` is `true`. Must already exist and should generally be a different bucket than `bucketName`.

```yaml
spec:
  backend:
    oci:
      lockBucketName: "vault-lock-bucket"
```
