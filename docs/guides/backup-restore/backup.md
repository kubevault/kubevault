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

- [AppBinding](/docs/concepts/vault-server-crds/appbinding.md)
- [Function](https://stash.run/docs/latest/concepts/crds/function/)
- [Task](https://stash.run/docs/latest/concepts/crds/task/)
- [BackupConfiguration](https://stash.run/docs/latest/concepts/crds/backupconfiguration/)
- [RestoreSession](https://stash.run/docs/latest/concepts/crds/restoresession/)


## Deploy Vault using KubeVault

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

````bash
# create demo namespace
$ kubectl create ns demo
namespace/demo created
````

We're going to use Kubernetes secret to store the unseal-keys & root-token. A sample `VaultServer` manifest file may look like this:

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: 1.10.3
  replicas: 3
  allowedSecretEngines:
    namespaces:
      from: All
  backend:
    raft:
      storage:
        storageClassName: "standard"
        resources:
          requests:
            storage: 1Gi
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  monitor:
    agent: prometheus.io
    prometheus:
      exporter:
        resources: {}
  terminationPolicy: WipeOut
```

Now, let's deploy the `VaultServer`:

```bash
$ kubectl apply -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/backup-restore/vaultserver.yaml
vaultserver.kubevault.com/vault created
```

`KubeVault` operator will create a `AppBinding` CRD on `VaultServer` deployment, which contains the necessary information
to take backup of the Vault instances. It'll have the same name & be created on the same namespace as the `Vault`. 
Read more about `AppBinding` [here](/docs/concepts/vault-server-crds/appbinding.md).

```bash
$ kubectl get appbinding -n demo vault -oyaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault
  namespace: demo
spec:
  appRef:
    apiGroup: kubevault.com
    kind: VaultServer
    name: vault
    namespace: demo
  clientConfig:
    service:
      name: vault
      port: 8200
      scheme: http
  parameters:
    apiVersion: config.kubevault.com/v1alpha1
    backend: raft
    backupTokenSecretRef:
      name: vault-backup-token
    kind: VaultServerConfiguration
    kubernetes:
      serviceAccountName: vault
      tokenReviewerServiceAccountName: vault-k8s-token-reviewer
      usePodServiceAccountForCSIDriver: true
    path: kubernetes
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
    unsealer:
      mode:
        kubernetesSecret:
          secretName: vault-keys
      secretShares: 5
      secretThreshold: 3
    vaultRole: vault-policy-controller
```

Now, let's wait until all the vault pods come up & `VaultServer` phase becomes `Ready`.

```bash
$ kubectl get pods -n demo
NAME      READY   STATUS    RESTARTS   AGE
vault-0   2/2     Running   0          2m8s
vault-1   2/2     Running   0          91s
vault-2   2/2     Running   0          65s
```

```bash
$ kubectl get vaultserver -n demo
NAME    REPLICAS   VERSION   STATUS   AGE
vault   3          1.10.3    Ready    2m50s
```

At this stage, we've successfully deployed `Vault` using `KubeVault` operator & ready for taking `Backup`.

Before, taking the backup, let's write some data in a `KV` secret engine. Let's export the necessary environment variables & port-forward from `vault` service 
or exec into the vault pod in order to interact with it.

```bash
$ export VAULT_TOKEN=(kubectl vault root-token get vaultserver vault -n demo --value-only)
$ export VAULT_ADDR='http://127.0.0.1:8200'
$ kubectl port-forward -n demo svc/vault 8200
```

We can see the currently enabled list of secret engines. 

```bash
$ vault secrets list
Path                                       Type         Accessor              Description
----                                       ----         --------              -----------
cubbyhole/                                 cubbyhole    cubbyhole_bb7c56f9    per-token private secret storage
identity/                                  identity     identity_fa8431fa     identity store
k8s.kubevault.com.kv.demo.vault-health/    kv           kv_5129d194           n/a
sys/                                       system       system_c7e0879a       system endpoints used for control, policy and debugging
```

Let's enable a `KV` type secret engine:

```bash
$ vault secrets enable kv
Success! Enabled the kv secrets engine at: kv/
```

Write some dummy data in the secret engine path:

```bash
$ vault kv put kv/name name=appscode
Success! Data written to: kv/name
```

Verify data written in `KV` secret engine:

```bash
$ vault kv get kv/name
==== Data ====
Key     Value
---     -----
name    appscode
```

### Prepare Backend

We are going to store our backed up data into a GCS bucket. We have to create a Secret with necessary credentials and a Repository crd to use this backend. If you want to use a different backend, 
please read the respective backend configuration doc from [here](https://stash.run/docs/v2022.12.11/guides/backends/overview/).

#### Create Secret

Let’s create a secret called `gcs-secret` with access credentials to our desired GCS bucket,

```bash
$ echo -n 'restic-pass' > RESTIC_PASSWORD
$ echo -n 'project-id' > GOOGLE_PROJECT_ID
$ cat sa.json > GOOGLE_SERVICE_ACCOUNT_JSON_KEY

$ kubectl create secret generic -n demo gcs-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
```

Now, we are ready to backup our workload’s data to our desired backend.

#### Create Repository

Now, create a `Repository` using this secret. Below is the YAML of Repository crd we are going to create,

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: gcp-demo-repo
  namespace: demo
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: demo-vault
    storageSecretName: repository-creds
  usagePolicy:
    allowedNamespaces:
      from: Same
  wipeOut: false
```

```bash
$ kbuectl apply -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/backup-restore/repository.yaml
```

Now, we are ready to backup our sample data into this backend.

### Backup

We have to create a BackupConfiguration crd targeting the stash-demo StatefulSet that we have deployed earlier.
Stash will inject a sidecar container into the target. It will also create a CronJob to take periodic 
backup of /source/data directory of the target.

#### Create BackupConfiguration

Below is the YAML of the BackupConfiguration crd that we are going to create,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: demo-backup
  namespace: demo
spec:
  driver: Restic
  repository:
    name: gcp-demo-repo
    namespace: demo
  schedule: "*/5 * * * *"
  timeOut: 2h
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
  retentionPolicy:
    name: 'keep-last-5'
    keepLast: 5
    prune: true
```

Here,
- `spec.repository` refers to the Repository object gcs-repo that holds backend information.
- `spec.schedule` is a cron expression that indicates `BackupSession` will be created at 5 minute interval.
- `spec.target.ref` refers to the `AppBinding` of the `VaultServer`.

Let’s create the BackupConfiguration crd we have shown above,

```bash
$ kubectl apply -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/backup-restore/backup-configuration.yaml
```

#### Verify Backup Setup Successful

If everything goes well, the phase of the BackupConfiguration should be Ready. 
The Ready phase indicates that the backup setup is successful. Let’s verify the Phase of the 
BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo

NAME          TASK                  SCHEDULE      PAUSED   PHASE   AGE
demo-backup   vault-backup-1.10.3   */5 * * * *   true     Ready   92m

```

#### Verify Cronjob

```bash
$ kubectl get cronjob -n demo

NAME                         SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
stash-trigger--demo-backup   */5 * * * *   True      0        <none>          93m
```

#### Wait for BackupSession

The demo-backup  CronJob will trigger a backup on each scheduled slot by creating a BackupSession crd. 
The sidecar container watches for the BackupSession crd. When it finds one, it will take backup immediately.

Wait for the next schedule for backup. Run the following command to watch BackupSession crd,

```bash
kubectl get backupsession -n demo

NAME                INVOKER-TYPE          INVOKER-NAME   PHASE       DURATION   AGE
demo-backup-s2kwg   BackupConfiguration   demo-backup    Succeeded   39s        58s
```

#### Verify Backup

Once a backup is complete, Stash will update the respective Repository crd to reflect the backup. 
Check that the repository gcs-repo has been updated by the following command,

```bash
kubectl get repository -n demo

NAME            INTEGRITY   SIZE         SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
gcp-demo-repo   true        75.867 KiB   1                11m                      11m
```


Now, if we navigate to the GCS bucket, we are going to see backed up data is uploaded successfully:

<figure align="center">
 <img alt="Vault Backup" src="/docs/images/guides/backup-restore/backup.png">
  <figcaption align="center">Fig: Vault Backup</figcaption>
</figure>


Up next:
- Read about step-by-step Restore procedure [here](/docs/guides/backup-restore/restore.md)