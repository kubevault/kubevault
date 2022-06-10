---
title: Welcome | KubeVault
description: Welcome to KubeVault
menu:
  docs_{{ .version }}:
    identifier: readme-kubevault
    name: Readme
    parent: welcome
    weight: -1
menu_name: docs_{{ .version }}
section_menu_id: welcome
url: /docs/{{ .version }}/welcome/
aliases:
  - /docs/{{ .version }}/
  - /docs/{{ .version }}/README/
---

![KubeVault Overview](/docs/images/kubevault-overview.svg)

# KubeVault

KubeVault by AppsCode is a collection of tools for running HashiCorp [Vault](https://www.vaultproject.io/) on Kubernetes. 

## Operator
You can deploy and manage Vault on Kubernetes clusters using KubeVault operator. From here you can learn all about Vault operator's architecture and how to deploy and use Vault operator.

- [Concepts](/docs/concepts/). Concepts explain the CRDs (CustomResourceDefinition) used by Vault operator.

- [Setup](/docs/setup/). Setup contains instructions for installing
  the Vault operator in various cloud providers.

- [Monitoring](/docs/guides/monitoring/overview.md). Monitoring contains instructions for setup prometheus with Vault server

- [Guides](/docs/guides/). Guides show you how to perform tasks with Vault operator.

- [Reference](/docs/reference/operator). Detailed exhaustive lists of
command-line options, configuration options, API definitions, and procedures.

## CLI

[Command line interface](https://github.com/kubevault/cli) for KubeVault. This is intended to be used as a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).

- [Reference](/docs/reference/cli). Detailed exhaustive lists of command-line options, configuration options, API definitions, and procedures.
- [Installation](/docs/setup/install/kubectl_plugin.md) Install Kubectl Vault Plugin.

## Unsealer

[Unsealer](https://github.com/kubevault/unsealer) automates the process of [initializing](https://www.vaultproject.io/docs/commands/operator/init.html) and [unsealing](https://www.vaultproject.io/docs/concepts/seal.html#unsealing) HashiCorp Vault instances running.

- [Reference](/docs/reference/unsealer). Detailed exhaustive lists of command-line options, configuration options, API definitions, and procedures.

## CSI Driver

KubeVault works seamlessly with [Secrets Store CSI driver for Kubernetes secrets](https://github.com/kubernetes-sigs/secrets-store-csi-driver).

- [Reference](https://secrets-store-csi-driver.sigs.k8s.io/) Kubernetes Secrets Store CSI Driver.


> We're always looking for help improving our documentation, so please don't hesitate to [file an issue](https://github.com/kubevault/kubevault/issues/new) if you see some problem. Or better yet, submit your own [contributions](/docs/CONTRIBUTING.md) to help
make our docs better.
