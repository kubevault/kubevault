---
title: Troubleshooting KubeVault Installation
description: Troubleshooting guide for KubeVault installation
menu:
  docs_{{ .version }}:
    identifier: install-kubevault-troubleshoot
    name: Troubleshooting
    parent: installation-guide
    weight: 40
product_name: kubevault
menu_name: docs_{{ .version }}
section_menu_id: setup
---

## Installing in GKE Cluster

If you are installing KubeVault on a GKE cluster, you will need cluster admin permissions to install KubeVault operator. Run the following command to grant admin permision to the cluster.

```bash
$ kubectl create clusterrolebinding "cluster-admin-$(whoami)" \
  --clusterrole=cluster-admin                                 \
  --user="$(gcloud config get-value core/account)"
```

In addition, if your GKE cluster is a [private cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/private-clusters), you will need to either add an additional firewall rule that allows master nodes access port `8443/tcp` on worker nodes, or change the existing rule that allows access to ports `443/tcp` and `10250/tcp` to also allow access to port `8443/tcp`. The procedure to add or modify firewall rules is described in the official GKE documentation for private clusters mentioned before.

## Detect KubeVault version

To detect KubeVault version, exec into the operator pod and run `kubevault version` command.

```bash
$ POD_NAMESPACE=kubevault
$ POD_NAME=$(kubectl get pods -n $POD_NAMESPACE -l app.kubernetes.io/name=kubevault -o jsonpath={.items[0].metadata.name})
$ kubectl exec $POD_NAME -c operator -n $POD_NAMESPACE -- /kubevault version

Version = {{< param "info.version" >}}
VersionStrategy = tag
Os = alpine
Arch = amd64
CommitHash = 85b0f16ab1b915633e968aac0ee23f877808ef49
GitBranch = release-0.5
GitTag = {{< param "info.version" >}}
CommitTimestamp = 2020-08-10T05:24:23

$ kubectl exec -it $POD_NAME -c operator -n $POD_NAMESPACE restic version
restic 0.9.6
compiled with go1.9 on linux/amd64
```
