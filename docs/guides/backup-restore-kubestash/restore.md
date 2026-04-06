---
title: Vault Backup Restore Overview
menu:
  docs_v2026.2.27:
    identifier: restore-backup-restore-guides-kubestash
    name: Restore
    parent: backup-restore-guides-kubestash
    weight: 10
menu_name: docs_v2026.2.27
section_menu_id: guides
info:
  cli: v0.24.0
  installer: v2026.2.27
  operator: v0.24.0
  unsealer: v0.24.0
  version: v2026.2.27
---

# Restore Vault Cluster using KubeStash

This guide will show you how you can restore your Vault cluster with KubeStash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeVault` in your cluster following the steps [here](/docs/v2026.2.27/setup/README).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how Stash backup and restore Vault cluster, please check the following concept section [here](/docs/v2026.2.27/concepts/backup-restore/overview).

- If you are not familiar with how KubeStash backup and restore ZooKeeper, please check the following guide [here](/docs/guides/zookeeper/backup/kubestash/overview/index.md).

You should be familiar with the following `KubeStash` concepts:

- [BackupStorage](https://kubestash.com/docs/latest/concepts/crds/backupstorage/)
- [RestoreSession](https://kubestash.com/docs/latest/concepts/crds/restoresession/)
- [Addon](https://kubestash.com/docs/latest/concepts/crds/addon/)
- [Function](https://kubestash.com/docs/latest/concepts/crds/function/)
- [Task](https://kubestash.com/docs/latest/concepts/crds/addon/#task-specification)

You may restore a Vault snapshot into the same Vault cluster from which snapshot was taken or into a 
completely new Vault deployment.

### Restore Snapshot for same Vault

Follow this guideline, if you want to restore a snapshot into the same Vault cluster. 
Vault cluster must be `Initialized` & `Unsealed` before trying to restore the snapshot.

Then, simply you can create a `RestoreSession` to restore the snapshot. A sample `RestoreSession` YAML may look like this:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-vault-restore
  namespace: demo
spec:
  target:
    apiGroup: appcatalog.appscode.com
    kind: AppBinding
    namespace: demo
    name: vault
  dataSource:
    repository: s3-vault-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: vault-addon
    tasks:
      - name: vault-restore
        params:
          keyPrefix: <vault-appbinding-key-prefix-value>
```

#### Create RestoreSession

Create the `RestoreSession` for restore the snapshot:

```yaml
$ kubectl apply -f restore-session.yaml
restoresession.core.kubestash.com/sample-vault-restore created
```

Now, wait for `RestoreSession` to succeed:

```bash
$ kubectl get restoresession -n demo

NAME                    REPOSITORY      PHASE       DURATION   AGE
sample-vault-restore    s3-vault-repo   Succeeded   19s        27s

```

Once the `RestoreSession` is Succeeded, snapshot will be successfully restored into the Vault cluster. 

### Restore Snapshot for different Vault

Follow this guideline, if you want to restore a snapshot into a different Vault cluster.

First, deploy a `PostgreSQL` database with the required tables (`vault_kv_store`, `vault_ha_locks`), set `max_connections` as per your needs, and use it as the backend for a new `VaultServer` — exactly as described in the [Backup guide](/docs/v2026.2.27/guides/backup-restore-kubestash/backup). The new `Vault` must be `Initialized` & `Unsealed`, and will have a completely different set of `unseal keys` & `root token` from the source `Vault`.

Once the new `Vault` is `Ready`, create the same storage secret and apply the same `BackupStorage` to sync the snapshot metadata. 

Then, scale down both the `VaultServer` and the KubeVault operator to `0` before applying the `RestoreSession` to restore the data safely.

```bash
kubectl scale deploy <kubevault-operator-deployment-name> -n <ns> --replicas=0
kubectl scale sts <vaultServer-statefulSet-name> -n <ns> --replicas=0
```

Then after restoreSession get successful, scale everything back up.

