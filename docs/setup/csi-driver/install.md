---
title: Install Vault CSI Driver
menu:
  docs_0.2.0:
    identifier: install-csi-driver
    name: Install
    parent: csi-driver-setup
    weight: 10
menu_name: docs_0.2.0
section_menu_id: setup
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

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
# Kubernetes 1.13+ (CSI driver spec 1.0.0)
$ curl -fsSL https://raw.githubusercontent.com/kubevault/csi-driver/0.2.0/hack/deploy/install.sh | bash
```

After successful installation, you should have `csi-vault-***` pod running in the `kube-system` namespace.


#### Customizing Installer

The installer script and associated yaml files can be found in the [/hack/deploy](https://github.com/kubevault/csi-driver/tree/0.2.0/hack/deploy) folder. You can see the full list of flags available to installer using `-h` flag.

```console
$ curl -fsSL https://raw.githubusercontent.com/kubevault/csi-driver/0.2.0/hack/deploy/install.sh | bash -s -- -h
install.sh -install Vault csi driver

install.sh [options]

options:
-h, --help                                show brief help
-n, --namespace=NAMESPACE                 specify namespace (default: kube-system)
    --csi-vault-docker-registry           docker registry used to pull csi-vault image (default: kubevault)
    --csi-vault-image-pull-secret         name of secret used to pull csi-vault images
    --csi-vault-image-tag                 docker image version of csi vault
    --csi-attacher-docker-registry        docker registry used to pull csi attacher image (default: quay.io/k8scsi)
    --csi-attacher-image-pull-secret      name of secret used to pull csi attacher image
    --csi-attacher-image-tag              docker image version of csi attacher
    --csi-provisioner-docker-registry     docker registry used to pull csi provisioner image (default: quay.io/k8scsi)
    --csi-provisioner-image-pull-secret   name of secret used to pull csi provisioner image
    --csi-provisioner-image-tag           docker image version of csi provisioner
    --csi-registrar-docker-registry       docker registry used to pull csi registrar image (default: quay.io/k8scsi)
    --csi-registrar-image-pull-secret     name of secret used to pull csi registrar image
    --csi-registrar-image-tag             docker image version of csi registrar
    --csi-driver-name                     name of csi driver to install (default: secrets.csi.kubevault.com)
    --csi-required-attachment             indicates csi volume driver requires an attach operation (default: false)
    --install-appbinding                  indicates appbinding crd need to be installed (default: true)
    --monitoring-agent                    specify which monitoring agent to use (default: none)
    --monitor-attacher                    specify whether to monitor Vault CSI driver attacher (default: false)
    --monitor-plugin                      specify whether to monitor Vault CSI driver plugin (default: false)
    --monitor-provisioner                 specify whether to monitor Vault CSI driver provisioner (default: false)
    --prometheus-namespace                specify the namespace where Prometheus server is running or will be deployed (default: same namespace as csi-vault)
    --servicemonitor-label                specify the label for ServiceMonitor crd. Prometheus crd will use this label to select the ServiceMonitor. (default: 'app: csi-vault')
    --uninstall                           uninstall vault csi driver
    --purge                               purges csi driver crd objects and crds
```

</div>
<div class="tab-pane fade" id="helm" role="tabpanel" aria-labelledby="helm-tab">

## Using Helm

Vault CSI driver can be installed via [Helm](https://helm.sh) using the [chart](https://github.com/appscode/kubevault/csi-driver/tree/0.2.0/chart/csi-vault) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `csi-vault`

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search appscode/csi-vault
NAME              	CHART VERSION	APP VERSION	DESCRIPTION
appscode/csi-vault	0.2.0        	0.2.0      	HashiCorp Vault CSI Driver for Kubernetes

# Kubernetes 1.13+ (CSI driver spec 1.0.0)
$ helm install appscode/csi-vault --name csi-vault --version 0.2.0 --namespace kube-system
```

To see the detailed configuration options, visit [here](https://github.com/kubevault/csi-driver/tree/chart/chart/csi-vault)
</div>


## Verify  Installation

To check if Vault CSI driver installed successfully, run the following command:

```console
$ kubectl get csinodeinfos
NAME              AGE
2gb-pool-77jne6   31s
```

If you can see the node's list, then your installation is ok.