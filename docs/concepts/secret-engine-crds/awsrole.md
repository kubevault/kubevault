---
title: AwsRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: awsrole-secret-engine-crds
    name: AwsRole
    parent: secret-engine-crds-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# AWSRole CRD

Vault operator will configure [root IAM credentials](https://www.vaultproject.io/api/secret/aws/index.html#configure-root-iam-credentials) and create [role](https://www.vaultproject.io/api/secret/aws/index.html#create-update-role) according to `AWSRole` CRD (CustomResourceDefinition) specification. If the user deletes the `AWSRole` CRD, then respective role will also be deleted from Vault.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AWSRole
metadata:
  name: <name>
  namespace: <namespace>
spec:
  ...
status:
  ...
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{spec.clusterName}.{spec.namespace}.{spec.name}`

## AWSRole Spec

AWSRole `spec` contains root IAM credentials configuration and role information.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AWSRole
metadata:
  name: aws-role
  namespace: demo
spec:
  credentialType: iam_user
  policy:
    Version: '2012-10-17'
    Statement:
    - Effect: Allow
      Action: ec2:*
      Resource: "*"
  ref:
    namespace: demo
    name: vault-app
  config:
    credentialSecret: aws-cred
    region: us-east-1
    leaseConfig:
      lease: 1h
      leaseMax: 1h
```

AWSRole Spec has following fields:

### spec.ref

`spec.ref` specifies the name and namespace of [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains information to communicate with Vault.

```yaml
spec:
  ref:
    name: vault-app
    namespace: demo
```

### spec.config

`spec.config` is a required field that specifies the configuration of the root IAM credentials to communicate with AWS. If credentials already exist, this will overwrite them.

```yaml
spec:
  config:
    credentialSecret: aws-cred
    region: us-east-1
    leaseConfig:
      lease: 1h
      leaseMax: 1h
```

It has following fields:

- `config.credentialSecret` : `Required`. Specifies the name of the secret containing AWS credentials. The secret must contains following fields:
    - `access_key`
    - `secret_key`

- `config.region` : `Required`. Specifies the AWS region.

- `config.iamEndpoint` : `Optional`. Specifies a custom HTTP IAM endpoint to use.

- `config.stsEndpoint` : `Optional`. Specifies a custom HTTP STS endpoint to use.

- `config.maxRetries` : `Optional`. Specifies the number of max retries the client should use for recoverable errors.

- `config.leaseConfig` : `Optional`. Specifies the lease configuration.

    ```yaml
    config:
      leaseConfig:
        lease: 1h
        leaseMax: 1h
    ```

    It has following fields:
    - `leaseConfig.lease` : `Optional`. Specifies the lease value. Accepts time suffixed strings ("1h").
    - `leaseConfig.leaseMax` : `Optional`. Specifies the maximum lease value. Accepts time suffixed strings ("1h").

### spec.credentialType

`spec.credentialType` is a required field that specifies the type of credential to be used when retrieving credentials from the role. Supported types: `iam_user`, `assumed_role` and `federation_token`.

```yaml
spec:
  credentialType: iam_user
```

### spec.roleARNs

`spec.roleARNs` is an optional field that specifies the list of ARNs of the AWS roles this Vault role is allowed to assume.

```yaml
spec:
  roleARNs:
    - arn:aws:iam::452618475015:role/hello.world
```

### spec.policyARNs

`spec.policyARNs` is an optional field that specifies the list of ARNs of the AWS managed policies to be attached to IAM users when they are requested.

```yaml
spec:
  policyARNs:
    - arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess
```

### spec.policyDocument

`spec.policyDocument` is an optional field that specifies the IAM policy document for the role.

```yaml
spec:
  policyDocument: |
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": "ec2:*",
          "Resource": "*"
        }
      ]
    }
```

### spec.defaultSTSTTL

`spec.defaultSTSTTL` is an optional field that specifies the default TTL for STS credentials. When a TTL is not specified when STS credentials are requested, and a default TTL is specified on the role, then this default TTL will be used. Valid only when `spec.credentialType` is one of `assumed_role` or `federation_token`.

```yaml
spec:
  defaultSTSTTL: "1h"
```

### spec.maxSTSTTL

`spec.maxSTSTTL` is an optional field that specifies the max allowed TTL for STS credentials. Valid only when `spec.credentialType` is one of `assumed_role` or `federation_token`.

```yaml
spec:
  maxSTSTTL: "1h"
```

### spec.policy

`spec.policy` is an optional field that specifies the IAM policy in JSON format. This field is for backwards compatibility only.

### spec.arn

`spec.arn` is an optional field that specifies the full ARN reference to the desired existing policy. This field is for backwards compatibility only.

## AWSRole Status

`status` shows the status of the AWSRole. It is maintained by Vault operator. It contains following fields:

- `phase` : Indicates whether the role successfully applied in vault or not or in progress or failed

- `conditions` : Represent observations of a AWSRole.
