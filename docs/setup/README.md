---
title: Table of Contents | Setup
description: Table of Contents | Setup
menu:
  docs_{{ .version }}:
    identifier: setup-readme
    name: Readme
    parent: setup
    weight: -1
menu_name: docs_{{ .version }}
section_menu_id: setup
url: /docs/{{ .version }}/setup/
aliases:
  - /docs/{{ .version }}/setup/README/
---
# Setup

Setup contains instructions for installing the Vault operator and Vault CSI driver in Kubernetes.

- Vault operator
  - [Install Vault operator](/docs/setup/operator/install.md). Installation instructions for Vault operator.
  - [Uninstall Vault operator](/docs/setup/operator/uninstall.md). Instructions for uninstallating Vault operator.
- Vault CSI driver
  - [Install Vault CSI driver](/docs/setup/csi-driver/install.md). Installation instructions for Vault CSI driver.
  - [Uninstall Vault CSI driver](/docs/setup/csi-driver/uninstall.md). Instructions for uninstallating Vault CSI driver.
- Kubectl Plugin
  - [Install Kubectl Plugin](/docs/setup/cli/install.md). Installation instructions for KubeVault `kubectl` plugin.
- Developer Guide
  - [Overview](/docs/setup/developer-guide/overview.md). Outlines everything you need to know from setting up your dev environment to how to build and test Vault operator.
  - [Release process](/docs/setup/developer-guide/release.md). Steps for releasing a new version of Vault operator.
