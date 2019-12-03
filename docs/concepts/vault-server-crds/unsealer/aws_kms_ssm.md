---
title: AWS KMS | Vault Unsealer
menu:
  docs_{{ .version }}:
    identifier: aws-kms-ssm-unsealer
    name: Aws KMS
    parent: unsealer-vault-server-crds
    weight: 1
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# mode.awsKmsSsm

To use **awsKmsSsm** mode specify `mode.awsKmsSsm`. In this mode, unseal keys and root token will be stored in [AWS System Manager Parameter store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html) and they will be encrypted using AWS encryption key.

```yaml
spec:
  unsealer:
    mode:
      awsKmsSsm:
        kmsKeyID: <key_id>
        region: <region>
        ssmKeyPrefix: <key_prefix>
        credentialSecret: <secret_name>
        endpoint: <vcp-endpoint>
```

`mode.awsKmsSsm` has the following fields:

## awsKmsSsm.kmsKeyID

`awsKmsSsm.kmsKeyID` is a required field that specifies the ID or ARN of the AWS KMS key to encrypt values.

```yaml
spec:
  unsealer:
    mode:
      awsKmsSsm:
        kmsKeyID: "aaaaa-bbbb-cccc-ddd-eeeeeeee"
```

## awsKmsSsm.region

`awsKmsSsm.region` is a required field that specifies the AWS region.

```yaml
spec:
  unsealer:
    mode:
      awsKmsSsm:
        region: "us-east-1"
```

## awsKmsSsm.ssmKeyPrefix

`awsKmsSsm.ssmKeyPrefix` is an optional field that specifies the prefix for SSM parameters. If this is not specified, then Unsealer will store parameters at the root of SSM.

```yaml
spec:
  unsealer:
    mode:
      awsKmsSsm:
        ssmKeyPrefix: "/cluster/demo"
```

## awsKmsSsm.credentialSecret

`awsKmsSsm.credentialSecret` is an optional field that specifies the name of the secret containing AWS access key and AWS secret key. If this is not specified, then Unsealer will attempt to retrieve credentials from the AWS metadata service. The secret contains the following data fields:

- `access_key`
- `secret_key`

```yaml
spec:
  unsealer:
    mode:
      awsKmsSsm:
        credentialSecret: "aws-cred"
```
