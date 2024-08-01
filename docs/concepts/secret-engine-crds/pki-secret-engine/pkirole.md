---
title: PKIRole | Vault Secret Engine
menu:
  docs_{{ .version }}:
    identifier: pkirole-secret-engine-crds
    name: PKIRole
    parent: pki-crds-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# PKIRole

## What is PKIRole

An `PKIRole` is a Kubernetes `CustomResourceDefinition` (CRD) which allows a user to create PKI secret engine role in a Kubernetes native way.

When an `PKIRole` is created, the KubeVault operator [configures](https://www.vaultproject.io/docs/secrets/pki/index.html#setup) a Vault role that maps to a set of permissions in PKI as well as an PKI credential type. When users generate credentials, they are generated against this role. If the user deletes the `PKIRole` CRD,
then the respective role will also be deleted from Vault.

![PKIRole CRD](/docs/images/concepts/pki_role.svg)

## PKIRole CRD Specification

Like any official Kubernetes resource, a `PKIRole` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `PKIRole` object is shown below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: PKIRole
metadata:
  name: pki-role
  namespace: demo
spec:
  secretEngineRef:
    name: pki-secret-engine
  allowedDomains:
    - "kubevault.com"
  allowSubdomains: true
  maxTTL: "720h"
  additionalPayload:
    "allow_ip_sans": "true"
```

> Note: To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`

Here, we are going to describe the various sections of the `PKIRole` crd.

### PKIRole Spec

PKIRole `spec` contains role information.

```yaml
spec:
  secretEngineRef:
    name: <secret-engine-name>
  allowedDomains: <allowed domain names>
  allowSubdomains: <true>
  defaultTTL: <default-TTL>
  maxTTL: <max-TTL>
  additionalPayload:
    "key": "value"
```

`PKIRole` spec has the following fields:

#### spec.secretEngineRef

`spec.secretEngineRef` is a `required` field that specifies the name of a `SecretEngine`.

```yaml
spec:
  secretEngineRef:
    name: pki-secret-engine
```

#### spec.allowedDomains

`spec.allowedDomains` is a `required` field that specifies the domains this role is allowed to issue certificates for

```yaml
spec:
  allowedDomains:
    - "kubevault.com"
```

#### spec.allowSubdomains

`spec.allowSubdomains` is an `optional` field that specifies the if subdomains is allowed.

```yaml
spec:
  allowSubdomains: true
```

#### spec.defaultTTL

`spec.defaultTTL` is an `optional` field that specifies the default TTL for certificates.

```yaml
spec:
  maxTTL: "1h"
```

#### spec.maxTTL

`spec.maxTTL` is an `optional` field that specifies the max allowed TTL for certificates. 

```yaml
spec:
  maxTTL: "1h"
```

#### spec.additionalPayload

`spec.additionalPayload` is an `optional` field which can used to provide any key value of [vault-api](https://developer.hashicorp.com/vault/api-docs/secret/pki#create-update-role)
which will be used to create the role.

```yaml
spec:
  additionalPayload:
    "key1": "value1"
    "key2": "value2"
  
```

### PKIRole Status

`status` shows the status of the PKIRole. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation, which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully applied to Vault or not.

- `conditions` : Represent observations of an PKIRole.