```bash
kubectl scale deploy <kubevault-operator-deployment-name> -n <ns> --replicas=3
kubectl scale sts <vaultServer-statefulSet-name> -n <ns> --replicas=3
```

`Vault` snapshot carries the signature of `unseal keys`. So, we need to restore the snapshot forcefully & to bypass this, we need to modify the `params` section of `RestoreSession` accordingly.

Let's take a look at some of the more relevant flags that we can set:

```bash
- --force=${force:=false}
- --key-prefix=${keyPrefix:=}
- --old-key-prefix=${oldKeyPrefix:=}
```

By default, the --force flag is false, so in order to restore the snapshot into a different Vault cluster, this must be set to true.
Moreover, once the snapshot is restored, the new Vault will be expecting the older unseal keys to unseal itself & the new unseal keys will not be required/valid anymore. So, we'll also migrate the older unseal keys & root token in place of the new ones.
Since KubeStash also takes backup of the Vault unseal keys & root token along with the snapshot, we can retrieve the older ones. To correctly get those, we must set the --old-key-prefix flag properly.

```bash
- --force=${force:=true}
- --key-prefix=${keyPrefix:=<restore-cluster-key-prefix>}
- --old-key-prefix=${oldKeyPrefix:=<old-key-prefix>}
```

KeyPrefix is generated by KubeVault operator using the structure: k8s.{kubevault.com or cluster UID}.{vault-namespace}.{vault-name}. You can check the AppBinding (created with the same name as the VaultServer) to find the correct prefix:

`kubectl get appbinding <vaultServer-name> -n <namespace> -o yaml`

Now create the RestoreSession:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-vault-restore
  namespace: demo
spec:
  target:
    apiGroup: appcatalog.appscode.com
    kind: AppBinding
    namespace: demo
    name: vault
  dataSource:
    repository: s3-vault-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: vault-addon
    tasks:
      - name: vault-restore
        params:
          keyPrefix: "<restore-vault-appbinding-key-prefix-value>"
          force: "true"
          oldKeyPrefix: "<backup-vault-appbinding-key-prefix-value>"
```

Create RestoreSession:

```bash
$ kubectl apply -f restore-session.yaml
restoresession.core.kubestash.com/sample-vault-restore created
```

Now, wait for RestoreSession to succeed:
```bash
$ kubectl get restoresession -n demo

NAME                    REPOSITORY      PHASE       DURATION   AGE
sample-vault-restore    s3-vault-repo   Succeeded   19s        27s
```

Once the RestoreSession is Succeeded, the snapshot will be successfully restored into the new Vault cluster.

To verify whether the Vault data has been successfully restored, export the necessary environment variables and port-forward the `vault-restore` service:

```bash
$ export VAULT_TOKEN=(kubectl vault root-token get vaultserver <vault-name> -n <ns> --value-only)
$ export VAULT_ADDR='http://127.0.0.1:8200'
$ kubectl port-forward -n demo svc/vault-restore 8200
```

Now, verify the currently enabled secret engines. The presence of the `kv` secret engine confirms the snapshot was successfully restored:

```bash
$ vault secrets list
Path                                                              Type         Accessor              Description
----                                                              ----         --------              -----------
cubbyhole/                                                        cubbyhole    cubbyhole_7e41c3a4    per-token private secret storage
identity/                                                         identity     identity_c14d17b1     identity store
k8s.89ce0ef5-4453-4b9b-a35f-b7aff8e48bf2.kv.demo.vault-health/    kv           kv_75c6f9fb           n/a
kv/                                                               kv           kv_9a19f9b9           n/a
sys/                                                              system       system_4609a4ac       system endpoints used for control, policy and debugging
```

Also verify the secrets stored in the `kv` engine:

```bash
$ vault kv get kv/name
	==== Data ====
Key     Value
---     -----
name    appscode
```

The restored data matches the original, confirming the restore was successful.

Up next:
- Read about step-by-step Backup procedure [here](/docs/v2026.2.27/guides/backup-restore-kubestash/backup)