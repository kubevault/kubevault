---
title: AWSAccessKeyRequest | Vault Secret Engine
menu:
  docs_0.1.0:
    identifier: awsaccesskeyrequest-secret-engine-crds
    name: AWSAccessKeyRequest
    parent: secret-engine-crds-concepts
    weight: 15
menu_name: docs_0.1.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# AWSAccessKeyRequest CRD

`AWSAccessKeyRequest` CRD is to request AWS credential from vault. If `AWSAccessKeyRequest` is approved, then Vault operator will issue credential from vault and create kubernetes secret containing credential. The secret name will be specified in `status.secret.name` field.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AWSAccessKeyRequest
metadata:
  name: <name>
  namespace: <namespace>
spec:
  ...
status:
  ...
```

## AWSAccessKeyRequest Spec

AWSAccessKeyRequest `spec` contains information about AWS role and subject.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AWSAccessKeyRequest
metadata:
  name: aws-cred
  namespace: demo
spec:
  roleRef:
    name: aws-role
    namespace: demo
  subjects:
  - kind: ServiceAccount
    name: sa
    namespace: demo
```

AWSAccessKeyRequest Spec has following fields:

### spec.roleRef

`spec.roleRef` is a required field that specifies the [AWSRole](/docs/concepts/secret-engine-crds/awsrole.md) against which credential will be issued.

```yaml
spec:
  roleRef:
    name: aws-role
    namespace: demo
```

It has following field:

- `roleRef.name` : `Required`. Specifies the name of the object being referenced.
- `roleRef.namespace` : `Required`. Specifies the namespace of the referenced object.

### spec.subjects

`spec.subjects` is a required field that contains a reference to the object or user identities a role binding applies to. It will have read access of the credential secret. This can either hold a direct API object reference, or a value for non-objects such as user and group names.

```yaml
spec:
  subjects:
    - kind: ServiceAccount
      name: sa
      namespace: demo
```

### spec.ttl

`spec.ttl` is an optional field that specifies the TTL for the use of the STS token. This is specified as a string with a duration suffix.

```yaml
spec:
  ttl: "1h"
```

### spec.roleARN

`spec.roleARN` is an optional field that specifies the ARN of the role to assume if `credential_type` on the Vault role is `assumed_role`. Must match one of the allowed role ARNs in the Vault role. Optional if the Vault role only allows a single AWS role ARN, required otherwise.

```yaml
spec:
  roleARN: "arn:aws:iam::452618475015:role/hello.world"
```

### spec.useSTS
`spec.useSTS` is an optional field. If this is `true`, `/aws/sts` endpoint will be used to retrieve credential. Otherwise, `/aws/creds` endpoint will be used to retrieve credential.

```yaml
spec:
  useSTS: true
```

## AWSAccessKeyRequest Status

`status` shows the status of the AWSAccessKeyRequest. It is maintained by Vault operator. It contains following fields:

- `secret` : Specifies the name of the secret containing AWS credential.

- `lease` : Contains lease information of the issued credential.

- `conditions` : Represent observations of a AWSAccessKeyRequest.

    ```yaml
    status:
      conditions:
        - type: Approved
    ```

  It has following field:
  - `conditions[].type` : `Required`. Specifies request approval state. Supported type: `Approved` and `Denied`.
  - `conditions[].reason` : `Optional`. Specifies brief reason for the request state.
  - `conditions[].message` : `Optional`. Specifies human readable message with details about the request state.

> Note: AWS credential will be issued if `conditions[].type` is `Approved`. Otherwise, Vault operator will not issue any credential.
