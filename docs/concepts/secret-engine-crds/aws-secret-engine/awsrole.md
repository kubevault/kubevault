---
title: AwsRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: awsrole-secret-engine-crds
    name: AwsRole
    parent: aws-crds-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# AWSRole

## What is AWSRole

An `AWSRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create AWS secret engine role in a Kubernetes native way.

When an `AWSRole` is created, the KubeVault operator [configures](https://www.vaultproject.io/docs/secrets/aws/index.html#setup) a Vault role that maps to a set of permissions in AWS as well as an AWS credential type. When users generate credentials, they are generated against this role. If the user deletes the `AWSRole` CRD,
then the respective role will also be deleted from Vault.

![AWSRole CRD](/docs/images/concepts/aws_role.svg)

## AWSRole CRD Specification

Like any official Kubernetes resource, a `AWSRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `AWSRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AWSRole
metadata:
  name: aws-cred
  namespace: demo
spec:
  secretEngineRef:
    name: aws-secret-engine
  credentialType: iam_user
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
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `AWSRole` crd.

### AWSRole Spec

AWSRole `spec` contains root IAM credentials configuration and role information.

```yaml
spec:
  secretEngineRef:
    name: <secret-engine-name>
  path: <aws-secret-engine-path>
  credentialType: <credential-type>
  roleARNs:
    - "ARN1"
    - "ARN2"
  policyARNs:
    - "ARN1"
    - "ARN2"
  policyDocument: <IAM-policy-document>
  policy: <policy-in-yaml-format>
  defaultSTSTTL: <default-TTL-for-STS>
  maxSTSTTL: <max-TTL-for-STS>
```

`AWSRole` spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: aws-secret-engine
```

#### spec.credentialType

`spec.credentialType` is a `required` field that specifies the type of credential to be used when retrieving credentials from the role. Supported types: `iam_user`, `assumed_role` and `federation_token`.

```yaml
spec:
  credentialType: iam_user
```

#### spec.roleARNs

`spec.roleARNs` is an `optional` field that specifies the list of ARNs of the AWS roles this Vault role is allowed to assume.

```yaml
spec:
  roleARNs:
    - arn:aws:iam::452618475015:role/hello.world
```

#### spec.policyARNs

`spec.policyARNs` is an `optional` field that specifies the list of ARNs of the AWS managed policies to be attached to IAM users when they are requested.

```yaml
spec:
  policyARNs:
    - arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess
```

#### spec.policyDocument

`spec.policyDocument` is an `optional` field that specifies the IAM policy document for the role.

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

#### spec.policy

`spec.policy` is an `optional` field that specifies the IAM policy in JSON format.
 This field is for backward compatibility only.

```yaml
spec:
  policy:
    Version: '2012-10-17'
    Statement:
    - Effect: Allow
      Action: ec2:*
      Resource: "*"
```

#### spec.defaultSTSTTL

`spec.defaultSTSTTL` is an `optional` field that specifies the default TTL for STS credentials. When a TTL is not specified when STS credentials are requested, and a default TTL is specified
on the role, then this default TTL will be used. This is valid only when `spec.credentialType` is one of `assumed_role` or `federation_token`.

```yaml
spec:
  defaultSTSTTL: "1h"
```

#### spec.maxSTSTTL

`spec.maxSTSTTL` is an `optional` field that specifies the max allowed TTL for STS credentials. This is valid only when `spec.credentialType` is one of `assumed_role` or `federation_token`.

```yaml
spec:
  maxSTSTTL: "1h"
```

### AWSRole Status

`status` shows the status of the AWSRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation, which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of an AWSRole.
