---
title: Vault Backup Restore Overview
menu:
  docs_{{ .version }}:
    identifier: restore-backup-restore-guides
    name: Restore
    parent: backup-restore-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Restore Vault Cluster using Stash

This guide will show you how you can restore your Vault cluster with Stash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using Minikube.
- Install KubeVault in your cluster following the steps [here](/docs/setup/README.md).
- Install Stash in your cluster following the steps [here](https://stash.run/docs/latest/setup/).
- Install Stash `kubectl` plugin following the steps [here](https://stash.run/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how Stash backup and restore Vault cluster, please check the following concept section [here](/docs/concepts/backup-restore/overview.md).

You have to be familiar with following custom resources:

- [AppBinding](/docs/concepts/vault-server-crds/appbinding.md)
- [Function](https://stash.run/docs/latest/concepts/crds/function/)
- [Task](https://stash.run/docs/latest/concepts/crds/task/)
- [BackupConfiguration](https://stash.run/docs/latest/concepts/crds/backupconfiguration/)
- [RestoreSession](https://stash.run/docs/latest/concepts/crds/restoresession/)

You may restore a Vault snapshot into the same Vault cluster from which snapshot was taken or into a 
completely new Vault deployment.

### Restore Snapshot for same Vault

Follow this guideline, if you want to restore a snapshot into the same Vault cluster. 
Vault cluster must be `Initialized` & `Unsealed` before trying to restore the snapshot.

Then, simply you can create a `RestoreSession` to restore the snapshot. A sample `RestoreSession` YAML may look like this:

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: vault-restore-session
  namespace: demo
spec:
  repository:
    name: gcp-demo-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: vault
  runtimeSettings:
    container:
      securityContext:
        runAsUser: 0
        runAsGroup: 0
  rules:
  - snapshots: [latest]

```

#### Create RestoreSession

Create the `RestoreSession` for restore the snapshot:

```yaml
$ kubectl apply -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/backup-restore/restore-session.yaml
```

Now, wait for `RestoreSession` to succeed:

```bash
$ kubectl get restoresession -n demo

NAME                    REPOSITORY      PHASE       DURATION   AGE
vault-restore-session   gcp-demo-repo   Succeeded   19s        27s

```

Once the `RestoreSession` is Succeeded, snapshot will be successfully restored into the Vault cluster. 

### Restore Snapshot for different Vault

Follow this guideline, if you want to restore a snapshot into a different Vault cluster.

You need to deploy the new `Vault` cluster & it must be `Initialized` & `Unsealed`. This `Vault` has a
completely different set of `unseal keys` & `root token` from the `Vault` from which the snapshot was taken.

`Vault` snapshot carries the signature of `unseal keys`. So, we need to restore the snapshot forcefully & to bypass 
this, we need to modify our restore `function` accordingly. A `Function` CRD may look like this:

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: Function
metadata:
  name: vault-restore-1.10.3
spec:
  args:
  - restore-vault
  - --provider=${REPOSITORY_PROVIDER:=}
  - --bucket=${REPOSITORY_BUCKET:=}
  - --endpoint=${REPOSITORY_ENDPOINT:=}
  - --region=${REPOSITORY_REGION:=}
  - --path=${REPOSITORY_PREFIX:=}
  - --storage-secret-name=${REPOSITORY_SECRET_NAME:=}
  - --storage-secret-namespace=${REPOSITORY_SECRET_NAMESPACE:=}
  - --scratch-dir=/tmp
  - --enable-cache=${ENABLE_CACHE:=true}
  - --max-connections=${MAX_CONNECTIONS:=0}
  - --wait-timeout=${waitTimeout:=300}
  - --hostname=${HOSTNAME:=}
  - --source-hostname=${SOURCE_HOSTNAME:=}
  - --interim-data-dir=${INTERIM_DATA_DIR}
  - --namespace=${NAMESPACE:=default}
  - --appbinding=${TARGET_NAME:=}
  - --appbinding-namespace=${TARGET_NAMESPACE:=}
  - --snapshot=${RESTORE_SNAPSHOTS:=}
  - --vault-args=${args:=}
  - --output-dir=${outputDir:=}
  - --license-apiservice=${LICENSE_APISERVICE:=}
  - --force=${force:=false}
  - --key-prefix=${keyPrefix:=}
  - --old-key-prefix=${oldKeyPrefix:=}
  image: stashed/stash-vault:1.10.3
```

Let's take a look at some of the more relevant flags, that we can set to override the existing flags:

```bash
- --force=${force:=false}
- --key-prefix=${keyPrefix:=}
- --old-key-prefix=${oldKeyPrefix:=}
```

By default, the `--force` flag is `false`, so in order to restoring the snapshot into a differnt Vault cluster, 
this must be set to `true`.

Moreover, once the snapshot will be restored, the newly `Vault` will be expecting the older `unseal keys` to unseal itself & 
the new `unseal keys` will not be required/valid anymore. So, we'll also migrate the older `unseal keys` & `root token` in place of
the new `unseal keys` & `root token`.

Since, `Stash` will also take backup of the Vault `unseal keys` & `root token` along with the snapshot, we can get the
older `unseal keys` & `root token`. To correctly get those, we must set the `--old-key-prefix` flag properly.

```bash
- --force=${force:=true}
- --key-prefix=${keyPrefix:=}
- --old-key-prefix=${oldKeyPrefix:=<old-key-prefix>}
```

`KeyPrefix` will be generated by the following structure by `KubeVault` operator: 
`k8s.{kubevault.com or cluster UID}.{vault-namespace}.{vault-name}`. In case of Vault deployment using Vault Helm-chart
or if you want to save it with a different prefix, you need to override the `KeyPrefix` section. 

The default `key-prefix`, associated `Task` for `Backup` & `Restore` can be found in the Vault `AppBinding` YAML:

```yaml
stash:
      addon:
        backupTask:
          name: vault-backup-1.10.3
          params:
          - name: keyPrefix
            value: k8s.kubevault.com.demo.vault
        restoreTask:
          name: vault-restore-1.10.3
          params:
          - name: keyPrefix
            value: k8s.kubevault.com.demo.vault

```

Now, we need to apply the changes in our restore `Function` CRD. Now, we can create the `RestoreSession`
to restore the Vault cluster by the similar way mentioned above.

Up next:
- Read about step-by-step Backup procedure [here](/docs/guides/backup-restore/backup.md)