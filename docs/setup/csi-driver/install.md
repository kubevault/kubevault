---
title: Install
description: Vault CSI Driver Install
menu:
  product_vault:
    identifier: install-csi-driver
    name: Install
    parent: setup
    weight: 10
product_name: csi-driver
menu_name: product_vault
section_menu_id: setup
---

# Installation Guide

Vault CSI driver can be installed via a script or as a Helm chart.

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

To install Vault CSI driver in your Kubernetes cluster, run the following command:

```console
$ curl -fsSL https://raw.githubusercontent.com/kubevault/csi-driver/0.1.0/hack/deploy/install.sh | bash
```

After successful installation, you should have `csi-vault-***` pod running in the `kube-system` namespace.

</div>
<div class="tab-pane fade" id="helm" role="tabpanel" aria-labelledby="helm-tab">

## Using Helm

Vault CSI driver can be installed via [Helm](https://helm.sh) using the [chart](https://github.com/appscode/kubevault/csi-driver/tree/0.1.0/chart/csi-vault) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `csi-vault`

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search appscode/csi-vault
NAME                      	CHART VERSION	APP VERSION	DESCRIPTION                                                 
appscode/csi-vault        	0.1.0        	0.1.0      	HashiCorp Vault CSI Driver for Kubernetes    

$ helm install appscode/csi-vault --name csi-vault --version 0.1.0 --namespace kube-system
```

To see the detailed configuration options, visit [here](https://github.com/kubevault/csi-driver/tree/chart/chart/csi-vault)
</div>


# Verify  Installation

To check if Vault CSI driver installed successfully, run the following command:

```console
$ kubectl get csinodeinfos
NAME              AGE
2gb-pool-77jne6   31s
```

If you can see the node's list, then your installation is ok.