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

To detect KubeVault version, exec into the operator pod and run `vault-operator version` command.

```bash
$ POD_NAMESPACE=kubevault
$ POD_NAME=$(kubectl get pods -n $POD_NAMESPACE -l app.kubernetes.io/instance=kubevault -o jsonpath={.items[0].metadata.name})
$ kubectl exec $POD_NAME -c operator -n $POD_NAMESPACE -- /vault-operator version

Version = {{< param "info.version" >}}
VersionStrategy = tag
GitTag = {{< param "info.version" >}}
GitBranch = HEAD
CommitHash = ad15b48a5ace19e0ec79934f7ebce709fb6dba59
CommitTimestamp = 2021-07-31T05:39:40
GoVersion = go1.16.6
Compiler = gcc
Platform = linux/amd64
```
