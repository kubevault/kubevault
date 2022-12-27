---
title: Vault Backup Restore Overview | KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: overview-backup-restore-concepts
    name: Overview
    parent: backup-restore-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/README.md).

# Backup & Restore Vault Using Stash

KubeVault uses [Stash](https://stash.run) to backup and restore Vault. Stash by AppsCode is a cloud native data backup and recovery solution for Kubernetes workloads. Stash utilizes [restic](https://github.com/restic/restic) to securely backup stateful applications to any cloud or on-prem storage backends (for example, S3, GCS, Azure Blob storage, Minio, NetApp, Dell EMC etc.).

## How Backup Works

The following diagram shows how Stash takes a backup of a Vault cluster. Open the image in a new tab to see the enlarged version.

<figure align="center">
 <img alt="Vault Backup Overview" src="/docs/images/concepts/backup.svg">
  <figcaption align="center">Fig: Vault Backup Overview</figcaption>
</figure>

The backup process consists of the following steps:

1. At first, a user creates a secret with access credentials of the backend where the backed up data will be stored.

2. Then, the user creates a `Repository` crd that specifies the backend information along with the secret that holds the credentials to access the backend.

3. Then, the user creates a `BackupConfiguration` crd targeting the [AppBinding CRD](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) of the desired Vault cluster. The `BackupConfiguration` object also specifies the `Task` to use to take backup of the Vault cluster.

4. Stash operator watches for `BackupConfiguration` crd.

5. Once Stash operator finds a `BackupConfiguration` crd, it creates a CronJob with the schedule specified in `BackupConfiguration` object to trigger backup periodically.

6. On the next scheduled slot, the CronJob triggers a backup by creating a `BackupSession` crd.

7. Stash operator also watches for `BackupSession` crd.

8. When it finds a `BackupSession` object, it resolves the respective `Task` and `Function` and prepares a Job definition to take backup.

9. Then, it creates the Job to take backup the targeted Vault cluster.

10. The backup Job reads necessary information to connect with the Vault from the `AppBinding` crd. It also reads backend information and access credentials from `Repository` crd and Storage Secret respectively.

11. Then, the Job dumps snapshot from the targeted Vault and uploads the output to the backend. Stash stores the dumped files temporarily before uploading into the backend. Hence, you should provide a PVC template using `spec.interimVolumeTemplate` field of `BackupConfiguration` crd to use to store those dumped files temporarily.

12. Finally, when the backup is completed, the Job sends Prometheus metrics to the Pushgateway running inside Stash operator pod. It also updates the `BackupSession` and `Repository` status to reflect the backup procedure.

## How Restore Process Works

The following diagram shows how Stash restores backed up data into a Vault cluster. Open the image in a new tab to see the enlarged version.

<figure align="center">
 <img alt="Vault Restore Overview" src="/docs/images/concepts/restore.svg">
  <figcaption align="center">Fig: Vault Restore Process</figcaption>
</figure>

The restore process consists of the following steps:

1. At first, a user creates a `RestoreSession` crd targeting the `AppBinding` of the desired Vault where the backed up data will be restored. It also specifies the `Repository` crd which holds the backend information and the `Task` to use to restore the target.

2. Stash operator watches for `RestoreSession` object.

3. Once it finds a `RestoreSession` object, it resolves the respective `Task` and `Function` and prepares a Job definition to restore.

4. Then, it creates the Job to restore the target.

5. The Job reads necessary information to connect with the Vault from respective `AppBinding` crd. It also reads backend information and access credentials from `Repository` crd and Storage Secret respectively.

6. Then, the job downloads the backed up data from the backend and insert into the desired Vault. Stash stores the downloaded files temporarily before inserting into the targeted Vault. Hence, you should provide a PVC template using `spec.interimVolumeTemplate` field of `RestoreSession` crd to use to store those restored files temporarily.

7. Finally, when the restore process is completed, the Job sends Prometheus metrics to the Pushgateway and update the `RestoreSession` status to reflect restore completion.

## Next Steps

- Backup your Vault cluster using Stash following the guide from [here](/docs/guides/backup-restore/overview.md).
