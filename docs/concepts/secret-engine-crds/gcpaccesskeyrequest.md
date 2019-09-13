---
title: GCPAccessKeyRequest | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: gcpaccesskeyrequest-secret-engine-crds
    name: GCPAccessKeyRequest
    parent: secret-engine-crds-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# GCPAccessKeyRequest CRD

`GCPAccessKeyRequest` CRD is to generate gcp secret (i.e. OAuth2 Access Token or Service Account Key) using vault. If `GCPAccessKeyRequest` is approved, then vault operator will issue credentials from vault and create Kubernetes Secret containing these credentials. The Secret name will be specified in `status.secret.name` field.

When a `GCPAccessKeyRequest` is created, it make an  access key request to vault under a `roleset`. Hence a [GCPRole](/docs/concepts/secret-engine-crds/gcprole.md) CRD which is successfully configured, is prerequisite for creating a `GCPAccessKeyRequest`.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: GCPAccessKeyRequest
metadata:
  name: <name>
  namespace: <namespace>
spec:
  roleRef:
    ... ... ...
  subjects:
    ... ... ...
status:
  ... ... ...
```

Vault operator performs the following operations when a GCPAccessKeyRequest CRD is created:

- Checks whether `status.conditions[].type` is `Approved` or not
- If Approved, makes gcp access key request to vault
- Creates a Kubernetes Secret which contains the gcp secrets
- Provides permissions of that kubernetes secret to specified objects or user identities

Example [GCPRole](/docs/concepts/secret-engine-crds/gcprole.md): 
```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: GCPRole
metadata: 
  name: gcp-role
  namespace: demo
spec:
  ref:
    name: vault-app
    namespace: demo
  config:
    credentialSecret: gcp-cred
  secretType: access_token
  project: ackube
  bindings: 'resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
        roles = ["roles/viewer"]
      }'
  tokenScopes: ["https://www.googleapis.com/auth/cloud-platform"]
```
Example GCPAccessKeyRequest under `gcp-role` roleset:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: GCPAccessKeyRequest
metadata:
  name: gcp-credential
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
    - type: Approved
```

## GCPAccessKeyRequest Spec

GCPAccessKeyRequest `Spec` contains information about [GCPRole](/docs/concepts/secret-engine-crds/gcprole.md) and subjects.

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

### spec.roleRef

`spec.roleRef` is a required field that specifies the [GCPRole](/docs/concepts/secret-engine-crds/gcprole.md) against which credential will be issued.

```yaml
spec:
  roleRef:
    name: gcp-role
    namespace: demo
```

It has following field:

- `roleRef.name` : `Required`. Specifies the name of the object being referenced.
- `roleRef.namespace` : `Required`. Specifies the namespace of the referenced object.

### spec.subjects

`spec.subjects` is a required field that contains a list of reference to the object or user identity a role binding applies to. It will have read access of the credential secret. This can either hold a direct API object reference, or a value for non-objects such as user and group names.

```yaml
spec:
  subjects:
    - kind: ServiceAccount
      name: sa
      namespace: demo
```

### spec.keyAlgorithm 

Specifies the key algorithm used to generate key. Defaults to 2k RSA key You probably should not choose other values (i.e. 1k), but accepted values are `KEY_ALG_UNSPECIFIED`, `KEY_ALG_RSA_1024`, `KEY_ALG_RSA_2048`  

```yaml
spec:
  keyAlgorithm: KEY_ALG_RSA_2048
```

### spec.keyType

Specifies the private key type to generate. Defaults to JSON credentials file. Accepted values are `TYPE_UNSPECIFIED`, `TYPE_GOOGLE_CREDENTIALS_FILE`

```yaml
spec:
  keyType: TYPE_GOOGLE_CREDENTIALS_FILE
``` 

## GCPAccessKeyRequest Status

`status` shows the status of the GCPAccessKeyRequest. It is maintained by Vault operator. It contains following fields:

- `secret` : Specifies the name of the secret containing GCP credential.

- `lease` : Contains lease information of the issued credential.

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

> Note: GCP credential will be issued if `conditions[].type` is `Approved`. Otherwise, Vault operator will not issue any credential.
