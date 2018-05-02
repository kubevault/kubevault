---
title: Weclome | Vault operator
description: Welcome to Vault operator
menu:
  product_vault-operator_0.1.0:
    identifier: readme-vault
    name: Readme
    parent: welcome
    weight: -1
product_name: vault-operator
menu_name: product_vault-operator_0.1.0
section_menu_id: welcome
url: /products/vault-operator/0.1.0/welcome/
aliases:
  - /products/vault-operator/0.1.0/
  - /products/vault-operator/0.1.0/README/
---
# Vault operator
 Vault operator by AppsCode is a Kubernetes operator for [restic](https://restic.net). If you are running production workloads in Kubernetes, you might want to take backup of your disks. Using Vault operator, you can backup Kubernetes volumes mounted in following types of workloads:

- Deployment
- DaemonSet
- ReplicaSet
- ReplicationController
- StatefulSet

From here you can learn all about Vault operator's architecture and how to deploy and use Vault operator.

- [Concepts](/docs/concepts/). Concepts explain some significant aspect of Vault operator. This is where you can learn about what Vault operator does and how it does it.

- [Setup](/docs/setup/). Setup contains instructions for installing
  the Vault operator in various cloud providers.

- [Guides](/docs/guides/). Guides show you how to perform tasks with Vault operator.

- [Reference](/docs/reference/). Detailed exhaustive lists of
command-line options, configuration options, API definitions, and procedures.

We're always looking for help improving our documentation, so please don't hesitate to [file an issue](https://github.com/kubevault/operator/issues/new) if you see some problem. Or better yet, submit your own [contributions](/docs/CONTRIBUTING.md) to help
make our docs better.

---

**Vault operator binaries collects anonymous usage statistics to help us learn how the software is being used and how we can improve it. To disable stats collection, run the operator with the flag** `--enable-analytics=false`.

---
