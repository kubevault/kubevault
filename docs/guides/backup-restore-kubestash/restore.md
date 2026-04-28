---
title: Vault Backup Restore KubeStash Overview 
menu:
  docs_{{ .version }}:
    identifier: restore-backup-restore-guides-kubestash
    name: Restore
    parent: backup-restore-guides-kubestash
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Restore Vault Cluster using KubeStash

This guide will show you how you can restore your Vault cluster with KubeStash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](https://kubedb.com/docs/latest/setup/install/kubedb).
- Install `KubeVault` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore Vault cluster, please check the following concept section [here](/docs/guides/backup-restore-kubestash/overview.md).

You should be familiar with the following `KubeStash` concepts:

- [BackupStorage](https://kubestash.com/docs/latest/concepts/crds/backupstorage/)
- [RestoreSession](https://kubestash.com/docs/latest/concepts/crds/restoresession/)
- [Addon](https://kubestash.com/docs/latest/concepts/crds/addon/)
- [Function](https://kubestash.com/docs/latest/concepts/crds/function/)
- [Task](https://kubestash.com/docs/latest/concepts/crds/addon/#task-specification)

You may restore a Vault snapshot into the same Vault cluster from which snapshot was taken or into a 
completely new Vault deployment.

### Restore Snapshot for same Vault Cluster

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
    name: restore-vault
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
          keyPrefix: k8s.4a6d4bda-4c08-49a5-b708-5e6d4b2f10f3.demo.restore-vault
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

First, deploy a `PostgreSQL` database with the required tables (`vault_kv_store`, `vault_ha_locks`), set `max_connections` as per your needs, and use it as the backend for a new `VaultServer` — exactly as described in the [Backup guide](/docs/guides/backup-restore-kubestash/backup.md). The new `Vault` must be `Initialized` & `Unsealed`, and will have a completely different set of `unseal keys` & `root token` from the source `Vault`.

Once the new `Vault` is `Ready`, create the same storage secret and apply the same `BackupStorage` to sync the snapshot metadata. 

Verify the snapshot is present in the new cluster by running:

```bash
$ kubectl get snapshots -n demo
```

Then, scale down both the `VaultServer` and the KubeVault operator to `0` before applying the `RestoreSession` to restore the data safely.

```bash
kubectl scale deploy <kubevault-operator-deployment-name> -n <ns> --replicas=0
kubectl scale sts <vaultServer-statefulSet-name> -n <ns> --replicas=0
```


`Vault` snapshot carries the signature of `unseal keys`. So, we need to restore the snapshot forcefully & to bypass this, we need to modify the `params` section of `RestoreSession` accordingly.

Let's take a look at some of the more relevant flags that we can set:

```bash
- --force=${force:=false}
- --key-prefix=${keyPrefix:=}
- --old-key-prefix=${oldKeyPrefix:=}
```

These flags are described in detail below:

1. **`--force`**
   - Default value: `false`.
   - Must be set to `true` when restoring a snapshot into a **different** Vault cluster (i.e., a cluster other than the one from which the backup was taken). This is required even when restoring to a separate cluster in a different Kubernetes environment.
   - Without this flag, the restore process will refuse to overwrite an existing Vault instance.

2. **`--key-prefix`** (i.e., `keyPrefix`)
   - This is the prefix used for the **restore target cluster's** Vault unseal keys and root token stored in Kubernetes Secrets.
   - `KubeVault` operator auto-generates this prefix using the format:
     ```
     k8s.<cluster-uid>.<vault-namespace>.<vault-name>
     ```
     For example: `k8s.c977fff4-e3b5-4232-8e3e-3bb52106e057.demo.restore-vault`
   - The cluster UID is the UID of the Kubernetes cluster itself (not the VaultServer). For clusters managed by `KubeVault`, this is embedded automatically.
   - **How to find it:** Inspect the `AppBinding` of the **restore target** Vault cluster:
     ```bash
     kubectl get appbinding -n <vault-namespace> <vault-name> -oyaml
     ```
     Look under `spec.parameters.stash.addon.restoreTask.params` for the `keyPrefix` value:
     ```yaml
     spec:
       parameters:
         stash:
           addon:
             restoreTask:
               name: vault-restore-1.10.3
               params:
               - name: keyPrefix
                 value: k8s.c977fff4-e3b5-4232-8e3e-3bb52106e057.demo.restore-vault
     ```

3. **`--old-key-prefix`** (i.e., `oldKeyPrefix`)
   - This is the prefix used for the **backup source cluster's** Vault unseal keys and root token.
   - After a snapshot is restored, the target Vault will unseal itself using the **old** unseal keys (from the source cluster). The `KubeVault` operator uses `--old-key-prefix` to locate and migrate those old unseal keys and root token in place of the new ones.
   - **How to find it:** Inspect the `AppBinding` of the **backup source** Vault cluster (or the `AppBinding` that was used during the backup):
     ```bash
     kubectl get appbinding -n <vault-namespace> <vault-name> -oyaml
     ```
     Look under `spec.parameters.stash.addon.backupTask.params` for the `keyPrefix` value:
     ```yaml
     spec:
       parameters:
         stash:
           addon:
             backupTask:
               name: vault-backup-1.10.3
               params:
               - name: keyPrefix
                 value: k8s.4a6d4bda-4c08-49a5-b708-5e6d4b2f10f3.demo.vault
     ```
     This value becomes the `--old-key-prefix` when restoring into the target cluster.

> **Note:** This process works for restoring into a **separate Kubernetes cluster** as well. In that case, the `--old-key-prefix` refers to the key prefix of the source cluster (found in its `AppBinding`), and the `--key-prefix` refers to the key prefix of the new target cluster (found in its own `AppBinding`). Make sure to set `--force=true` in both same-cluster and cross-cluster restore scenarios.

So, the final `params` to set in the `RestoreSession` for a cross-cluster restore would look like:

```bash
- --force=${force:=true}
- --key-prefix=${keyPrefix:=k8s.<restore-cluster-uid>.<vault-namespace>.<vault-name>}
- --old-key-prefix=${oldKeyPrefix:=k8s.<backup-cluster-uid>.<vault-namespace>.<vault-name>}
```

Let's create a secret called `encrypt-secret` with the Restic password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD 
secret "encrypt-secret" created
```

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
    name: restore-vault
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
          keyPrefix: k8s.c977fff4-e3b5-4232-8e3e-3bb52106e057.demo.restore-vault
          force: "true"
          oldKeyPrefix: k8s.4a6d4bda-4c08-49a5-b708-5e6d4b2f10f3.demo.vault
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

Then after restoreSession get successful, scale everything back up.

```bash
kubectl scale deploy <kubevault-operator-deployment-name> -n <ns> --replicas=1
kubectl scale sts <vaultServer-statefulSet-name> -n <ns> --replicas=3
```

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
- Read about step-by-step Backup procedure [here](/docs/guides/backup-restore-kubestash/backup.md)