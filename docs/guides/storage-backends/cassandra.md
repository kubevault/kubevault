---
title: Vault Server Cassandra Backend
menu:
  docs_{{ .version }}:
    identifier: vault-server-cassandra-backend
    name: Cassandra
    parent: storage-backends-guides
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Deploy VaultServer with Cassandra Backend

You can easily deploy and manage [OpenBao](https://openbao.org/) in the Kubernetes cluster using KubeVault operator. In this tutorial, we are going to deploy Vault on the Kubernetes cluster using KubeVault operator, with [Apache Cassandra](https://cassandra.apache.org/) as the storage backend, managed by KubeDB.

![Vault Server](/docs/images/guides/vault-server/vault_server_guide.svg)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy VaultServer with Cassandra Backend

> This tutorial uses [KubeDB](https://kubedb.com/docs/latest/setup/) to run Cassandra. The KubeDB operator must be installed in your cluster.

### Deploy Cassandra

Let's deploy a sample `Cassandra` database:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra
  namespace: demo
spec:
  version: "4.1.8"
  topology:
    rack:
      - name: r0
        replicas: 2
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f cassandra.yaml
cassandra.kubedb.com/cassandra created

$ kubectl wait -n demo --for=condition=Ready cassandra/cassandra --timeout=10m
cassandra.kubedb.com/cassandra condition met
```

KubeDB creates an `AppBinding` named `cassandra` for the database, which holds the Cassandra hosts and the `cassandra-auth` secret with the `username` and `password`.

**Create the Keyspace and the Table:**

The Cassandra storage backend does not create the keyspace or the table, so they must exist before Vault starts. KubeDB generates the credentials in the `cassandra-auth` secret.

```bash
$ CASSANDRA_USER=$(kubectl get secret -n demo cassandra-auth -o jsonpath='{.data.username}' | base64 -d)
$ CASSANDRA_PASS=$(kubectl get secret -n demo cassandra-auth -o jsonpath='{.data.password}' | base64 -d)

$ kubectl exec -n demo -i cassandra-rack-r0-0 -c cassandra -- \
    cqlsh -u "$CASSANDRA_USER" -p "$CASSANDRA_PASS" -e 'CREATE KEYSPACE IF NOT EXISTS vault WITH replication = {'"'"'class'"'"': '"'"'SimpleStrategy'"'"', '"'"'replication_factor'"'"': 1};
CREATE TABLE IF NOT EXISTS vault."entries" (
  bucket text,
  key    text,
  value  blob,
  PRIMARY KEY (bucket, key)
) WITH CLUSTERING ORDER BY (key ASC);'
```

`entries` is a reserved word in CQL, so the table name is quoted (`vault."entries"`). The statements are in single quotes so that the shell keeps those double quotes, which is why the CQL string literals are written as `'"'"'`. The string starts right after `-e '` and ends right after the last `;`, because a line break before the first or after the last statement makes `cqlsh` fail with `no viable alternative at input ';'`.

> `SimpleStrategy` with a replication factor of `1` is enough for this tutorial. For production, use a replication strategy and factor that match your cluster.

## Deploy Vault using KubeVault

We're going to use Kubernetes secret to store the unseal-keys & root-token. A sample `VaultServer` with Cassandra backend manifest file may look like this:

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
    cassandra:
      databaseRef:
        name: cassandra
        namespace: demo
      keyspace: "vault"
      table: "entries"
      protocolVersion: "4"
      consistency: "ONE"
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```

The `databaseRef` points to the `AppBinding` that KubeDB created for the database. The operator takes the Cassandra hosts and the credentials from it, so we don't need to set them here. The Cassandra storage backend does not support high availability, so the `VaultServer` runs with a single replica.

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
    backend: cassandra
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

At this stage, we've successfully deployed `Vault` using `KubeVault` operator with Cassandra as the storage backend.

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
Storage Type    cassandra
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
$ kubectl delete cassandra -n demo cassandra
$ kubectl delete ns demo
```

Delete the `VaultServer` before the `Cassandra` database. With `terminationPolicy: WipeOut`, the operator connects to Cassandra to remove the data Vault stored there, so the database has to be reachable until the `VaultServer` is gone.

To see all the fields you can set under `spec.backend.cassandra`, please check the [Cassandra storage backend](/docs/concepts/vault-server-crds/storage/cassandra.md) concept.

For more details on how to interact with the vault server, please check the [Vault Server guide](/docs/guides/vault-server/vault-server.md).
