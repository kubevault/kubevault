---
title: AzureRole | Vault Secret Engine
menu:
  docs_0.2.0:
    identifier: azurerole-secret-engine-crds
    name: AzureRole
    parent: secret-engine-crds-concepts
    weight: 10
menu_name: docs_0.2.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# AzureRole CRD

Most secrets engines must be configured in advance before they can perform their functions. When a AzureRole CRD is created, the vault operator will perform the following operations:

- [Enable](https://www.vaultproject.io/docs/secrets/azure/index.html#setup) the Vault Azure secret engine if it is not already enabled
- [Configure](https://www.vaultproject.io/api/secret/azure/index.html#configure-access) Vault Azure secret engine
- [Create](https://www.vaultproject.io/api/secret/azure/index.html#create-update-role) role according to `AzureRole` CRD specification


```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AzureRole
metadata:
  name: <role-name>
  namespace: <role-namespace>
spec:
  ref:
    name: <appbinding-name>
    namespace: <appbinding-namespace>
  applicationObjectID: <application-object-id>
  azureRoles: <list-of-azure-roles>
  config:
    credentialSecret: <azure-secret-name>
    environment: <azure-environment>  
  ttl: <ttl-time>
  maxTTL: <max-ttl-time>
status: ...
```

## AzureRole Spec

AzureRole `spec` contains information which will be required to enable and configure azure secret engine and finally create azure role.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AzureRole
metadata:
  name: demo-role
  namespace: demo
spec:
  ref:
    name: vault-app
    namespace: demo
  applicationObjectID: c1cb042d-96d7-423a-8dba-243c2e5010d3
  config:
    credentialSecret: azure-cred
    environment: AzurePublicCloud
  ttl: 1h
  maxTTL: 1h
```

### spec.ref

`spec.ref` specifies the name and namespace of [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains information to communicate with Vault.

```yaml
spec:
  ref:
    name: vault-app
    namespace: demo
```
### spec.azureRoles

List of Azure roles to be assigned to the generated service principal. The array must be in JSON format, properly escaped as a string. 

```yaml
spec:
  azureRoles: `[
                       {
                           "role_name": "Contributor",
                           "scope":  "/subscriptions/<uuid>/resourceGroups/Website"
                       }
                   ]`
```

### spec.applicationObjectID

Application Object ID for an existing service principal that will be used instead of creating dynamic service principals. If present, azure_roles will be ignored. See [roles docs](https://www.vaultproject.io/docs/secrets/azure/index.html#roles) for details on role definition.

```yaml
spec:
  applicationObjectID: c1cb042d-96d7-423a-8dba-243c2e5010d3
```
### spec.config

`spec.config` is a required field that contains [information](https://www.vaultproject.io/api/secret/azure/index.html#configure-access) to communicate with Azure. It has the following fields:
- **credentialSecret**: `required`, Specifies the secret name containing azure credentials
    - **subscription-id**: `required`, The subscription id for the Azure Active Directory stored in `data["subscription-id"]=<value>` 
    - **tenant-id**: `required`, The tenant id for the Azure Active Directory stored in `data["tenant-id"]=<value>`
    - **client-id**: `optional`, The OAuth2 client id to connect to Azure stored in `data["client-id"]=<value>`
    - **client-secret**: `optional`, The OAuth2 client secret to connect to Azure stored in `data["client-secret"]=<value>`
- **environment**: `optional`, The Azure environment. If not specified, Vault will use Azure Public Cloud.

```yaml
spec:
  config:
    credentialSecret: azure-cred
    environment: AzurePublicCloud
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: azure-cred
  namespace: demo
data:
  client-secret: TU1hRjdRZWVzTGZx******
  subscription-id: MWJmYzlmNjYtMzE*****
  client-id: MmI4NzFkNGEtNzU3**********
  tenant-id: NzcyMjY4ZTUtZDk***********
```

### spec.ttl

Specifies the default TTL for service principals generated using this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to the system/engine default TTL time.

```yaml
spec:
  ttl: 1h
```

### spec.maxTTL

Specifies the maximum TTL for service principals generated using this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to the system/engine max TTL time.

```yaml
spec:
  maxTTL: 1h
```

## AzureRole Status

`status` shows the status of the AzureRole. It is maintained by Vault operator. It contains following fields:

- `phase` : Indicates whether the role successfully applied in vault or not or in progress or failed

- `conditions` : Represent observations of a AzureRole.
