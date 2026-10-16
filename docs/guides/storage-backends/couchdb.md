---
title: Vault Server CouchDB Backend
menu:
  docs_{{ .version }}:
    identifier: vault-server-couchdb-backend
    name: CouchDB
    parent: storage-backends-guides
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Deploy VaultServer with CouchDB Backend

You can easily deploy and manage [OpenBao](https://openbao.org/) in the Kubernetes cluster using KubeVault operator. In this tutorial, we are going to deploy Vault on the Kubernetes cluster using KubeVault operator, with [Apache CouchDB](https://couchdb.apache.org/) as the storage backend.

![Vault Server](/docs/images/guides/vault-server/vault_server_guide.svg)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy VaultServer with CouchDB Backend

### Deploy Sample CouchDB

Let's deploy a sample single-node CouchDB server.

Below is the YAML of the `Service` and `Deployment` that we are going to create for this tutorial:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: couchdb
  namespace: demo
  labels:
    app: couchdb
spec:
  ports:
  - port: 5984
    name: http
  selector:
    app: couchdb
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: couchdb
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: couchdb
  template:
    metadata:
      labels:
        app: couchdb
    spec:
      containers:
      - name: couchdb
        image: couchdb:3.3.3
        imagePullPolicy: IfNotPresent
        env:
        - name: COUCHDB_USER
          value: "admin"
        - name: COUCHDB_PASSWORD
          value: "password"
        ports:
        - containerPort: 5984
          name: http
        readinessProbe:
          httpGet:
            path: /_up
            port: 5984
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          limits:
            cpu: "500m"
            memory: "512Mi"
          requests:
            cpu: "100m"
            memory: "256Mi"
```

Create the above resources,

```bash
$ kubectl apply -f couchdb.yaml
service/couchdb created
deployment.apps/couchdb created

$ kubectl wait -n demo --for=condition=Ready pod -l app=couchdb --timeout=5m
pod/couchdb-6b8f7c9d5-k8w2s condition met
```

**Create the Database:**

The CouchDB storage backend does not create the database, so it must exist before Vault starts. Let's create a database called `vault`,

```bash
$ kubectl exec -n demo deploy/couchdb -- \
    curl -s -X PUT http://admin:password@127.0.0.1:5984/vault
{"ok":true}

$ kubectl exec -n demo deploy/couchdb -- \
    curl -s http://admin:password@127.0.0.1:5984/_all_dbs
["_users","vault"]
```

**Create Credential Secret:**

The credential secret holds the CouchDB login in the keys `username` and `password`.

```bash
$ kubectl create secret generic couchdb-cred -n demo \
    --from-literal=username=admin \
    --from-literal=password=password
secret/couchdb-cred created
```

## Deploy Vault using KubeVault

We're going to use Kubernetes secret to store the unseal-keys & root-token. A sample `VaultServer` with CouchDB backend manifest file may look like this:

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
    couchdb:
      endpoint: "http://couchdb.demo.svc:5984/vault"
      credentialSecretRef:
        name: couchdb-cred
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```

The `endpoint` is the full URL of the database, including the database name. The CouchDB storage backend does not support high availability, so the `VaultServer` runs with a single replica. It has no TLS options either: an `https://` endpoint only works if the server certificate is signed by a CA that the Vault pod already trusts.

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
    backend: couchdb
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

At this stage, we've successfully deployed `Vault` using `KubeVault` operator with CouchDB as the storage backend.

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
Storage Type    couchdb
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
$ kubectl delete secret -n demo couchdb-cred
$ kubectl delete -f couchdb.yaml
$ kubectl delete ns demo
```

Delete the `VaultServer` before CouchDB. With `terminationPolicy: WipeOut`, the operator connects to CouchDB to remove the documents Vault stored there, so CouchDB has to be reachable until the `VaultServer` is gone.

To see all the fields you can set under `spec.backend.couchdb`, please check the [CouchDB storage backend](/docs/concepts/vault-server-crds/storage/couchdb.md) concept.

For more details on how to interact with the vault server, please check the [Vault Server guide](/docs/guides/vault-server/vault-server.md).
