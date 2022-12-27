---
title: Vault Backup Restore Overview
menu:
  docs_{{ .version }}:
    identifier: overview-backup-restore-guides
    name: Overview
    parent: backup-restore-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Vault Backup Restore Overview

`Vault` provides a set of standard operating procedures `(SOP)` for backing up a Vault cluster. It protects your Vault cluster
against data corruption or sabotage of which Disaster Recovery Replication might not be able to protect against.

`KubeVault` supports number of different [Storage Backend](/docs/concepts/vault-server-crds/storage/overview.md) types. Therefore, the exact steps to backup Vault will depend on your
selected storage backend. The two recommended storage backend types are Consul and Integrated Storage 
(also known as Raft). 

`KubeVault` currently supports Backup & Restore for `Raft` storage backend. So, this document assumes that `Raft` storage backend is being used. 

## Backup & Restore process for Raft

Your `VaultServer` must be in `Ready` state for Backup & Restore process to work. This will take the snapshot
using a consistent mode that forwards the request to the cluster leader, and the leader will verify it is still
in power before taking the snapshot.

A simple `VaultServer` YAML with `Raft` storage backend may look like this:

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: 1.10.3
  replicas: 3
  allowedSecretEngines:
    namespaces:
      from: All
  backend:
    raft:
      storage:
        storageClassName: "standard"
        resources:
          requests:
            storage: 1Gi
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```

Let's take a look at some relevant fields:

`spec.backend` contains the Backend storage information, `Raft` in this case:

```yaml
backend:
  raft:
    storage:
      resources:
        requests:
          storage: 1Gi
      storageClassName: standard
```

`spec.unsealer` contains `VaultServer` unsealing option. In this case which is `Kubernetes` Secret. So, on Vault deployment
a Secret will be created on the same namespace, which will create the Vault unseal-keys & root-token.

```yaml
unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
```

`KubeVault` operator will create an `AppBinding` with all the necessary information for backup & restore.
`AppBinding` has the information about the Unsealer option of the `VaultServer`. During the Backup,
Vault `unseal-keys` & `root-token` will also be backed-up for the completeness of the Backup process. 

`KubeVault` created `AppBinding` YAML may look like this:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault
  namespace: demo
spec:
  appRef:
    apiGroup: kubevault.com
    kind: VaultServer
    name: vault
    namespace: demo
  clientConfig:
    service:
      name: vault
      port: 8200
      scheme: http
  parameters:
    apiVersion: config.kubevault.com/v1alpha1
    backend: raft
    backupTokenSecretRef:
      name: vault-backup-token
    kind: VaultServerConfiguration
    kubernetes:
      serviceAccountName: vault
      tokenReviewerServiceAccountName: vault-k8s-token-reviewer
      usePodServiceAccountForCSIDriver: true
    path: kubernetes
    stash:
      addon:
        backupTask:
          name: vault-backup-1.10.3
          params:
            - name: keyPrefix
              value: k8s.kubevault.com.demo.vault
        restoreTask:
          name: vault-restore-1.10.3
          params:
            - name: keyPrefix
              value: k8s.kubevault.com.demo.vault
    unsealer:
      mode:
        kubernetesSecret:
          secretName: vault-keys
      secretShares: 5
      secretThreshold: 3
    vaultRole: vault-policy-controller
```

Read more about `AppBinding` [here](/docs/concepts/vault-server-crds/appbinding.md).

Here:
- `spec.parameters.stash` section contains the stash parameters for Backup & Restore tasks. 
- `spec.parameters.stash.addon` contains the information about the `Task` for backup & restore. 
It also contains the `params` which indicates the `keyPrefix` that is prepended with the name of vault `unseal-keys` & `root-token`, e.g. `k8s.kubevault.com.demo.vault-root-token`, 
`k8s.kubevault.com.demo.vault-root-token-unseal-key-0`, `k8s.kubevault.com.demo.vault-root-token-unseal-key-1`, etc.

```yaml
stash:
  addon:
    backupTask:
      name: vault-backup-1.10.3
      params:
      - name: keyPrefix
        value: k8s.kubevault.com.demo.vault
    restoreTask:
      name: vault-restore-1.10.3
      params:
      - name: keyPrefix
        value: k8s.kubevault.com.demo.vault
```


`KubeVault` operator will create a `K8s Secret` containing a `token` during the Vault deployment, which contains the necessary permission
for the Backup & Restore process. This information is available in the `AppBinding` created by the operator. AppBinding
`spec.parameters.backupTokenSecretRef` contains the reference of that secret.

```yaml
spec:
  parameters:
    backupTokenSecretRef:
      name: vault-backup-token

```

A sample policy document / permission may look like this:

```hcl
path "sys/storage/raft/snapshot" {
        capabilities = ["read"]
}

path "sys/storage/raft/snapshot-force" {
        capabilities = ["read"]
}
```

If your `Vault` deployment isn't managed by `KubeVault`, then you'll need to create the `AppBinding` & `Secret` containing 
the permissions required for backup & restore separately.

Up next:
- Read about step-by-step Backup procedure [here](/docs/guides/backup-restore/backup.md)
- Read about step-by-step Restore procedure [here](/docs/guides/backup-restore/restore.md)

