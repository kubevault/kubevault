# VaultPolicyBinding CRD

Vault operator will create Vault Kuberenetes [Role](https://www.vaultproject.io/api/auth/kubernetes/index.html#create-role) according to `VaultPolicyBinding` CRD (CustomResourceDefinition) specification. If the user deletes the VaultPolicyBinding CRD, then respective role will also be deleted from Vault.

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: <name>
  namespace: <namespace>
spec:
  ...
status:
  ...
```

## VaultPolicyBinding Spec

VaultPolicyBinding `spec` contains information that necessary for creating Vault Kubernetes Role.

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: secret-admin
  namespace: demo
spec:
  policies: ["secret-admin"]
  serviceAccountNames: ["sa1","sa2"]
  serviceAccountNamespaces: ["default","demo"]
  TTL: "1000"
  maxTTL: "2000"
  Period: "1000"
```

VaultPolicyBinding Spec has following fields:

### spec.roleName

`spec.roleName` is an optional field that specifies the name of the Vault Kubernetes role.

```yaml
spec:
  roleName: demo
```

> Note: If `spec.roleName` is not specified, then the name of role in Vault will follow this format: `k8s.{spec.clusterName}.{spec.namespace}.{spec.name}`

### spec.authPath

`spec.authPath` is an optional field that specifies the path where kubernetes auth is enabled. Default value is `kubernetes`.

```yaml
spec:
  authPath: k8s
```

### spec.policies

`spec.policies` is a required field that specifies the list of [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) CRD names. These `VaultPolicy` CRD should be in the namespace of the `VaultPolicyBinding` CRD. 

```yaml
spec:
  policies: ["secret-admin"]
```

### spec.serviceAccountNames

`spec.serviceAccountNames` is a required field that specifies the list of service account names. They will have the access to use this role.

```yaml
spec:
  serviceAccountNames: ["sa1","sa2"]
```

### spec.serviceAccountNamespaces

`spec.serviceAccountNamespaces` is a required field that specifies the list of namespaces where `spce.serviceAccountNames` are in. 

```yaml
spec:
  serviceAccountNamespaces: ["demo","default"]
```

### spec.TTL

`spec.TTL` is an optional field that specifies the TTL period of the token issued using this role in seconds.

```yaml
spec:
  TTL: "300"
```

### spec.maxTTL

`spec.maxTTL` is an optional field that specifies the maximum allowed lifetime of the token issued in seconds using this role.

```yaml
spec:
  maxTTL: "300"
```

### spec.period

`spec.period` is an optional field. If set, indicates that the token generated using this role should never expire. The token should be renewed within the duration specified by this value. At each renewal, the token's TTL will be set to the value of this field.

```yaml
spec:
  period: "300"
```

## VaultPolicyBinding Status

`status` shows the status of VaultPolicyBinding. It is maintained by Vault operator. It contains following fields:

- `status` : Indicates whether the role successfully created in Vault or not or in progress or failed.

- `conditions` : Represent observations of a VaultPolicyBinding.
