---
title: Vault Backup Restore Overview
menu:
  docs_v2026.2.27:
    identifier: backup-backup-restore-guides-kubestash
    name: Backup
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

# Backup Vault Cluster using KubeStash

This guide will show you how you can take backup of your Vault cluster with KubeStash.

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
- [BackupConfiguration](https://kubestash.com/docs/latest/concepts/crds/backupconfiguration/)
- [BackupSession](https://kubestash.com/docs/latest/concepts/crds/backupsession/)
- [Addon](https://kubestash.com/docs/latest/concepts/crds/addon/)
- [Function](https://kubestash.com/docs/latest/concepts/crds/function/)
- [Task](https://kubestash.com/docs/latest/concepts/crds/addon/#task-specification)


To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy Sample PostgreSQL Database

Let's deploy a sample `PostgreSQL` database.

**Create PostgreSQL CR:**

At first, we will create `user.conf` file containing required configuration settings. You need to set max_connections according to your needs. The default value is 100, but you can increase it if you want to allow more connections to the database.
To know more about this configuration file, check [here](/docs/guides/postgres/configuration/using-config-file.md)
```ini
$ cat user.conf
max_connections=200
```

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo pg-configuration --from-file=./user.conf
secret/pg-configuration created
```

Below is the YAML of a sample `PostgreSQL` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: postgres-quickstart
  namespace: demo
spec:
  version: "16.13"
  storageType: Durable
  replicas: 3
  configuration:
    secretName: pg-configuration
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Create the above `PostgreSQL` CR,

```bash
$ kubectl apply -f postgres-quickstart.yaml
postgres.kubedb.com/postgres-quickstart created
```

KubeDB will deploy a `PostgreSQL` database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get pg -n demo postgres-quickstart
NAME                  VERSION    STATUS   AGE
postgres-quickstart   16.13      Ready    5m1s
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,
Verify that the `AppBinding` has been created successfully using the following command,
```bash
$ kubectl get secret -n demo 
NAME                          TYPE                       DATA   AGE
postgres-quickstart-auth      kubernetes.io/basic-auth   2      5m20s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=postgres-quickstart
NAME                          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
postgres-quickstart           ClusterIP   10.96.23.177   <none>        5432/TCP,2379/TCP            5m55s
postgres-quickstart-pods      ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP   5m55s
postgres-quickstart-standby   ClusterIP   10.96.26.118   <none>        5432/TCP                     5m55s

$ kubectl get appbindings -n demo
NAME                       TYPE                  VERSION   AGE
postgres-quickstart        kubedb.com/postgres   16.1      9m30s
```

The PostgreSQL storage backend does not automatically create the table. You need to create the schema and indexes.

```sql
kubectl exec -n demo -it postgres-quickstart-0 -- psql -U postgres
CREATE DATABASE vault;
\c vault;
   
CREATE TABLE vault_kv_store (
  parent_path TEXT COLLATE "C" NOT NULL,
  path        TEXT COLLATE "C",
  key         TEXT COLLATE "C",
  value       BYTEA,
  CONSTRAINT pkey PRIMARY KEY (path, key)
);

CREATE INDEX parent_path_idx ON vault_kv_store (parent_path);
```

Store for HAEnabled backend:

```sql
CREATE TABLE vault_ha_locks (
  ha_key                                      TEXT COLLATE "C" NOT NULL,
  ha_identity                                 TEXT COLLATE "C" NOT NULL,
  ha_value                                    TEXT COLLATE "C",
  valid_until                                 TIMESTAMP WITH TIME ZONE NOT NULL,
  CONSTRAINT ha_key PRIMARY KEY (ha_key)
);
```

**Create Connection Secret**

```bash
export PG_PASS=$(kubectl get secret -n demo postgres-quickstart-auth -o jsonpath='{.data.password}' | base64 -d)
kubectl create secret generic my-postgres-conn -n demo \
  --from-literal=connection_url="postgres://postgres:${PG_PASS}@postgres-quickstart.demo.svc:5432/vault?sslmode=disable"
```

## Deploy Vault using KubeVault

We're going to use Kubernetes secret to store the unseal-keys & root-token. A sample `VaultServer` with Postgres backend manifest file may look like this:

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  terminationPolicy: WipeOut
  replicas: 3
  version: 1.18.4 
  serviceTemplates:
  - alias: vault
    metadata:
      annotations:
        name: vault
    spec:
      type: NodePort
  backend:
    postgresql:
      credentialSecretRef:
        name: my-postgres-conn
      table: vault_kv_store
      haEnabled: "true"
      haTable: vault_ha_locks
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
```

Now, let's deploy the `VaultServer`:

```bash
$ kubectl apply -f vaultserver.yaml
vaultserver.kubevault.com/vault created
```

`KubeVault` operator will create a `AppBinding` CRD on `VaultServer` deployment, which contains the necessary information
to take backup of the Vault instances. It'll have the same name & be created on the same namespace as the `Vault`. 
Read more about `AppBinding` [here](/docs/v2026.2.27/concepts/vault-server-crds/appbinding).

```bash
$ kubectl get appbinding -n demo vault -oyaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  creationTimestamp: "2026-03-30T12:45:18Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: vault
    app.kubernetes.io/managed-by: kubevault.com
    app.kubernetes.io/name: vaultservers.kubevault.com
  name: vault
  namespace: demo
  ownerReferences:
    - apiVersion: kubevault.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: VaultServer
      name: vault
      uid: aa35b2d6-2452-4e9d-8c17-cdbefdfb1ad0
  resourceVersion: "110680"
  uid: 2385966e-235c-4d56-87d1-44a890642a57
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
    backend: postgresql
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
              value: k8s.62e86b47-6e7c-4849-89fb-88062a22f451.demo.vault
        restoreTask:
          name: vault-restore-1.10.3
          params:
            - name: keyPrefix
              value: k8s.62e86b47-6e7c-4849-89fb-88062a22f451.demo.vault
    unsealer:
      mode:
        kubernetesSecret:
          secretName: vault-keys
      secretShares: 5
      secretThreshold: 3
    vaultRole: vault-policy-controller
  type: VaultServer
  version: 1.18.4

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
vault   3          1.18.4    Ready    2m50s
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

We are going to store our backed up data into a S3 bucket. We have to create a Secret with necessary credentials to use this backend. If you want to use a different backend, 
please read the respective backend configuration doc from [here](https://kubestash.com/docs/v2026.2.26/guides/backends/overview/).

#### Create Secret

Let's create a secret called `s3-secret` with access credentials to our desired s3 bucket,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ echo -n '<your-aws-access-key-id-here>' > AWS_ACCESS_KEY_ID
$ echo -n '<your-aws-secret-access-key-here>' > AWS_SECRET_ACCESS_KEY
$ kubectl create secret generic -n demo s3-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./AWS_ACCESS_KEY_ID \
    --from-file=./AWS_SECRET_ACCESS_KEY
secret/s3-secret created
```

Now, we are ready to backup our workload’s data to our desired backend.

**Create BackupStorage:**

Now, create a `BackupStorage` using this secret. Below is the YAML of `BackupStorage` CR we are going to create,

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: s3-storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      endpoint: https://s3.us-east-2.amazonaws.com 
      bucket: vault-backup
      region: us-east-2
      prefix: backup-test
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: Delete
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f backupstorage.yaml
backupstorage.storage.kubestash.com/s3-storage created
```
You can verify that the `BackupStorage` has been created successfully using the following command,
```bash
kubectl get backupstorage -n demo
NAME         PROVIDER   DEFAULT   DELETION-POLICY   TOTAL-SIZE   PHASE   AGE
s3-storage   s3                   Delete                         Ready   42s
```
Now, we are ready to backup our database to our desired backend.

**Create RetentionPolicy:**

Now, let's create a `RetentionPolicy` to specify how the old Snapshots should be cleaned up.

Below is the YAML of the `RetentionPolicy` object that we are going to create,

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: RetentionPolicy
metadata:
  name: demo-retention
  namespace: demo
spec:
  default: true
  failedSnapshots:
    last: 2
  maxRetentionPeriod: 2mo
  successfulSnapshots:
    last: 5
  usagePolicy:
    allowedNamespaces:
      from: All
```

Let’s create the above `RetentionPolicy`,

```bash
$ kubectl apply -f retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```
You can verify that the `RetentionPolicy` has been created successfully using the following command,
```bash
kubectl get retentionpolicy -A
NAMESPACE   NAME             MAX-RETENTION-PERIOD   DEFAULT   AGE
demo        demo-retention   2mo                    true      51s
```

Now, we are ready to backup our sample data into this backend.

### Backup

We have to create a `BackupConfiguration` targeting respective vaultServer `vault`. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.

At first, we need to create a secret with a Restic password for backup data encryption.

**Create Secret:**

Let's create a secret called `encrypt-secret` with the Restic password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD 
secret "encrypt-secret" created
```

Below is the YAML for `BackupConfiguration` CR to backup the `vault` vaultServer that we have deployed earlier,
Get the KeyPrefix value using this command:
```bash
$ kubectl get appbinding -n demo <vaultServer-name> -o jsonpath='{.spec.parameters.stash.addon.backupTask.params[?(@.name=="keyPrefix")].value}'
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-vault-backup
  namespace: demo
spec:
  target:
    apiGroup: appcatalog.appscode.com
    kind: AppBinding
    namespace: demo
    name: vault
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: frequent-backup
      scheduler:
        schedule: "*/5 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: s3-vault-repo
          backend: s3-backend
          directory: /vault/data/backup
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: vault-addon
        tasks:
          - name: vault-backup
            params:
              keyPrefix: <vault-appbinding-key-prefix-value>
```

Here,
- `.spec.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.
- `.spec.target` refers to the targeted `vault` VaultServer that we created earlier.

Let’s create the BackupConfiguration crd we have shown above,

```bash
$ kubectl apply -f backup-configuration.yaml
backupconfiguration.core.kubestash.com/sample-vault-backup created
```

#### Verify Backup Setup Successful

If everything goes well, the phase of the BackupConfiguration should be Ready. 
The Ready phase indicates that the backup setup is successful. Let’s verify the Phase of the 
BackupConfiguration,

```bash
$ kubectl get backupconfiguration -A
NAMESPACE   NAME                     PHASE   PAUSED   AGE
demo        sample-vault-backup     Ready   true     47h
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME                  INTEGRITY   SNAPSHOT-COUNT   SIZE     PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-vault-repo                     0                0 B      Ready                            3m
```

#### Verify Cronjob

```bash
$ kubectl get cronjob -n demo

NAME                                          SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-vault-backup-frequent-backup   */5 * * * *   True      0        <none>          93m
```

#### Wait for BackupSession

The demo-backup  CronJob will trigger a backup on each scheduled slot by creating a BackupSession crd. 
The sidecar container watches for the BackupSession crd. When it finds one, it will take backup immediately.

Wait for the next schedule for backup. Run the following command to watch BackupSession crd,

```bash
kubectl get backupsession -n demo

NAME                                             INVOKER-TYPE          INVOKER-NAME          PHASE       DURATION   AGE
sample-vault-backup-frequent-backup-1774961500   BackupConfiguration   sample-vault-backup   Succeeded   48s        2d
```

#### Verify Backup

Once a backup is complete, Stash will update the respective Repository crd to reflect the backup. 
Check that the repository gcs-repo has been updated by the following command,

```bash
kubectl get repository -n demo

NAME            INTEGRITY   SIZE         SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
s3-vault-repo   true        75.867 KiB   1                11m                      11m
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=s3-vault-repo
NAME                                                                  REPOSITORY      SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
s3-vault-repo-postgres-quickstart-backup-frequent-backup-1725449400   s3-vault-repo   frequent-backup   2026-03-27T11:35:02Z   Delete            Succeeded   16h
```


Now, if we navigate to the s3 bucket, we can see backed up data is uploaded successfully:

<figure align="center">

 <img alt="Vault Backup" src="/docs/v2026.2.27/images/guides/backup-restore/s3-ckup.png">

  <figcaption align="center">Fig: Vault Backup (KubeStash)</figcaption>

</figure>


Up next:
- Read about step-by-step Restore procedure [here](/docs/v2026.2.27/guides/backup-restore-kubestash/restore)