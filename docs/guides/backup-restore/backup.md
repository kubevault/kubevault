---
title: Vault Backup Restore Overview
menu:
  docs_{{ .version }}:
    identifier: backup-backup-restore-guides
    name: Backup
    parent: backup-restore-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup Vault Cluster using Stash

This guide will show you how you can take backup of your Vault cluster with Stash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using Minikube.
- Install KubeVault in your cluster following the steps [here](/docs/setup/README.md).
- Install Stash Enterprise in your cluster following the steps [here](https://stash.run/docs/latest/setup/install/enterprise/).
- Install Stash `kubectl` plugin following the steps [here](https://stash.run/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how Stash backup and restore Vault cluster, please check the following concept section [here](/docs/concepts/backup-restore/overview.md).

You have to be familiar with following custom resources:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [Function](https://stash.run/docs/latest/concepts/crds/function/)
- [Task](https://stash.run/docs/latest/concepts/crds/task/)
- [BackupConfiguration](https://stash.run/docs/latest/concepts/crds/backupconfiguration/)
- [RestoreSession](https://stash.run/docs/latest/concepts/crds/restoresession/)