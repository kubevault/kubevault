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

