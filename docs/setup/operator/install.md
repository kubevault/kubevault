---
title: Install
description: Vault operator Install
menu:
  product_vault-operator_0.1.0:
    identifier: install-vault
    name: Install
    parent: setup
    weight: 10
product_name: vault-operator
menu_name: product_vault-operator_0.1.0
section_menu_id: setup
---

# Installation Guide

Vault operator can be installed via a script or as a Helm chart.

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="true">Script</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="helm-tab" data-toggle="tab" href="#helm" role="tab" aria-controls="helm" aria-selected="false">Helm</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="script" role="tabpanel" aria-labelledby="script-tab">

## Using Script

To install Vault operator in your Kubernetes cluster, run the following command:

```console
$ curl -fsSL https://raw.githubusercontent.com/kubevault/operator/0.1.0/hack/deploy/install.sh | bash
```

After successful installation, you should have a `vault-operator-***` pod running in the `kube-system` namespace.

```console
$ kubectl get pods -n kube-system | grep vault-operator
vault-operator-846d47f489-jrb58       1/1       Running   0          48s
```

#### Customizing Installer

The installer script and associated yaml files can be found in the [/hack/deploy](https://github.com/kubevault/operator/tree/0.1.0/hack/deploy) folder. You can see the full list of flags available to installer using `-h` flag.

```console
$ curl -fsSL https://raw.githubusercontent.com/kubevault/operator/0.1.0/hack/deploy/install.sh | bash -s -- -h
install.sh - install Vault operator

install.sh [options]

options:
-h, --help                             show brief help
-n, --namespace=NAMESPACE              specify namespace (default: kube-system)
    --docker-registry                  docker registry used to pull Vault images (default: kubevault)
    --image-pull-secret                name of secret used to pull Vault images
    --run-on-master                    run Vault operator on master
    --enable-mutating-webhook          enable/disable mutating webhooks for Kubernetes workloads
    --enable-validating-webhook        enable/disable validating webhooks for Stash crds
    --bypass-validating-webhook-xray   if true, bypasses validating webhook xray checks
    --enable-status-subresource        if enabled, uses status sub resource for crds
    --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
    --enable-analytics                 send usage events to Google Analytics (default: true)
    --uninstall                        uninstall stash
    --purge                            purges stash crd objects and crds
    --install-catalog                  installs Vault server version catalog (default: all)
```

If you would like to run Vault operator pod in `master` instances, pass the `--run-on-master` flag:

```console
$ curl -fsSL https://raw.githubusercontent.com/kubevault/operator/0.1.0/hack/deploy/install.sh \
    | bash -s -- --run-on-master
```

Vault operator will be installed in a `kube-system` namespace by default. If you would like to run Vault operator pod in `vault` namespace, pass the `--namespace=vault` flag:

```console
$ kubectl create namespace vault
$ curl -fsSL https://raw.githubusercontent.com/kubevault/operator/0.1.0/hack/deploy/install.sh \
    | bash -s -- --namespace=vault [--run-on-master]
```

If you are using a private Docker registry, you need to pull the following images:

 - [kubevault/vault-operator](https://hub.docker.com/r/kubevault/vault-operator)
 - [kubevault/vault-unsealer](https://hub.docker.com/r/kubevault/vault-unsealer)
 - [kubevault/csi-vault](https://hub.docker.com/r/kubevault/csi-vault)

To pass the address of your private registry and optionally a image pull secret use flags `--docker-registry` and `--image-pull-secret` respectively.

```console
$ kubectl create namespace vault
$ curl -fsSL https://raw.githubusercontent.com/kubevault/operator/0.1.0/hack/deploy/install.sh \
    | bash -s -- --docker-registry=MY_REGISTRY [--image-pull-secret=SECRET_NAME]
```

Vault operator implements [validating admission webhooks](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#validatingadmissionwebhook) to validate KubeVault CRDs and **mutating webhooks** for KubeVault crds. This is enabled by default for Kubernetes 1.9.0 or later releases. To disable this feature, pass the `--enable-validating-webhook=false` and `--enable-mutating-webhook=false` flag respectively.

```console
$ curl -fsSL https://raw.githubusercontent.com/kubevault/operator/0.1.0/hack/deploy/install.sh \
    | bash -s -- --enable-validating-webhook=false --enable-mutating-webhook=false
```

</div>
<div class="tab-pane fade" id="helm" role="tabpanel" aria-labelledby="helm-tab">

## Using Helm
Vault operator can be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/kubevault/operator/tree/0.1.0/chart/vault-operator) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `my-release`:

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search appscode/vault
NAME                    CHART VERSION APP VERSION   DESCRIPTION
appscode/vault-operator 0.1.0         0.1.0         Vault Operator by AppsCode - HashiCorp Vault Operator for Kubernetes
appscode/vault-catalog  0.1.0         0.1.0         Vault Catalog by AppsCode - Catalog for vault versions

# Step 1: Install vault-operator chart
$ helm install appscode/vault-operator --name vault-operator --version 0.1.0 \
  --namespace kube-system

# Step 2: wait until crds are registered
$ kubectl get crds -l app=vault -w
NAME                                        AGE
vaultservers.kubevault.com                  12s
vaultserverversions.catalog.kubevault.com    8s

# Step 3: Install catalog of Vault versions
$ helm install appscode/vault-catalog --name vault-catalog

# Step 3(a): Install catalog of Vault versions
$ helm install appscode/vault-catalog --name vault-catalog --version 0.1.0 \
  --namespace kube-system

# Step 3(b): Or, if previously installed, upgrade catalog of Vault versions
$ helm upgrade vault-catalog appscode/vault-catalog --version 0.1.0 \
  --namespace kube-system
```

To see the detailed configuration options, visit [here](https://github.com/kubevault/operator/tree/master/chart/vault-operator).

</div>

### Installing in GKE Cluster

If you are installing Vault operator on a GKE cluster, you will need cluster admin permissions to install Vault operator. Run the following command to grant admin permision to the cluster.

```console
$ kubectl create clusterrolebinding "cluster-admin-$(whoami)" \
  --clusterrole=cluster-admin \
  --user="$(gcloud config get-value core/account)"
```


## Verify installation
To check if Vault operator pods have started, run the following command:
```console
$ kubectl get pods --all-namespaces -l app=vault --watch

NAMESPACE     NAME                              READY     STATUS    RESTARTS   AGE
kube-system   vault-operator-859d6bdb56-m9br5   2/2       Running   2          5s
```

Once the operator pods are running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm CRD groups have been registered by the operator, run the following command:
```console
$ kubectl get crd -l app=vault

NAME                                 AGE
recoveries.vault.appscode.com        5s
repositories.vault.appscode.com      5s
restics.vault.appscode.com           5s
```

Now, you are ready to [take your first backup](/docs/guides/README.md) using Vault operator.


## Configuring RBAC
Vault operator creates multiple CRDs: `Restic`, `Repository` and `Recovery`. Vault operator installer will create 2 user facing cluster roles:

| ClusterRole         | Aggregates To | Desription                            |
|---------------------|---------------|---------------------------------------|
| appscode:vault:edit | admin, edit   | Allows edit access to Vault operator CRDs, intended to be granted within a namespace using a RoleBinding. |
| appscode:vault:view | view          | Allows read-only access to Vault operator CRDs, intended to be granted within a namespace using a RoleBinding. |

These user facing roles supports [ClusterRole Aggregation](https://kubernetes.io/docs/admin/authorization/rbac/#aggregated-clusterroles) feature in Kubernetes 1.9 or later clusters.


## Using kubectl for Restic
```console
# List all Restic objects
$ kubectl get restic --all-namespaces

# List Restic objects for a namespace
$ kubectl get restic -n <namespace>

# Get Restic YAML
$ kubectl get restic -n <namespace> <name> -o yaml

# Describe Restic. Very useful to debug problems.
$ kubectl describe restic -n <namespace> <name>
```


## Using kubectl for Recovery
```console
# List all Recovery objects
$ kubectl get recovery --all-namespaces

# List Recovery objects for a namespace
$ kubectl get recovery -n <namespace>

# Get Recovery YAML
$ kubectl get recovery -n <namespace> <name> -o yaml

# Describe Recovery. Very useful to debug problems.
$ kubectl describe recovery -n <namespace> <name>
```


## Detect Vault operator version
To detect Vault operator version, exec into the operator pod and run `vault version` command.

```console
$ POD_NAMESPACE=kube-system
$ POD_NAME=$(kubectl get pods -n $POD_NAMESPACE -l app=vault -o jsonpath={.items[0].metadata.name})
$ kubectl exec -it $POD_NAME -c operator -n $POD_NAMESPACE vault version

Version = 0.1.0
VersionStrategy = tag
Os = alpine
Arch = amd64
CommitHash = 85b0f16ab1b915633e968aac0ee23f877808ef49
GitBranch = release-0.5
GitTag = 0.1.0
CommitTimestamp = 2017-10-10T05:24:23

$ kubectl exec -it $POD_NAME -c operator -n $POD_NAMESPACE restic version
restic 0.8.3
compiled with go1.9 on linux/amd64
```
