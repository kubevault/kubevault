---
title: Raft | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: raft-storage
    name: Raft
    parent: storage-vault-server-crds
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Raft

In the `Raft` storage backend, vault data will be stored in provided file system path. Vault documentation for `Raft` storage backend can be found in [here](https://www.vaultproject.io/docs/configuration/storage/raft.html).

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault
  namespace: default
spec:
  replicas: 3
  version: 1.7.3
  serviceTemplates:
    - alias: vault
      metadata:
        annotations:
          name: vault
      spec:
        type: NodePort
    - alias: stats
      spec:
        type: ClusterIP
  backend:
    raft:
      path: "/vault/data"
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
  monitor:
    agent: prometheus.io
    prometheus:
      exporter:
        resources: {}
  terminationPolicy: WipeOut

```

## spec.backend.raft

To use `Raft` as backend storage in Vault, we need to specify `spec.backend.raft` in [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.
More information about the `Raft` backend storage can be found in [here](https://www.vaultproject.io/docs/configuration/storage/raft.html)

```yaml
spec:
  backend:
    raft:
      path: <filesystem_path_to_store_data>
      performanceMultiplier: <integer_multiplier_to_scale_timing_parameters>
      trailingLogs: <number_of_log_entries_left>
      snapshotThreshold: <minimum_number_of_commit_entries>
      maxEntrySize: <max_number_of_bytes_for_raft_entry>
      autoPilotReconcileInterval: <interval_autopilot_needs_to_pick_up_state_chyanges>
```

Here, we are going to describe the various attributes of the `spec.backend.raft` field.

### raft.path

`Path` specifies the filesystem path where the vault data gets stored. This value can be overridden by setting the `VAULT_RAFT_PATH` environment variable. `default: ""`

```yaml
spec:
  backend:
    raft:
      path: "/vault/data"
```

### raft.performanceMultiplier

An integer multiplier used by servers to scale key Raft timing parameters. Tuning this affects the time it takes Vault to detect leader failures and to perform leader elections, at the expense of requiring more network and CPU resources for better performance. `default: 0`
```yaml
spec:
  backend:
    raft:
      performanceMultiplier: 0
```

### raft.trailingLogs

This controls how many log entries are left in the log store on disk after a snapshot is made. `default: 10000`
```yaml
spec:
  backend:
    raft:
      trailingLogs: 10000
```

### raft.snapshotThreshold

This controls the minimum number of raft commit entries between snapshots that are saved to disk. `default: 8192`
```yaml
spec:
  backend:
    raft:
      snapshotThreshold: 8192
```

### raft.maxEntrySize

This configures the maximum number of bytes for a raft entry. It applies to both Put operations and transactions. `default: 1048576`
```yaml
spec:
  backend:
    raft:
      maxEntrySize: 1048576
```

### raft.autoPilotReconcileInterval

This is the interval after which autopilot will pick up any state changes. `default: ""`
```yaml
spec:
  backend:
    raft:
      autoPilotReconcileInterval: ""
```