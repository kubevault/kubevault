---
title: Google Cloud KMS | Vault Unsealer
menu:
  docs_{{ .version }}:
    identifier: google-kms-gcs-unsealer
    name: Google Cloud KMS
    parent: unsealer-vault-server-crds
    weight: 1
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# mode.googleKmsGcs

To use **googleKmsGcs** mode specify `mode.googleKmsGcs`. In this mode, unseal keys and root token will be stored in [Google Cloud Storage](https://cloud.google.com/storage/docs/) and they will be encrypted using google cryptographic keys.

```yaml
spec:
  unsealer:
    mode:
      googleKmsGcs:
        bucket: <bucket_name>
        kmsProject: <project_name>
        kmsLocation: <location>
        kmsKeyRing: <key_ring_name>
        kmsCryptoKey: <crypto_key_name>
        credentialSecret: <secret_name>
```

`mode.googleKmsGcs` has the following fields:

## googleKmsGcs.bucket

`googleKmsGcs.bucket` is a required field that specifies the name of the bucket to store keys.

```yaml
spec:
  unsealer:
    mode:
      googleKmsGcs:
        bucket: "vault-key-store"
```

## googleKmsGcs.kmsProject

`googleKmsGcs.kmsProject` is a required field that specifies the name of the projects under which the keyring is created.

```yaml
spec:
  unsealer:
    mode:
      googleKmsGcs:
        kmsProject: "project"
```

## googleKmsGcs.kmsLocation

`googleKmsGcs.kmsLocation` is a required field that specifies the location of the keyring.

```yaml
spec:
  unsealer:
    mode:
      googleKmsGcs:
        kmsLocation: "global"
```

## googleKmsGcs.kmsKeyRing

`googleKmsGcs.kmsKeyRing` is a required field that specifies the name of the keyring.

```yaml
spec:
  unsealer:
    mode:
      googleKmsGcs:
        kmsKeyRing: "key-ring"
```

## googleKmsGcs.kmsCryptoKey

`googleKmsGcs.kmsCryptoKey` is a required field that specifies the name of the crypto key.

```yaml
spec:
  unsealer:
    mode:
      googleKmsGcs:
        kmsCryptoKey: "key"
```

## googleKmsGcs.credentialSecret

`googleKmsGcs.credentialSecret` is an optional field that specifies the name of the secret containing google credentials. If this is not specified, then the instance service account will be used (if it is running on google cloud). The secret contains the following field:

- `sa.json`

```yaml
spec:
  unsealer:
    mode:
      googleKmsGcs:
        credentialSecret: "google-cred"
```
