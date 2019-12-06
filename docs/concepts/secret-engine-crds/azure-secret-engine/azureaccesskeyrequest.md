---
title: AzureAccessKeyRequest | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: azureaccesskeyrequest-secret-engine-crds
    name: AzureAccessKeyRequest
    parent: azure-crds-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# AzureAccessKeyRequest

## What is AzureAccessKeyRequest

An `AzureAccessKeyRequest` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to request for Azure credentials in a Kubernetes native way. If an `AzureAccessKeyRequest` is approved, then the KubeVault operator will issue credentials using a Vault server and create a Kubernetes secret containing the Azure credentials. The secret name will be specified in `status.secret.name` field. The operator will also create appropriate `ClusterRole` and `ClusterRoleBinding` for the k8s secret.

When an `AzureAccessKeyRequest` is created, it makes a request to a Vault server for a new service principal under a `role`. Hence we need to deploy an [AzureRole](/docs/concepts/secret-engine-crds/azure-secret-engine/azurerole.md) CRD before creating an `AzureAccessKeyRequest`.

![AzureAccessKeyRequest CRD](/docs/images/concepts/azure_accesskey_request.svg)

The KubeVault operator performs the following operations when an AzureAccessKeyRequest CRD is created:

- Checks whether `status.conditions[].type` is `Approved` or not
- If Approved, makes Azure access key request to Vault
- Creates a Kubernetes Secret which contains the Azure credentials
- Sets the name of the k8s secret to AzureAccessKeyRequest's `status.secret.name`
- Assigns read permissions on that Kubernetes secret to specified subjects or user identities

## AzureAccessKeyRequest CRD Specification

Like any official Kubernetes resource, a `AzureAccessKeyRequest` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `AzureAccessKeyRequest` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AzureAccessKeyRequest
metadata:
  name: azure-cred-req
  namespace: demo
spec:
  roleRef:
    name: azure-role
    namespace: demo
  subjects:
    - kind: ServiceAccount
      name: sa
      namespace: demo
status:
  ... ...
```

Here, we are going to describe the various sections of the `AzureAccessKeyRequest` crd.

### AzureAccessKeyRequest Spec

AzureAccessKeyRequest `Spec` contains information about [AzureRole](/docs/concepts/secret-engine-crds/azure-secret-engine/azurerole.md) and subjects.

```yaml
spec:
  roleRef:
    apiGroup: <role-apiGroup>
    kind: <role-kind>
    name: <role-name>
    namespace: <role-namespace>
  subjects:
    - kind: <subject-kind>
      apiGroup: <subject-apiGroup>
      name: <subject-name>
      namespace: <subject-namespace>
```

AzureAccessKeyRequest spec has the following fields:

#### spec.roleRef

`spec.roleRef` is a `required` field that specifies the [AzureRole](/docs/concepts/secret-engine-crds/azure-secret-engine/azurerole.md) against which credentials will be issued.

It has the following fields:

- `roleRef.apiGroup` : `Optional`. Specifies the APIGroup of the resource being referenced.
- `roleRef.kind` : `Optional`. Specifies the kind of the resource being referenced.
- `roleRef.name` : `Required`. Specifies the name of the object being referenced.
- `roleRef.namespace` : `Required`. Specifies the namespace of the referenced object.

```yaml
spec:
  roleRef:
    name: azure-role
    namespace: demo
```

#### spec.subjects

`spec.subjects` is a `required` field that contains a list of references to the object or
user identities on whose behalf this request is made. These object or user identities will have
read access to the k8s credential secret. This can either hold a direct API object reference or a value for non-objects such as user and group names.

It has the following fields:

- `kind` : `Required`. Specifies the kind of object being referenced. Values defined by this API group are "User", "Group", and "ServiceAccount". If the Authorizer does not recognize the kind value, the Authorizer will report an error.

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

### AzureAccessKeyRequest Status

`status` shows the status of the AzureAccessKeyRequest. It is managed by the KubeVault operator. It contains the following fields:

- `secret`: Specifies the name of the secret containing Azure credential.

- `lease`: Contains lease information of the issued credential.

- `conditions` : Represent observations of a AzureAccessKeyRequest.

  ```yaml
  status:
    conditions:
      - type: Approved
  ```

  It has following field:

  - `conditions[].type` : `Required`. Specifies request approval state. Supported type: `Approved` and `Denied`.
  - `conditions[].reason` : `Optional`. Specifies brief reason for the request state.
  - `conditions[].message` : `Optional`. Specifies human readable message with details about the request state.

> Note: Azure credential will be issued if `conditions[].type` is `Approved`. Otherwise, the KubeVault operator will not issue any credentials.
