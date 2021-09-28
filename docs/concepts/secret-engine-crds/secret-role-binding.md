---
title: Secret Role Binding
menu:
  docs_{{ .version }}:
    identifier: secret-role-binding-secret-engine-crds
    name: SecretRoleBinding
    parent: secret-engine-crds-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# SecretRoleBinding

## What is SecretRoleBinding

A `SecretRoleBinding` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to bind a set of roles to a set of users.
Using the `SecretRoleBinding` it's possible to bind various roles e.g: `AWSRole`, `GCPRole`, `ElasticsearchRole`, `MongoDBRole`, etc. to Kubernetes ServiceAccounts.
A `SecretRoleBinding` has three different phases e.g: `Processing`, `Success`, `Failed`. Once a `SecretRoleBinding` is successful, it will create a `VaultPolicy` and a `VaultPolicyBinding`.


## SecretRoleBinding CRD Specification

Like any official Kubernetes resource, a `SecretRoleBinding` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.
A sample `SecretRoleBinding` object that binds `AWSRole` to a Kubernetes `ServiceAccount` is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretRoleBinding
metadata:
  name: secret-r-binding
  namespace: dev
spec:
  roles:
    - kind: AWSRole
      name: aws-role
  subjects:
    - kind: ServiceAccount
      name: test-user-account 
      namespace: test
```

Here, we are going to describe the various sections of the `SecretRoleBinding` CRD.

### SecretRoleBinding Spec

SecretAccessRequest `spec` contains information about the role and the subjects.

```yaml
spec:
  roles:
    - kind: <role-kind>
      name: <role-name>
  subjects:
    - kind: <subject-kind>
      name: <subject-name>
      namespace: <subject-namespace>
```

`SecretRoleBinding` spec has the following fields:

#### spec.roles

`spec.roles` is a `required` field that specifies the roles list for which the `VaultPolicy` will be created.

It has the following fields:

- `roleRef.apiGroup` : `Optional`. Specifies the APIGroup of the resource being referenced.

- `roleRef.kind` : `Required`. Specifies the kind of the resource being referenced.

- `roleRef.name` : `Required`. Specifies the name of the object being referenced.


```yaml
spec:
  roles:
    - kind: <role-kind>
      name: <role-name>
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

- `namespace`: `Optional`. Specifies the namespace of the object being referenced.

```yaml
spec:
  subjects:
    - kind: ServiceAccount
      name: test-user-account
      namespace: test
```

### SecretRoleBinding Status

`status` shows the status of the `SecretRoleBinding`. It contains the following fields:

- `conditions` : Represent observations of a `SecretAccessRequest`. It has the following fields:
  - `conditions[].type` : Specifies request approval state. Supported type: `VaultPolicySuccess` and `VaultPolicyBindingSuccess`, `SecretRoleBindingSuccess`.
  - `conditions[].status` : Specifies request approval status. Supported type: `True`, `False`.
  - `conditions[].reason` : Specifies brief reason for the request state.
  - `conditions[].message` : Specifies human-readable message with details about the request state.
  - `conditions[].observerGeneration`: Specifies ObserverGeneration for the request state.

- `phase` : Represent the phase of the `SecretRoleBinding`. Supported type: `Success` and `Failed`, `Processing`.

- `policyRef` : Represent the `VaultPolicy` created by the `SecretRoleBinding`. 
  - `policyRef.name`: The name of the `VaultPolicy` created by the `SecretRoleBinding`.
  - `policyRef.namespace`: The namespace of the `VaultPolicy` created by the `SecretRoleBinding`.

- `policyBindingRef` : Represent the `VaultPolicyBinding` created by the `SecretRoleBinding`.
  - `policyRef.name`: The name of the `VaultPolicyBinding` created by the `SecretRoleBinding`.
  - `policyRef.namespace`: The namespace of the `VaultPolicyBinding` created by the `SecretRoleBinding`.


A Successful `SecretAccessRequest.status` may look like this:

