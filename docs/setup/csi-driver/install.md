---
title: Install Vault CSI Driver
menu:
  docs_{{ .version }}:
    identifier: install-csi-driver
    name: Install
    parent: csi-driver-setup
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: setup
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Installation Guide

Vault CSI driver can be installed via a script or as a Helm chart.

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="helm3-tab" data-toggle="tab" href="#helm3" role="tab" aria-controls="helm3" aria-selected="true">Helm 3 (Recommended)</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="helm2-tab" data-toggle="tab" href="#helm2" role="tab" aria-controls="helm2" aria-selected="false">Helm 2</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="false">Script</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="helm3" role="tabpanel" aria-labelledby="helm3-tab">

## Using Helm 3

Vault CSI driver can be installed via [Helm](https://helm.sh) using the [chart](https://github.com/appscode/kubevault/installer/tree/{{< param "info.version" >}}/charts/csi-vault) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `csi-vault`

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search repo appscode/csi-vault --version {{< param "info.version" >}}
NAME                CHART VERSION   APP VERSION DESCRIPTION
appscode/csi-vault  {{< param "info.version" >}}            {{< param "info.version" >}}        HashiCorp Vault CSI Driver for Kubernetes

# Kubernetes 1.14+ (CSI driver spec 1.0.0)
$ helm install csi-vault appscode/csi-vault \
  --version {{< param "info.version" >}} \
  --namespace kube-system
```

To see the detailed configuration options, visit [here](https://github.com/kubevault/installer/tree/{{< param "info.version" >}}/charts/csi-vault).

</div>
<div class="tab-pane fade" id="helm2" role="tabpanel" aria-labelledby="helm2-tab">

## Using Helm 2

Vault CSI driver can be installed via [Helm](https://helm.sh) using the [chart](https://github.com/appscode/kubevault/installer/tree/{{< param "info.version" >}}/charts/csi-vault) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `csi-vault`

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search appscode/csi-vault --version {{< param "info.version" >}}
NAME              	CHART VERSION	APP VERSION	DESCRIPTION
appscode/csi-vault	{{< param "info.version" >}}        	{{< param "info.version" >}}      	HashiCorp Vault CSI Driver for Kubernetes

# Kubernetes 1.14+ (CSI driver spec 1.0.0)
$ helm install appscode/csi-vault --name csi-vault \
  --version {{< param "info.version" >}} \
  --namespace kube-system
```

To see the detailed configuration options, visit [here](https://github.com/kubevault/installer/tree/{{< param "info.version" >}}/charts/csi-vault).

</div>
<div class="tab-pane fade" id="script" role="tabpanel" aria-labelledby="script-tab">

## Using YAML

If you prefer to not use Helm, you can generate YAMLs from Vault CSI driver chart and deploy using `kubectl`. Here we are going to show the prodecure using Helm 3.

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search repo appscode/csi-vault --version {{< param "info.version" >}}
NAME                CHART VERSION   APP VERSION DESCRIPTION
appscode/csi-vault  {{< param "info.version" >}}            {{< param "info.version" >}}        HashiCorp Vault CSI Driver for Kubernetes

# Kubernetes 1.14+ (CSI driver spec 1.0.0)
$ helm template csi-vault appscode/csi-vault \
  --version {{< param "info.version" >}} \
  --namespace kube-system \
  --no-hooks | kubectl apply -f -
```

To see the detailed configuration options, visit [here](https://github.com/kubevault/installer/tree/{{< param "info.version" >}}/charts/csi-vault).

</div>
</div>

## Verify  Installation

To check if Vault CSI driver installed successfully, run the following command:

```console
$ kubectl get csinodeinfos
NAME              AGE
2gb-pool-77jne6   31s
```

If you can see the node's list, then your installation is ok.
