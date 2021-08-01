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

This guide will show you how to upgrade various KubeVault components. Here, we are going to show how to upgrade from an old KubeVault version to the new version, how to migrate between the enterprise edition and community edition, and how to update the license, etc.

## Upgrading KubeVault from `v2021.xx.xx` to `v2021.06.23`

In order to upgrade from KubeVault `v2021.xx.xx` to `v2021.06.23`, please follow the following steps.

#### 1. Update KubeVault Catalog CRDs

KubeVault `v2021.06.23` has added some new fields in the `***Version` CRDs. Unfortunatley, Helm [does not upgrade the CRDs](https://github.com/helm/helm/issues/6581) bundled in a Helm chart if the CRDs already exist. So, to upgrde the KubeVault catalog CRD, please run the command below:

```bash
kubectl apply -f https://github.com/kubevault/installer/raw/v2021.06.23/kubevault-catalog-crds.yaml
```

#### 2. Upgrade KubeVault Operator

Now, upgrade the KubeVault helm chart using the following command. You can find the latest installation guide [here](/docs/setup/README.md). We recommend that you do **not** follow the legacy installation guide, as the new process is much more simpler.

```bash
# Upgrade KubeVault Community operator chart
$ helm upgrade kubevault appscode/kubevault \
  --version {{< param "info.version" >}} \
  --namespace kubevault \
  --set-file global.license=/path/to/the/license.txt

# Upgrade KubeVault Enterprise operator chart
$ helm upgrade kubevault appscode/kubevault \
    --version {{< param "info.version" >}} \
    --namespace kubevault \
    --set-file global.license=/path/to/the/license.txt \
    --set kubevault-enterprise.enabled=true \
    --set kubevault-autoscaler.enabled=true
```

#### 3. Install/Upgrade Stash Operator

Now, upgrade Stash if had previously installed Stash following the instructions [here](https://stash.run/docs/v2021.06.23/setup/upgrade/). If you had not installed Stash before, please install Stash Enterprise Edition following the instructions [here](https://stash.run/docs/v2021.06.23/setup/).


## Upgrading KubeVault from `v2021.01.26`(`v0.16.x`) and older to `v2021.03.17`(`v0.17.x`)

In KubeVault `v2021.01.26`(`v0.16.x`) and prior versions, KubeVault used separate charts for KubeVault community edition, KubeVault enterprise edition, and KubeVault catalogs. In KubeVault `v2021.03.17`(`v0.17.x`), we have moved to a single combined chart for all the components for a better user experience. This enables seamless migration between the KubeVault community edition and KubeVault enterprise edition. It also removes the burden of installing individual helm charts manually. KubeVault still depends on [Stash](https://stash.run) as the backup/recovery operator and Stash must be [installed](https://stash.run/docs/latest/setup/) separately. 

In order to upgrade from KubeVault `v2021.01.26`(`v0.16.x`) to `v2021.03.17`(`v0.17.x`), please follow the following steps.

#### 1. Uninstall KubeVault Operator

Uninstall the old KubeVault operator by following the appropriate uninstallation guide of the KubeVault version that you are currently running.

>Make sure you are using the appropriate version of the uninstallation guide. The uninstallation guide for `v2021.03.17`(`v0.17.x`) will not work for `v2021.01.26`(`v0.16.x`) Use the dropdown at the sidebar of the documentation site to navigate to the appropriate version that you are currently running.

#### 2. Update KubeVault Catalog CRDs

KubeVault `v2021.03.17`(`v0.17.x`) has added some new fields in the `***Version` CRDs. Unfortunatley, Helm [does not upgrade the CRDs](https://github.com/helm/helm/issues/6581) bundled in a Helm chart if the CRDs already exist. So, to upgrde the KubeVault catalog CRD, please run the command below:

```bash
kubectl apply -f https://github.com/kubevault/installer/raw/v0.17.1/kubevault-catalog-crds.yaml
```

#### 3. Reinstall new KubeVault Operator

Now, follow the latest installation guide to install the new version of the KubeVault operator. You can find the latest installation guide [here](/docs/setup/README.md). We recommend that you do **not** follow the legacy installation guide, as the new process is much more simpler.

#### 4. Install/Upgrade Stash Operator

Now, upgrade Stash if had previously installed Stash following the instructions [here](https://stash.run/docs/v2021.03.17/setup/upgrade/). If you had not installed Stash before, please install Stash Enterprise Edition following the instructions [here](https://stash.run/docs/v2021.03.17/setup/).


## Migration Between Community Edition and Enterprise Edition

KubeVault `v2021.06.23` supports seamless migration between community edition and enterprise edition. You can run the following commands to migrate between them.

<ul class="nav nav-tabs" id="migrationTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="mgr-helm3-tab" data-toggle="tab" href="#mgr-helm3" role="tab" aria-controls="mgr-helm3" aria-selected="true">Helm 3</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="mgr-yaml-tab" data-toggle="tab" href="#mgr-yaml" role="tab" aria-controls="mgr-yaml" aria-selected="false">YAML</a>
  </li>
</ul>
<div class="tab-content" id="migrationTabContent">
  <div class="tab-pane fade show active" id="mgr-helm3" role="tabpanel" aria-labelledby="mgr-helm3">

#### Using Helm 3

**From Community Edition to Enterprise Edition:**

In order to migrate from KubeVault community edition to KubeVault enterprise edition, please run the following command,

```bash
helm upgrade kubevault -n kubevault appscode/kubevault \
  --reuse-values \
  --set kubevault-enterprise.enabled=true \
  --set kubevault-autoscaler.enabled=true \
  --set kubevault-catalog.skipDeprecated=false \
  --set-file global.license=/path/to/kubevault-enterprise-license.txt
```

**From Enterprise Edition to Community Edition:**

In order to migrate from KubeVault enterprise edition to KubeVault community edition, please run the following command,

```bash
helm upgrade kubevault -n kubevault appscode/kubevault \
  --reuse-values \
  --set kubevault-enterprise.enabled=false \
  --set kubevault-autoscaler.enabled=false \
  --set kubevault-catalog.skipDeprecated=false \
  --set-file global.license=/path/to/kubevault-community-license.txt
```

</div>
<div class="tab-pane fade" id="mgr-yaml" role="tabpanel" aria-labelledby="mgr-yaml">

**Using YAML (with helm 3)**

**From Community Edition to Enterprise Edition:**

In order to migrate from KubeVault community edition to KubeVault enterprise edition, please run the following command,

```bash
# Install KubeVault enterprise edition
helm template kubevault -n kubevault appscode/kubevault \
  --version {{< param "info.version" >}} \
  --set kubevault-enterprise.enabled=true \
  --set kubevault-autoscaler.enabled=true \
  --set kubevault-catalog.skipDeprecated=false \
  --set global.skipCleaner=true \
  --set-file global.license=/path/to/kubevault-enterprise-license.txt | kubectl apply -f -
```

**From Enterprise Edition to Community Edition:**

In order to migrate from KubeVault enterprise edition to KubeVault community edition, please run the following command,

```bash
# Install KubeVault community edition
helm template kubevault -n kubevault appscode/kubevault \
  --version {{< param "info.version" >}} \
  --set kubevault-enterprise.enabled=false \
  --set kubevault-autoscaler.enabled=false \
  --set kubevault-catalog.skipDeprecated=false \
  --set global.skipCleaner=true \
  --set-file global.license=/path/to/kubevault-community-license.txt | kubectl apply -f -
```

</div>
</div>

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
helm upgrade kubevault -n kubevault appscode/kubevault \
  --reuse-values \
  --set-file global.license=/path/to/new/license.txt
```

</div>
<div class="tab-pane fade" id="lu-yaml" role="tabpanel" aria-labelledby="lu-yaml">

#### Using YAML (with helm 3)

**Update License of Community Edition:**

```bash
helm template kubevault -n kubevault appscode/kubevault \
  --set kubevault-enterprise.enabled=false \
  --set kubevault-autoscaler.enabled=false \
  --set global.skipCleaner=true \
  --show-only appscode/kubevault-community/templates/license.yaml \
  --set-file global.license=/path/to/new/license.txt | kubectl apply -f -
```

**Update License of Enterprise Edition:**

```bash
helm template kubevault appscode/kubevault -n kubevault \
  --set kubevault-enterprise.enabled=true \
  --set kubevault-autoscaler.enabled=true \
  --set global.skipCleaner=true \
  --show-only appscode/kubevault-enterprise/templates/license.yaml \
  --set-file global.license=/path/to/new/license.txt | kubectl apply -f -
```

</div>
</div>
