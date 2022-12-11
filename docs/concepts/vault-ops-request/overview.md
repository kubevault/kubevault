---
title: Vault Ops Request Overview | KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: overview-ops-request-concepts
    name: Overview
    parent: ops-request-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/README.md).

# VaultOpsRequest

## What is VaultOpsRequest

`VaultOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for `Vault` administrative operations like restart, reconfigure TLS etc. in a Kubernetes native way.

## VaultOpsRequest CRD Specifications

Like any official Kubernetes resource, a `VaultOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and Status sections.

Here, some sample `VaultOpsRequest` CRs for different administrative operations is given below:

### Sample `VaultOpsRequest` for restarting `VaultServer`:

```yaml
apiVersion: ops.kubevault.com/v1alpha1
kind: VaultOpsRequest
metadata:
  name: restart-vault-request
  namespace: demo
spec:
  restart: {}
  type: Restart
  vaultRef:
    name: vault
status:
  conditions:
  - lastTransitionTime: "2022-12-04T09:27:07Z"
    message: Vault ops request is restarting nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2022-12-04T09:29:23Z"
    message: Successfully restarted all nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2022-12-04T09:29:23Z"
    message: Successfully completed the modification process.
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful

```

### Sample `VaultOpsRequest` Objects for Reconfiguring TLS of the `VaultServer`:

```yaml
apiVersion: ops.kubevault.com/v1alpha1
kind: VaultOpsRequest
metadata:
  name: vault-ops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  vaultRef:
    name: vault
  tls:
    issuerRef:
      name: vault-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        emailAddresses:
          - abc@appscode.com

```

```yaml
apiVersion: ops.kubevault.com/v1alpha1
kind: VaultOpsRequest
metadata:
  name: vault-ops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  vaultRef:
    name: vault
  tls:
    rotateCertificates: true

```

```yaml
apiVersion: ops.kubevault.com/v1alpha1
kind: VaultOpsRequest
metadata:
  name: vault-ops-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  vaultRef:
    name: vault
  tls:
    issuerRef:
      name: vault-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"

```

```yaml
apiVersion: ops.kubevault.com/v1alpha1
kind: VaultOpsRequest
metadata:
  name: vault-ops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  vaultRef:
    name: vault
  tls:
    remove: true

```

Here, we are going to describe the various sections of a `VaultOpsRequest` crd.

A `VaultOpsRequest` object has the following fields in the spec section.

### spec.vaultRef

`spec.vaultRef` is a required field that point to the `Vault` object for which the administrative operations will be performed. This field consists of the following sub-field:
- `spec.databaseRef.name` : specifies the name of the `Vault` object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the `VaultServer`. Currently, the following types of operations are allowed in `VaultOpsRequest`.
- Restart
- ReconfigureTLS

You can perform only one type of operation on a single `VaultOpsRequest` CR. You should not create two `VaultOpsRequest` simultaneously.

### spec.tls

If you want to reconfigure the TLS configuration of your `VaultServer` i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this `VaultServer`.
- `spec.tls.remove` specifies that we want to remove tls from this `VaultServer`.

### VaultOpsRequest `Status`

`.status` describes the current state and progress of a `VautlOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `VaultOpsRequest`. It can have the following three values:

| Phase      | Meaning                                                                             |
| ---------- |-------------------------------------------------------------------------------------|
| Successful | KubeVault has successfully performed the operation requested in the VaultOpsRequest |
| Failed     | KubeVault has failed the operation requested in the VaultOpsRequest                 |
| Denied     | KubeVault has denied the operation requested in the VaultOpsRequest                 |

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `VaultOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `VaultOpsRequest` processing. Each condition entry has the following fields:
- `types` specifies the type of the condition. `VaultOpsRequest` has the following types of conditions:

| Type                         | Meaning                                                                       |
|------------------------------|-------------------------------------------------------------------------------|
| `Progressing`                | Specifies that the operation is now in the progressing state                  |
| `Successful`                 | Specifies such a state that the operation on the vault was successful.        |
| `ResumeVaultServer`          | Specifies such a state that the vault is resumed by the operator              |
| `Failed`                     | Specifies such a state that the operation on the database failed.             |
| `UpdateStatefulSetResources` | Specifies such a state that the Statefulset resources has been updated        |
| `RestartNodes`               | Specifies whether the vault nodes has been restarted or not                   |
| `CertificateSynced`          | Specifies whether the certificates has been synced across all the vault nodes |


- The `status` field is a string, with possible values `True`, `False`, and `Unknown`.
  - `status` will be `True` if the current transition succeeded.
  - `status` will be `False` if the current transition failed.
  - `status` will be `Unknown` if the current transition was denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.
