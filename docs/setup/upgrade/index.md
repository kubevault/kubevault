---
title: Upgrade | KubeVault
description: KubeVault Upgrade
menu:
  docs_{{ .version }}:
    identifier: upgrade-kubevault
    name: Upgrade
    parent: setup
    weight: 20
product_name: kubevault
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Upgrading KubeVault

This guide will show you how to upgrade various KubeVault components. Here, we are going to show how to upgrade from an old KubeVault version to the new version, and how to update the license, etc.

## Upgrading KubeVault to `{{< param "info.version" >}}`

In order to upgrade from KubeVault to `{{< param "info.version" >}}`, please follow the following steps.

#### 1. Update KubeVault Catalog CRDs

Helm [does not upgrade the CRDs](https://github.com/helm/helm/issues/6581) bundled in a Helm chart if the CRDs already exist. So, to upgrde the KubeVault catalog CRD, please run the command below:

```bash
kubectl apply -f https://github.com/kubevault/installer/raw/{{< param "info.version" >}}/crds/kubevault-catalog-crds.yaml
```

#### 2. Upgrade KubeVault Operator

Now, upgrade the KubeVault helm chart using the following command. You can find the latest installation guide [here](/docs/setup/README.md). We recommend that you do **not** follow the legacy installation guide, as the new process is much more simple.

```bash
$ helm upgrade kubevault oci://ghcr.io/appscode-charts/kubevault \
  --version {{< param "info.version" >}} \
  --namespace kubevault \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

## Updating License

KubeVault support updating license without requiring any re-installation. KubeVault creates a Secret named `<helm release name>-license` with the license file. You just need to update the Secret. The changes will propagate automatically to the operator and it will use the updated license going forward.

Follow the below instructions to update the license:

- Get a new license and save it into a file.
- Then, run the following upgrade command based on your installation.

<ul class="nav nav-tabs" id="luTabs" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="lu-helm3-tab" data-toggle="tab" href="#lu-helm3" role="tab" aria-controls="lu-helm3" aria-selected="true">Helm 3</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="lu-yaml-tab" data-toggle="tab" href="#lu-yaml" role="tab" aria-controls="lu-yaml" aria-selected="false">YAML</a>
  </li>
</ul>
<div class="tab-content" id="luTabContent">
  <div class="tab-pane fade show active" id="lu-helm3" role="tabpanel" aria-labelledby="lu-helm3">

#### Using Helm 3

```bash
# detect current version
helm ls -A | grep kubevault

# update license key keeping the current version
helm upgrade kubevault oci://ghcr.io/appscode-charts/kubevault \
  --version=<cur_version> \
  --namespace=kubevault --create-namespace \
  --reuse-values \
  --set-file global.license=/path/to/new/license.txt \
  --wait --burst-limit=10000 --debug
```

</div>
<div class="tab-pane fade" id="lu-yaml" role="tabpanel" aria-labelledby="lu-yaml">

#### Using YAML (with helm 3)

```bash
# detect current version
helm ls -A | grep kubevault

# update license key keeping the current version
helm template kubevault oci://ghcr.io/appscode-charts/kubevault \
  --version=<cur_version> \
  --namespace=kubevault --create-namespace \
  --set global.skipCleaner=true \
  --show-only appscode/kubevault-operator/templates/license.yaml \
  --set-file global.license=/path/to/new/license.txt | kubectl apply -f -
```

</div>
</div>
