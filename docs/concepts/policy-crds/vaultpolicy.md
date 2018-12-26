# VaultPolicy CRD

Vault operator will create Vault [Policy](https://www.vaultproject.io/docs/concepts/policies.html) according to `VaultPolicy` CRD (CustomResourceDefinition) specification. If the user deletes the VaultPolicy CRD, then respective policy will also be deleted from Vault.

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: <name>
  namespace: <namespace>
spec:
  ...
status:
  ...
```

> Note: To resolve the naming conflict, name of policy in Vault will follow this format: `k8s.{spec.clusterName}.{spec.namespace}.{spec.name}`

## VaultPolicy Spec

VaultPolicy `spec` contains policy and vault information.

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: secret-admin
  namespace: demo
spec:
  vaultAppRef:
    name: vault
    namespace: demo
  policy: |
    path "secret/*" {
      capabilities = ["create", "read", "update", "delete", "list"]
    }
```

VaultPolicy Spec has following fields:

### spec.policy

`spec.policy` is a required field that specifies the vault policy in hcl format.

```yaml
spec:
  policy: |
      path "secret/*" {
        capabilities = ["create", "read", "update", "delete", "list"]
      }
```

### spec.vaultAppRef

`spec.vaultAppRef` is a required field that specifies name and namespace of [AppBinding](/docs/concepts/appbinding-crds/appbinding.md) that contains information to communicate with Vault.

```yaml
spec:
  vaultAppRef:
    name: vault
    namespace: demo
```

## VaultPolicy Status

VaultPolicy `status` shows the status of Vault Policy. It is maintained by Vault operator. It contains following fields:

- `status` : Indicates whether the policy successfully applied in vault or not or in progress or failed

- `conditions` : Represent observations of a VaultPolicy.
