---
title: GCPRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: gcprole-secret-engine-crds
    name: GCPRole
    parent: secret-engine-crds-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# GCPRole CRD

Most secrets engines must be configured in advance before they can perform their functions. When a GCPRole CRD is created, the vault operator will perform the following operations:

- [Enable](https://www.vaultproject.io/docs/secrets/gcp/index.html#setup) the Vault GCP secret engine if it is not already enabled
- [Configure](https://www.vaultproject.io/api/secret/gcp/index.html#write-config) Vault GCP secret engine
- [Create](https://www.vaultproject.io/api/secret/gcp/index.html#create-update-roleset) roleset according to `GCPRole` CRD specification

For maintaining similarity with other secret engines we will refer **roleset as role** in the following description.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: GCPRole
metadata:
  name: <name>
  namespace: <namespace>
spec:
  ref:
    name: <appbinding-name>
    namespace: <appbinding-namespace>
  bindings: <binding-configuration-string>
  config:
    credentialSecret: <gcp-credential-secret-name>
  project: <project-name>
  secretType: <secret-type>
  tokenScopes: <list-of-OAuth-scopes>
status: ...
```

## GCPRole Spec

GCPRole `spec` contains information which will be required to enable gcp secret engine, configure gcp secret engine and create gcp role.

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

### spec.ref

`spec.ref` specifies the name and namespace of [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains information to communicate with Vault.

```yaml
spec:
  ref:
    name: vault-app
    namespace: demo
```

### spec.config

`spec.config` is a required field that contains [information](https://www.vaultproject.io/api/secret/gcp/index.html#parameters) to communicate with GCP. It has the following fields:

- **credentialSecret**: `Required`, Specifies the secret name that contains google application credentials in `data["sa.json"]=<value.json>`
- **ttl**: `optional`, Specifies default config TTL for long-lived credentials (i.e. service account keys). Default value is 0s.
- **maxTTL**: `optional`, Specifies the maximum config TTL for long-lived credentials (i.e. service account keys). Default value is 0s.

```yaml
spec:
  config:
    credentialSecret: gcp-cred
    ttl: 0s
    maxTTL: 0s
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gcp-cred
  namespace: demo
data:
  sa.json: ewogICJ0eXBlIjogInNlcnZpY2VfYWNjb3VudAp9.............
```

### spec.secretType

Specifies the type of secret generated for this role set. Accepted values: `access_token`, `service_account_key`.

```yaml
spec:
  secretType: access_token
```

### spec.project

Specifies th name of the GCP project that this roleset's service account will belong to.

```yaml
spec:
  project: ackube
```

### spec.bindings

Specifies bindings configuration string.

```yaml
spec:
  bindings: 'resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
    roles = ["roles/viewer"]
    }'
```

### spec.tokenScopes

Specifies the list of OAuth scopes to assign to `access_token` secrets generated under this role set (`access_token` role sets only)

```yaml
spec:
  tokenScopes: ["https://www.googleapis.com/auth/cloud-platform"]
```

## GCPRole Status

`status` shows the status of the GCPRole. It is maintained by Vault operator. It contains following fields:

- `phase` : Indicates whether the role successfully applied in vault or not or in progress or failed

- `conditions` : Represent observations of a GCPRole.