```yaml
status:
  conditions:
  - lastTransitionTime: "2021-09-28T12:56:35Z"
    message: VaultPolicy phase is Successful
    observedGeneration: 1
    reason: VaultPolicySucceeded
    status: "True"
    type: VaultPolicySuccess
  - lastTransitionTime: "2021-09-28T12:56:35Z"
    message: VaultPolicyBinding is Successful
    observedGeneration: 1
    reason: VaultPolicyBindingSucceeded
    status: "True"
    type: VaultPolicyBindingSuccess
  - lastTransitionTime: "2021-09-28T12:56:35Z"
    message: SecretRoleBinding is Successful
    observedGeneration: 1
    reason: SecretRoleBindingSucceeded
    status: "True"
    type: SecretRoleBindingSuccess
  observedGeneration: 1
  phase: Success
  policyBindingRef:
    name: srb-dev-secret-r-binding
    namespace: demo
  policyRef:
    name: srb-dev-secret-r-binding
    namespace: demo

```

#### SecretRoleBinding status.policyRef

We can get the `VaultPolicy` if the `SecretRoleBinding` phase is `Success`:

```console
$ kubectl get vaultpolicy -n demo srb-dev-secret-r-binding -oyaml
```
```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  annotations:
    secretrolebindings.engine.kubevault.com/name: secret-r-binding
    secretrolebindings.engine.kubevault.com/namespace: dev
  creationTimestamp: "2021-09-28T13:04:15Z"
  finalizers:
  - kubevault.com
  generation: 1
  name: srb-dev-secret-r-binding
  namespace: demo
  ownerReferences:
  - apiVersion: kubevault.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: VaultServer
    name: vault
    uid: b73a5a72-d575-4b91-8e95-938828268535
  resourceVersion: "53571"
  uid: b4a2ba18-66c3-4f3c-aa35-71b0d66c845f
spec:
  policyDocument: |

    path "/k8s.-.aws.dev.aws-secret-engine/creds/k8s.-.dev.aws-role" {
      capabilities = ["read"]
    }

    path "/k8s.-.aws.dev.aws-secret-engine/sts/k8s.-.dev.aws-role" {
      capabilities = ["create", "update"]
    }
  vaultRef:
    name: vault
status:
  conditions:
  - lastTransitionTime: "2021-09-28T13:04:15Z"
    message: policy is ready to use
    reason: Provisioned
    status: "True"
    type: Available
  observedGeneration: 1
  phase: Success
```

#### SecretRoleBinding status.policyBindingRef

We can get the `VaultPolicyBinding` if the `SecretRoleBinding` phase is `Success`:

```console
$ kubectl get vaultpolicybinding -n demo srb-dev-secret-r-binding -oyaml
```
```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  annotations:
    secretrolebindings.engine.kubevault.com/name: secret-r-binding
    secretrolebindings.engine.kubevault.com/namespace: dev
  creationTimestamp: "2021-09-28T13:04:15Z"
  finalizers:
    - kubevault.com
  generation: 1
  name: srb-dev-secret-r-binding
  namespace: demo
  ownerReferences:
    - apiVersion: kubevault.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: VaultServer
      name: vault
      uid: b73a5a72-d575-4b91-8e95-938828268535
  resourceVersion: "53576"
  uid: c37dc7ca-03ca-4191-af6c-fe91e544394a
spec:
  policies:
    - ref: srb-dev-secret-r-binding
  subjectRef:
    kubernetes:
      name: k8s.-.demo.srb-dev-secret-r-binding
      path: kubernetes
      serviceAccountNames:
        - test-user-account
      serviceAccountNamespaces:
        - test
  vaultRef:
    name: vault
  vaultRoleName: k8s.-.demo.srb-dev-secret-r-binding
status:
  conditions:
    - lastTransitionTime: "2021-09-28T13:04:16Z"
      message: policy binding is ready to use
      reason: Provisioned
      status: "True"
      type: Available
  observedGeneration: 1
  phase: Success
```
