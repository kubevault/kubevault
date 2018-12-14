

## Prerequisite

- Kubernetes v1.12+
- `--allow-privileged` flag must be set to true for both the API server and the kubelet
- (If you use Docker) The Docker daemon of the cluster nodes must allow shared mounts
- Pre-installed HasiCorp Vault server.
- Pass `--feature-gates=CSIDriverRegistry=true,CSINodeInfo=true` to kubelet and kube-apiserver


## Supported [CSI Spec](https://github.com/container-storage-interface/spec) version

| CSI Spec Version | csi-vault:0.1.0 |
| ---------------- | :----------:    | 
| 0.3.0            |   &#10003;      | 