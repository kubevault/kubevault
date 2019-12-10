---
title: AzureRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: azurerole-secret-engine-crds
    name: AzureRole
    parent: azure-crds-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# AzureRole 

## What is AzureRole

An `AzureRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create an Azure secret engine role in a Kubernetes native way.

When an `AzureRole` is created, the KubeVault operator [configures](https://www.vaultproject.io/docs/secrets/azure/index.html#setup) a Vault role.
A role may be set up with either an existing service principal or a set of Azure roles that will be assigned to a dynamically created service principal.
If the user deletes the `AzureRole` CRD, then the respective role will also be deleted from Vault.

![AzureRole CRD](/docs/images/concepts/azure_role.svg)

## AzureRole CRD Specification

Like any official Kubernetes resource, a `AzureRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `AzureRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AzureRole
metadata:
  name: azure-role
  namespace: demo
spec:
  vaultRef:
    name: vault-app
  azureRoles: `[
              {
                "role_name": "Contributor",
                "scope":  "/subscriptions/<uuid>/resourceGroups/Website"
            }
          ]`
  applicationObjectID: c1cb042d-96d7-423a-8dba-243c2e5010d3
status:
  observedGeneration: 1
  phase: Success
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `AzureRole` crd.

### AzureRole Spec

AzureRole `spec` contains either new service principal configuration or existing service principal name required for configuring a role.

```yaml
spec:
  vaultRef:
    name: <appbinding-name>
  path: <azure-secret-engine-path>
  applicationObjectID: <existing-application-object-id>
  azureRoles: <list-of-azure-roles>
  ttl: <default-ttl>
  maxTTL: <max-ttl>
```

`AzureRole` spec has the following fields:

#### spec.vaultRef

`spec.vaultRef` is a `required` field that specifies the name of an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) reference which is used to connect with a Vault server. AppBinding must be in the same namespace with the AzureRole object.

```yaml
spec:
  vaultRef:
    name: vault-app
```

#### spec.path

`spec.path` is an `optional` field that specifies the path where the secret engine is enabled. The default path value is `azure`.

```yaml
spec:
  path: my-azure-path
```

#### spec.azureRoles

`spec.azureRoles` is an `optional` field that specifies a list of Azure roles to be assigned to the generated service principal. The array must be in JSON format, properly escaped as a string.

```yaml
spec:
  azureRoles: `[
                 {
                    "role_name": "Contributor",
                    "scope":  "/subscriptions/<uuid>/resourceGroups/Website"
                }
              ]`
```

#### spec.applicationObjectID

`spec.applicationObjectID` is an `optional` field that specifies  the Application Object ID for an existing service principal that will be used instead of creating dynamic service principals. If present, azure_roles will be ignored. See [roles docs](https://www.vaultproject.io/docs/secrets/azure/index.html#roles) for details on role definition.

```yaml
spec:
  applicationObjectID: c1cb042d-96d7-423a-8dba-243c2e5010d3
```

#### spec.ttl

Specifies the default TTL for service principals generated using this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to the system/engine default TTL time.

```yaml
spec:
  ttl: 1h
```

#### spec.maxTTL

Specifies the maximum TTL for service principals generated using this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to the system/engine max TTL time.

```yaml
spec:
  maxTTL: 1h
```

### AzureRole Status

`status` shows the status of the AzureRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation, which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of an AzureRole.
