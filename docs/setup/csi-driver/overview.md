---
title: Overview of Vault CSI Driver
menu:
  docs_{{ .version }}:
    identifier: overview-csi-driver
    name: Overview
    parent: csi-driver-setup
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: setup
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

## Prerequisite

- Kubernetes v1.14+
- `--allow-privileged` flag must be set to true for both the API server and the kubelet
- (If you use Docker) The Docker daemon of the cluster nodes must allow shared mounts
- Pre-installed HasiCorp Vault server.
- Pass `--feature-gates=CSIDriverRegistry=true,CSINodeInfo=true` to kubelet and kube-apiserver


## Supported [CSI Spec](https://github.com/container-storage-interface/spec) version

| CSI Spec Version | csi-vault:0.1.0 | csi-vault:0.2.0  |
| ---------------- | :----------:    | :-----------:    |
| 0.3.0            |   &#10003;      |  &#10007;        |
| 1.0.0            |   &#10007;      |  &#10003;        |

> N.B: For Kubernetes v1.13+ use `csi-vault:0.2.0`

## Supported `StorageClass` provisioner

| CSI Driver (csi-vault) Version | Provisioner Name            |  Kubernetes Version |
| ------------------------------ | --------------------------- | ------------------- |
| 0.1.0                          | `com.kubevault.csi.secrets` |  v1.12.x            |
| 0.3.0                          | `secrets.csi.kubevault.com` |  v1.14+             |
