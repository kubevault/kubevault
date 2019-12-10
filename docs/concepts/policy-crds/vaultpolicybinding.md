---
title: VaultPolicyBinding | KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: vaultpolicybinding-policy-crds
    name: VaultPolicyBinding
    parent: policy-crds-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# VaultPolicyBinding

## What is VaultPolicyBinding

A `VaultPolicyBinding` is a Kubernetes `CustomResourceDefinition` (CRD) which binds Vault server [policies](https://www.vaultproject.io/docs/concepts/policies.html) to an auth method role in a Kubernetes native way.

When a `VaultPolicyBinding` is created, the KubeVault operator will create an auth role according to CRD (CustomResourceDefinition) specification.
If the user deletes the VaultPolicyBinding CRD, then the respective role will also be deleted from Vault.

![VaultPolicyBinding CRD](/docs/images/concepts/vault_policy_binding.svg)

Auth method roles are associated with an authentication type/entity and a set of Vault policies. Currently supported auth methods for VaultPolicyBinding:

- [Kubernetes Auth Method](https://www.vaultproject.io/docs/auth/kubernetes.html): The Kubernetes auth method can be used to authenticate with Vault using a Kubernetes Service Account Token. This method of authentication makes it easy to introduce a Vault token into a Kubernetes Pod.

## VaultPolicyBinding CRD Specification

Like any official Kubernetes resource, a `VaultPolicyBinding` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `VaultPolicyBinding` object is shown below:

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: admin-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  policies:
  - ref: secret-admin
  subjectRef:
    kubernetes:
      serviceAccountNames:
        - "demo sa"
      serviceAccountNamespaces:
        - "demo"
      ttl: "1000"
      maxTTL: "2000"
      period: "1000"
status:
  observedGeneration: 1
  phase: Success
```

Here, we are going to describe the various sections of the `VaultPolicyBinding` crd.

### VaultPolicyBinding Spec

VaultPolicyBinding `spec` contains information that is necessary for creating an auth role.

```yaml
spec:
  vaultRef:
    name: <vault-appbinding-name>
  vaultRoleName: <role-name>
  policies:
  - name: <vault-policy-name>
    ref:  <VaultPolicy-crd-name>
  subjectRef:
    kubernetes:
      path: <k8s-auth-path>
      serviceAccountNames:
      - "sa1"
      - "sa2"
      serviceAccountNamespaces:
      - "ns1"
      ttl: <token-ttl>
      maxTTL: <token-maxTTL>
      period: <token-period>
```

VaultPolicyBinding spec has the following fields:

#### spec.vaultRef

`spec.vaultRef` is a `required` field that specifies the name of an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains information to communicate with a Vault server. The AppBinding object must be in the same namespace with VaultPolicyBinding object.

```yaml
spec:
  vaultRef:
    name: vault-app
```

#### spec.vaultRoleName

To avoid naming conflict, KubeVault operator will generate role names in Vault server in this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`. `spec.vaultRoleName`  is an `optional` field. If set, it will be used instead of the auto-generated role name.

```yaml
spec:
  vaultRoleName: my-custom-role
```

#### spec.policies

`spec.policies` is a `required` field that specifies a list of vault policy references. Each item of the list
can be **either** a vault policy name **or** VaultPolicy CRD name.

- `name`: Specifies the [vault policy](https://www.vaultproject.io/docs/concepts/policies.html) name.
   This name should be returned by `vault read sys/policy` command.

- `ref`: Specifies the name of [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) crd object. The KubeVault operator will get the vault policy name
   from the crd object.

```yaml
spec:
  policies:
  - name: policy1
  - ref: policy2
```

#### spec.subjectRef

`spec.subjectRef` is a `required` field that specifies the reference of vault users who will be granted
token with mentioned policies.

- `kubernetes`: Refers to vault users who will be authenticated via the Kubernetes auth method.

  - `path` : `Optional`. Specifies the path where the Kubernetes auth is enabled. The default value is `kubernetes`.

  - `serviceAccountNames` : `Required`. Specifies the list of service account names.
        They will have access to use this role.  If set to `"*"` all names are allowed,
        both this and serviceAccountNamespaces **cannot** be `"*"`.

  - `serviceAccountNamespaces` : `Required`. Specifies a list of namespaces allowed to access this role. This value set to "*" means
     all namespaces are allowed.

  - `ttl` : `Optional`. Specifies the TTL period of the token issued using this role in seconds. Default value "0".

  - `maxTTL` : `Optional`. Specifies the maximum allowed lifetime of tokens issued in seconds using this role.

  - `period` : `Optional`. If set indicates that the token generated using this role should never expire. The token should be renewed within the
     duration specified by this value. At each renewal, the token's TTL will be set to the value of this parameter.

```yaml
spec:
  subjectRef:
    kubernetes:
      serviceAccountNames:
      - "sa1"
      - "sa2"
      serviceAccountNamespaces:
      - "demo"
      ttl: "1000"
      maxTTL: "2000"
      period: "1000"
```

### VaultPolicyBinding Status

`status` shows the status of a VaultPolicyBinding. It is managed by the KubeVault operator. It contains the following fields:

- `observedGeneration`: Specifies the most recent generation observed for this resource. It corresponds to the resource's generation, which is updated on mutation by the API Server.

- `phase`: Indicates whether the role successfully created in the Vault or not.

- `conditions` : Represents the latest available observations of a VaultPolicyBinding's current state.
