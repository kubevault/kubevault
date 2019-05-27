---
title: AzureAccessKeyRequest | Vault Secret Engine
menu:
  docs_0.2.0:
    identifier: azureaccesskeyrequest-secret-engine-crds
    name: AzureAccessKeyRequest
    parent: secret-engine-crds-concepts
    weight: 15
menu_name: docs_0.2.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# AzureAccessKeyRequest CRD

`AzureAccessKeyRequest` CRD can be used to request a new service principal based on a named role using a Vault server. If `AzureAccessKeyRequest` is approved, then vault operator will issue credentials via a Vault server and create Kubernetes Secret containing these credentials. The Secret name will be set in `status.secret.name` field.

When a `AzureAccessKeyRequest` is created, it makes a request to a Vault server for a new service principal under a `role`. Hence an [AzureRole](/docs/concepts/secret-engine-crds/azurerole.md) CRD is a prerequisite for creating an `AzureAccessKeyRequest`.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AzureAccessKeyRequest
metadata:
  name: <name>
  namespace: <namespace>
spec:
  roleRef: ... ... ...
  subjects: ... ... ...
status: ... ... ...
```

Vault operator performs the following operations when a AzureAccessKeyRequest CRD is created:

- Checks whether `status.conditions[].type` is `Approved` or not
- If Approved, makes request to the Vault server for credentials
- Creates a Kubernetes Secret which contains the credentials
- Assigns read permissions on that Kubernetes secret to specified subjects or user identities

Sample [AzureRole](/docs/concepts/secret-engine-crds/azurerole.md):

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AzureRole
metadata:
  name: demo-role
  namespace: demo
spec:
  authManagerRef:
    name: vault-app
    namespace: demo
  applicationObjectID: c1cb042d-96d7-423a-8dba
  config:
    subscriptionID: 1bfc9f66-316d-433e-b13d
    tenantID: 772268e5-d940-4bf6-be82
    clientID: 2b871d4a-757e-4b2f-bc78
    clientSecret: azure-client-secret
    environment: AzurePublicCloud
  ttl: 0h
  maxTTL: 0h
```

Sample AzureAccessKeyRequest under `demo-role` role:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AzureAccessKeyRequest
metadata:
  name: azure-credential
  namespace: demo
spec:
  roleRef:
    name: demo-role
    namespace: demo
  subjects:
    - kind: ServiceAccount
      name: sa
      namespace: demo
status:
  conditions:
    - type: Approved
```

## AzureAccessKeyRequest Spec

AzureAccessKeyRequest `Spec` contains information about [AzureRole](/docs/concepts/secret-engine-crds/azurerole.md) and subjects.

```yaml
spec:
  roleRef: <azureRole-reference>
  subjects: <list-of-subjects>
```

### spec.roleRef

`spec.roleRef` is a required field that specifies the [AzureRole](/docs/concepts/secret-engine-crds/azurerole.md) against which credential will be issued.

```yaml
spec:
  roleRef:
    name: demo-role
    namespace: demo
```

It has following field:

- `roleRef.name` : `Required`. Specifies the name of the object being referenced.
- `roleRef.namespace` : `Required`. Specifies the namespace of the referenced object.

### spec.subjects

`spec.subjects` is a required field that contains a list of reference to the object or user identity a role binding applies to. It will have read access to the credential secret. This can either hold a direct API object reference, or a value for non-objects such as user and group names.

```yaml
spec:
  subjects:
    - kind: ServiceAccount
      name: sa
      namespace: demo
```

## AzureAccessKeyRequest Status

`status` shows the status of the AzureAccessKeyRequest. It is maintained by Vault operator. It contains following fields:

- `secret` : Specifies the name of the secret containing Azure credential.

- `lease` : Contains lease information of the issued credential.

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

> Note: Azure credential will be issued if `conditions[].type` is `Approved`. Otherwise, Vault operator will not issue any credential.
