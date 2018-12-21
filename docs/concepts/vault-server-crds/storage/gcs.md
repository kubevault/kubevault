# Google Cloud Storage (GCS)

In Google Cloud Storage (GCS) storage backend, data will be stored in [Google Cloud Storage](https://cloud.google.com/storage/docs/). Vault documentation for GCS storage can be found in [here](https://www.vaultproject.io/docs/configuration/storage/google-cloud-storage.html).


```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault-with-gcs
  namespace: demo
spec:
  nodes: 1
  version: "0.11.1"
  backend:
    gcs:
      bucket: "my-vault-storage"
      credentialSecret: "my-gcs-credential"
```

## spec.backend.gcs

To use GCS as backend storage in Vault specify `spec.backend.gcs` in [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    gcs:
      bucket: <bucket_name>
      chunkSize: <chunk_size>
      maxParallel: <max_parallet>
      haEnabled: <true/false>
      credentialSecret: <secret_name>
```

`spec.backend.gcs` has following fields:

#### gcs.bucket

`gcs.bucket` is a required field that specifies the name of the bucket to use for storage.

```yaml
spec:
  backend:
    gcs:
      bucket: "my-vault-storage"
```

#### gcs.chunkSize

`gcs.chunkSize` is an optional field that specifies the maximum size (in kilobytes) to send in a single request. If this field is not specified, then Vault will set value `8192`. If this filed is set to 0, Vault will attempt to send the whole object at once, but will not retry any failures.

```yaml
spec:
  backend:
    gcs:
      chunkSize: "1024"
```

#### gcs.credentialSecret

`gcs.credentialSecret` is a required field that specifies the name of the secret containing Google application credential. The secret contains following key:
  - `sa.json`

```yaml
spec:
  backend:
    gcs:
      credentialSecret: "google-application-credential"
```

#### gcs.haEnabled

`gcs.haEnabled` is an optional field that specifies if high availability mode is enabled. This field accepts boolean value. Default value is `false`.

```yaml
spec:
  backend:
    gcs:
      haEnabled: true
```

#### gcs.maxParallel

`maxParallel` is an optional field that specifies the maximum number of parallel operations to take place. This field accepts integer value. If this field is not specified, then Vault will set value `128`.

```yaml
spec:
  backend:
    gcs:
      maxParallel: 124
```
