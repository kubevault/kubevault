---
title: Secret Access Request
menu:
  docs_{{ .version }}:
    identifier: secret-access-request-secret-engine-crds
    name: SecretAccessRequest
    parent: secret-engine-crds-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# SecretAccessRequest

## What is SecretAccessRequest

A `SecretAccessRequest` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to request a Vault server for credentials in a Kubernetes native way.
A `SecretAccessRequest` can be created under various `roleRef` e.g: `AWSRole`, `GCPRole`, `ElasticsearchRole`, `MongoDBRole`, etc. A `SecretAccessRequest` has three different phases e.g: 
`WaitingForApproval`, `Approved`, `Denied`.  If `SecretAccessRequest` is approved, then the KubeVault operator will issue credentials and create Kubernetes secret containing credentials. The secret name will be specified in `status.secret.name` field.


![SecretAccessRequest CRD](/docs/images/concepts/database_accesskey_request.svg)

KubeVault operator performs the following operations when a `SecretAccessRequest` CRD is created:

- Checks whether `status.conditions[].type` is `Approved` or not
- If Approved, requests the Vault server for credentials
- Creates a Kubernetes Secret which contains the credentials
- Sets the name of the k8s secret to SecretAccessRequest's `status.secret.name`
- Assigns read permissions on that Kubernetes secret to specified subjects or user identities

## SecretAccessRequest CRD Specification

Like any official Kubernetes resource, a `SecretAccessRequest` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `SecretAccessRequest` object for the `AWSRole` is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: aws-cred-req
  namespace: dev
spec:
  roleRef:
    kind: AWSRole
    name: aws-role
  subjects:
    - kind: ServiceAccount
      name: test-user-account
      namespace: test
```

Here, we are going to describe the various sections of the `SecretAccessRequest` crd.

### SecretAccessRequest Spec

SecretAccessRequest `spec` contains information about database role and subject.

```yaml
spec:
  roleRef:
    apiGroup: <role-apiGroup>
    kind: <role-kind>
    name: <role-name>
  subjects:
    - kind: <subject-kind>
      apiGroup: <subject-apiGroup>
      name: <subject-name>
      namespace: <subject-namespace>
  ttl: <ttl-for-leases>
```

`SecretAccessRequest` spec has the following fields:

#### spec.roleRef

`spec.roleRef` is a `required` field that specifies the role against which credentials will be issued.

It has the following fields:

- `roleRef.apiGroup` : `Optional`. Specifies the APIGroup of the resource being referenced.

- `roleRef.kind` : `Required`. Specifies the kind of the resource being referenced.

- `roleRef.name` : `Required`. Specifies the name of the object being referenced.


```yaml
spec:
  roleRef:
    kind: AWSRole
    name: aws-role
```

#### spec.subjects

`spec.subjects` is a `required` field that contains a list of references to the object or user identities on whose behalf this request is made. These object or user identities will have read access to the k8s credential secret. This can either hold a direct API object reference or a value for non-objects such as user and group names.

It has the following fields:

- `kind` : `Required`. Specifies the kind of object being referenced. Values defined by
  these API groups are "User", "Group", and "ServiceAccount". If the Authorizer does not
  recognize the kind value, the Authorizer will report an error.

- `apiGroup` : `Optional`. Specifies the APIGroup that holds the API group of the referenced subject.
  Defaults to `""` for ServiceAccount subjects.

- `name` : `Required`. Specifies the name of the object being referenced.

- `namespace`: `Required`. Specifies the namespace of the object being referenced.

```yaml
spec:
  subjects:
    - kind: ServiceAccount
      name: sa
      namespace: demo
```

#### spec.ttl

`spec.ttl` is an `optional` field that specifies the TTL for the use
of the STS token. This is specified as a string with a duration suffix.

```yaml
spec:
  ttl: "1h"
```

### SecretAccessRequest Status

`status` shows the status of the `SecretAccessRequest`. It is managed by the KubeVault operator. It contains the following fields:

- `secret`: 
  - `secret.name`: Specifies the name of the secret containing the credential.
  - `secret.namespace`: Specifies the namespace of the secret containing the credential.

- `lease`: Contains lease information of the issued credential.

- `conditions` : Represent observations of a `SecretAccessRequest`. It has the following fields:
  - `conditions[].type` : Specifies request approval state. Supported type: `Approved` and `Denied`, `Available`.
  - `conditions[].status` : Specifies request approval status. Supported type: `True`, `False`.
  - `conditions[].reason` : Specifies brief reason for the request state.
  - `conditions[].message` : Specifies human-readable message with details about the request state.
  - `conditions[].observerGeneration`: Specifies ObserverGeneration for the request state.

- `phase` : Represent the phase of the `SecretAccessRequest`.

```yaml
status:
  conditions:
  - lastTransitionTime: "2021-09-28T09:36:45Z"
    message: 'This was approved by: kubectl vault approve secretaccessrequest'
    observedGeneration: 1
    reason: KubectlApprove
    status: "True"
    type: Approved
  - lastTransitionTime: "2021-09-28T09:36:49Z"
    message: The requested credentials successfully issued.
    observedGeneration: 1
    reason: SuccessfullyIssuedCredential
    status: "True"
    type: Available
  lease:
    duration: 1h0m0s
    id: k8s.-.aws.dev.aws-secret-engine/creds/k8s.-.dev.aws-role/ACUzSSp5aLVBzNhoqe6wEqaW
    renewable: true
  observedGeneration: 1
  phase: Approved
  secret:
    name: aws-cred-req-92m0n9
    namespace: dev

```

> Note: Credential will be issued only if the `status.phase` is `Approved`. Otherwise, the KubeVault operator will not issue any credentials.