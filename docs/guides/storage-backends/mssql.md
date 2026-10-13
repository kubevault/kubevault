---
title: Vault Server MSSQL Backend
menu:
  docs_{{ .version }}:
    identifier: vault-server-mssql-backend
    name: MSSQL
    parent: storage-backends-guides
    weight: 140
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Deploy VaultServer with MSSQL Backend

You can easily deploy and manage [HashiCorp Vault](https://www.vaultproject.io/) in the Kubernetes cluster using KubeVault operator. In this tutorial, we are going to deploy Vault on the Kubernetes cluster using KubeVault operator, with [Microsoft SQL Server](https://www.microsoft.com/sql-server) as the storage backend, managed by KubeDB.

![Vault Server](/docs/images/guides/vault-server/vault_server_guide.svg)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy VaultServer with MSSQL Backend

> This tutorial uses [KubeDB](https://kubedb.com/docs/latest/setup/) to run SQL Server. The KubeDB operator must be installed in your cluster.

### Deploy Sample MSSQLServer Database

Let's deploy a sample `MSSQLServer` database.

**Create MSSQLServer CR:**

Below is the YAML of a sample `MSSQLServer` CR that we are going to create for this tutorial. Setting `ACCEPT_EULA` to `Y` means that you accept the end-user license agreement of Microsoft SQL Server.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 1
  storageType: Durable
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Create the above `MSSQLServer` CR,

```bash
$ kubectl apply -f mssql.yaml
mssqlserver.kubedb.com/mssql created
```

KubeDB will deploy a SQL Server database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database. Let's wait until the database is ready to use,

```bash
$ kubectl wait -n demo --for=condition=Ready mssqlserver/mssql --timeout=5m
mssqlserver.kubedb.com/mssql condition met
```

Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=mssql
NAME            TYPE                       DATA   AGE
mssql-auth      kubernetes.io/basic-auth   2      4m20s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=mssql
NAME        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
mssql       ClusterIP   10.96.18.204   <none>        1433/TCP   4m35s
mssql-pods  ClusterIP   None           <none>        1433/TCP   4m35s
```

The `mssql-auth` secret holds the `username` and `password` that Vault is going to use. Vault creates the database (`openbao`) and the table on its own, if they do not exist.

## Deploy Vault using KubeVault

We're going to use Kubernetes secret to store the unseal-keys & root-token. A sample `VaultServer` with MSSQL backend manifest file may look like this:

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: sigilr-2.6.1.2
  replicas: 1
  allowedSecretEngines:
    namespaces:
      from: All
  backend:
    mssql:
      server: "mssql.demo.svc"
      port: "1433"
      credentialSecretRef:
        name: mssql-auth
      database: "openbao"
      table: "openbao"
      schema: "dbo"
      appName: "openbao"
      connectionTimeout: "30"
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```

The MSSQL storage backend does not support high availability, so the `VaultServer` runs with a single replica. It has no TLS options either, so the connection to SQL Server is not configured for TLS.

Now, let's deploy the `VaultServer`:

```bash
$ kubectl apply -f vaultserver.yaml
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
    backend: mssql
    backupTokenSecretRef:
      name: vault-backup-token
    deploymentMode: Local
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
  version: sigilr-2.6.1.2
```

Now, let's wait until all the vault pods come up & `VaultServer` phase becomes `Ready`.

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=vault
NAME      READY   STATUS    RESTARTS   AGE
vault-0   2/2     Running   0          2m8s
```

```bash
$ kubectl get vaultserver -n demo
NAME    REPLICAS   VERSION          STATUS   AGE
vault   1          sigilr-2.6.1.2   Ready    2m50s
```

At this stage, we've successfully deployed `Vault` using `KubeVault` operator with MSSQL as the storage backend.

Let's write some data in a `KV` secret engine. Let's export the necessary environment variables & port-forward from `vault` service
or exec into the vault pod in order to interact with it.

```bash
$ export VAULT_TOKEN=$(kubectl vault root-token get vaultserver vault -n demo --value-only)
$ export VAULT_ADDR='http://127.0.0.1:8200'
$ kubectl port-forward -n demo svc/vault 8200
```
Now check whether Vault server can be accessed (run this in another terminal):

```bash
$ vault status
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    5
Threshold       3
Version         2.6.1
Storage Type    mssql
Cluster Name    vault-cluster-3f8a1c92
Cluster ID      56c3a681-f799-2310-9813-d67877d25450
HA Enabled      false
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

## Cleanup

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete vaultserver -n demo vault
$ kubectl delete mssqlserver -n demo mssql
$ kubectl delete ns demo
```

Delete the `VaultServer` before the `MSSQLServer` database. With `terminationPolicy: WipeOut`, the operator connects to SQL Server to remove the data Vault stored there, so the database has to be reachable until the `VaultServer` is gone.

To see all the fields you can set under `spec.backend.mssql`, please check the [MSSQL storage backend](/docs/concepts/vault-server-crds/storage/mssql.md) concept.

For more details on how to interact with the vault server, please check the [Vault Server guide](/docs/guides/vault-server/vault-server.md).
