---
title: Install KubeVault
description: Installation guide for KubeVault
menu:
  docs_{{ .version }}:
    identifier: install-kubevault-enterprise
    name: KubeVault
    parent: installation-guide
    weight: 20
product_name: kubevault
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Install KubeVault

## Get a Free Trial License

In this section, we are going to show you how you can get a **30 days trial** license for KubeVault. You can get a license for your Kubernetes cluster by going through the following steps:

- At first, go to [AppsCode License Server](https://appscode.com/issue-license?p=kubevault) and fill up the form. It will ask for your Name, Email, the product you want to install, and your cluster ID (UID of the `kube-system` namespace).
- Provide your name and email address. **You must provide your work email address**.
- Then, select `KubeVault` in the product field.
- Now, provide your cluster ID. You can get your cluster ID easily by running the following command:

```bash
kubectl get ns kube-system -o=jsonpath='{.metadata.uid}'
```

- Then, you have to agree with the terms and conditions. We recommend reading it before checking the box.
- Now, you can submit the form. After you submit the form, the AppsCode License server will send an email to the provided email address with a link to your license file.
- Navigate to the provided link and save the license into a file. Here, we save the license to a `license.txt` file.

Here is a screenshot of the license form.

<figure align="center">
  <img alt="KubeVault Backend Overview" src="/docs/images/setup/enterprise_license_form.png">
  <figcaption align="center">Fig: KubeVault License Form</figcaption>
</figure>

You can create licenses for as many clusters as you want. You can upgrade your license any time without re-installing KubeVault by following the upgrading guide from [here](/docs/setup/upgrade/index.md#updating-license).

> KubeVault licensing process has been designed to work with CI/CD workflow. You can automatically obtain a license from your CI/CD pipeline by following the guide from [here](https://github.com/appscode/offline-license-server#offline-license-server).

## Purchase KubeVault License

If you are interested in purchasing KubeVault license, please contact us via sales@appscode.com for further discussion. You can also set up a meeting via our [calendly link](https://calendly.com/appscode/30min).

If you are willing to purchase KubeVault license but need more time to test in your dev cluster, feel free to contact sales@appscode.com. We will be happy to extend your trial period.

## Install

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="helm3-tab" data-toggle="tab" href="#helm3" role="tab" aria-controls="helm3" aria-selected="true">Helm 3 (Recommended)</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="false">YAML</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="helm3" role="tabpanel" aria-labelledby="helm3-tab">

## Using Helm 3

KubeVault can be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/kubevault/installer/tree/{{< param "info.installer" >}}/charts/kubevault) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install, follow the steps below:

```bash
$ helm install kubevault oci://ghcr.io/appscode-charts/kubevault \
    --version {{< param "info.version" >}} \
    --namespace kubevault --create-namespace \
    --set-file global.license=/path/to/the/license.txt \
    --wait --burst-limit=10000 --debug
```

To see the detailed configuration options, visit [here](https://github.com/kubevault/installer/tree/{{< param "info.installer" >}}/charts/kubevault).

</div>
<div class="tab-pane fade" id="script" role="tabpanel" aria-labelledby="script-tab">

## Using YAML

If you prefer to not use Helm, you can generate YAMLs from KubeVault chart and deploy using `kubectl`. Here we are going to show the procedure using Helm 3.

```bash
$ helm template kubevault oci://ghcr.io/appscode-charts/kubevault \
    --version {{< param "info.version" >}} \
    --namespace kubevault --create-namespace \
    --set-file global.license=/path/to/the/license.txt \
    --set global.skipCleaner=true | kubectl apply -f -
```

To see the detailed configuration options, visit [here](https://github.com/kubevault/installer/tree/{{< param "info.installer" >}}/charts/kubevault).

</div>
</div>

## Verify installation

To check if KubeVault operator pods have started, run the following command:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubevault"

NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubevault   kubevault-kubevault-operator-5d5cc4c7c9-mj5d5   1/1     Running   0          2m18s
```

Once the operator pod is running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm CRD groups have been registered by the operator, run the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubevault
```

Now, you are ready to [create your first database](/docs/guides/README.md) using KubeVault.
