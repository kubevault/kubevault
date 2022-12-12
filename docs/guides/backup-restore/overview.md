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

Your `VaultServer` must be initialised & unsealed for Backup & Restore process to work. This will take the snapshot
using a consistent mode that forwards the request to the cluster leader, and the leader will verify it is still
in power before taking the snapshot.

`Raft` storage backend stanza for `KubeVault` may look like this:

```yaml
backend:
  raft:
    storage:
      resources:
        requests:
          storage: 1Gi
      storageClassName: standard
```

This `Storage Backend` information is available in the `AppBinding` created by the operator.

```yaml
spec:
  parameters:
    backend: raft
```

`AppBinding` has the information about the Unsealer option of the `VaultServer`. During the Backup,
Vault `unseal-keys` & `root-token` will also be backed-up for the completeness of the Backup process. `AppBinding` section for Unsealer option `GoogleKmsGcs` may look like this:

```yaml
unsealer:
  mode:
    googleKmsGcs:
      bucket: vault-testing-keys
      credentialSecretRef:
        name: gcp-cred
      kmsCryptoKey: vault-testing-key
      kmsKeyRing: vault-testing
      kmsLocation: global
      kmsProject: appscode-testing
  secretShares: 5
  secretThreshold: 3
```

`spec.parameters.stash` section contains the stash parameters for Backup & Restore tasks. 
`spec.parameters.stash.addon` contains the information about the `Task` for backup & restore. 
It also contains the `params` which indicates the `keyPrefix` that is prepended with the name of vault 
`unseal-keys` & `root-token`, e.g. `k8s.kubevault.com.demo.vault-root-token`, 
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


