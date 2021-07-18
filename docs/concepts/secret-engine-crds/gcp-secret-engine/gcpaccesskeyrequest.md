---
title: GCPAccessKeyRequest | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: gcpaccesskeyrequest-secret-engine-crds
    name: GCPAccessKeyRequest
    parent: gcp-crds-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# GCPAccessKeyRequest

## What is GCPAccessKeyRequest

An `GCPAccessKeyRequest` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to request for GCP credentials in a Kubernetes native way. If an `GCPAccessKeyRequest` is approved, then the KubeVault operator will issue credentials using a Vault server and create a Kubernetes secret containing the GCP credentials. The secret name will be specified in `status.secret.name` field. The operator will also create appropriate `ClusterRole` and `ClusterRoleBinding` for the k8s secret.

When a `GCPAccessKeyRequest` is created, it make an  access key request to Vault under a `roleset`. Hence a [GCPRole](/docs/concepts/secret-engine-crds/gcp-secret-engine/gcprole.md) CRD is a prerequisite for creating a `GCPAccessKeyRequest`.

![GCPAccessKeyRequest CRD](/docs/images/concepts/gcp_accesskey_request.svg)

The KubeVault operator performs the following operations when a GCPAccessKeyRequest CRD is created:

- Checks whether `status.conditions[].type` is `Approved` or not
- If Approved, makes GCP access key request to Vault
- Creates a Kubernetes Secret which contains the GCP credentials
- Sets the name of the k8s secret to GCPAccessKeyRequest's `status.secret.name`
- Assigns read permissions on that Kubernetes secret to specified subjects or user identities

## GCPAccessKeyRequest CRD Specification

Like any official Kubernetes resource, a `GCPAccessKeyRequest` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `GCPAccessKeyRequest` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: GCPAccessKeyRequest
metadata:
  name: gcp-cred-req
  namespace: demo
spec:
  roleRef:
    name: gcp-role
    namespace: demo
  subjects:
    - kind: ServiceAccount
      name: sa
      namespace: demo
status:
  conditions:
  - lastTransitionTime: "2019-11-14T09:21:49Z"
    message: This was approved by kubectl vault approve gcpaccesskeyrequest
    reason: KubectlApprove
    type: Approved
    status: True
  lease:
    duration: 0s
  secret:
    name: gcp-cred-req-luc5p4
```

Here, we are going to describe the various sections of the `GCPAccessKeyRequest` crd.

### GCPAccessKeyRequest Spec

GCPAccessKeyRequest `Spec` contains information about
[GCPRole](/docs/concepts/secret-engine-crds/gcp-secret-engine/gcprole.md) and subjects.

```yaml
spec:
  roleRef: <GCPRole-reference>
  subjects: <list-of-subjects>
```

`Spec` contains two additional fields only if the referred GCPRole has `spec.secretType` of `service_account_key`.

```yaml
spec:
  roleRef: <GCPRole-reference>
  subjects: <list-of-subjects>
  keyAlgorithm: <algorithm_used_to_generate_key>
  keyType: <private_key_type>
```

GCPAccessKeyRequest spec has the following fields:

#### spec.roleRef

`spec.roleRef` is a `required` field that specifies the
[GCPRole](/docs/concepts/secret-engine-crds/gcp-secret-engine/gcprole.md) against which credentials will be issued.

It has the following fields:

- `roleRef.apiGroup` : `Optional`. Specifies the APIGroup of the resource being referenced.

- `roleRef.kind` : `Optional`. Specifies the kind of the resource being referenced.

- `roleRef.name` : `Required`. Specifies the name of the object being referenced.

- `roleRef.namespace` : `Required`. Specifies the namespace of the referenced object.

```yaml
spec:
  roleRef:
    name: gcp-role
    namespace: demo
```

#### spec.subjects

`spec.subjects` is a `required` field that contains a list of references to the object or user identities on whose behalf this request is made. These object or user identities will have read access to the k8s credential secret. This can either hold a direct API object reference or a value for non-objects such as user and group names.

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

#### spec.keyAlgorithm

`spec.keyAlgorithm` is an `optional` field that specifies the key algorithm used to generate the key. Defaults to 2k RSA key. You probably should not choose other values (i.e. 1k), but accepted values are `KEY_ALG_UNSPECIFIED`, `KEY_ALG_RSA_1024`, `KEY_ALG_RSA_2048`  

```yaml
spec:
  keyAlgorithm: KEY_ALG_RSA_2048
```

#### spec.keyType

`spec.keyType` is an `optional` field that specifies the private key type to generate.
Defaults to JSON credentials file. Accepted values are `TYPE_UNSPECIFIED`, `TYPE_GOOGLE_CREDENTIALS_FILE`

```yaml
spec:
  keyType: TYPE_GOOGLE_CREDENTIALS_FILE
```

### GCPAccessKeyRequest Status

`status` shows the status of the GCPAccessKeyRequest. It is managed by the KubeVault operator. It contains the following fields:

- `secret`: Specifies the name of the secret containing GCP credential.

- `lease`: Contains lease information of the issued credential.

- `conditions` : Represent observations of a GCPAccessKeyRequest.

    ```yaml
    status:
      conditions:
        - type: Approved
    ```

  It has following field:
  
  - `conditions[].type` : `Required`. Specifies request approval state. Supported type: `Approved` and `Denied`.
  
  - `conditions[].reason` : `Optional`. Specifies brief reason for the request state.
  
  - `conditions[].message` : `Optional`. Specifies human readable message with details about the request state.

> Note: GCP credential will be issued if `conditions[].type` is `Approved`. Otherwise, the KubeVault operator will not issue any credentials.
