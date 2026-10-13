---
title: Vault Server Etcd Backend
menu:
  docs_{{ .version }}:
    identifier: vault-server-etcd-backend
    name: Etcd
    parent: storage-backends-guides
    weight: 90
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Deploy VaultServer with Etcd Backend

You can easily deploy and manage [HashiCorp Vault](https://www.vaultproject.io/) in the Kubernetes cluster using KubeVault operator. In this tutorial, we are going to deploy Vault on the Kubernetes cluster using KubeVault operator, with [etcd](https://etcd.io/) as the storage backend, managed by KubeDB.

![Vault Server](/docs/images/guides/vault-server/vault_server_guide.svg)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy VaultServer with Etcd Backend

> This tutorial uses [KubeDB](https://kubedb.com/docs/latest/setup/) to run etcd, and [cert-manager](https://cert-manager.io/docs/installation/) to issue its TLS certificates. Both must be installed in your cluster.

### Deploy Sample Etcd Database

**Create a CA and a cert-manager `Issuer`:**

KubeDB asks [cert-manager](https://cert-manager.io/docs/installation/) to issue the etcd certificates. Let's create a self-signed CA and an `Issuer` backed by it. This CA is only for this tutorial, use your own CA or issuer in production.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout ca.key -out ca.crt -sha256 \
    -subj "/CN=etcd-ca" \
    -addext 'basicConstraints=critical,CA:TRUE' \
    -addext 'keyUsage=critical,keyCertSign,cRLSign'

$ kubectl create secret tls etcd-ca -n demo --cert=ca.crt --key=ca.key
secret/etcd-ca created
```

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: etcd-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: etcd-ca
```

```bash
$ kubectl apply -f etcd-ca-issuer.yaml
issuer.cert-manager.io/etcd-ca-issuer created
```

**Create the Certificates:**

KubeDB expects the server, peer and client certificates of etcd to exist as `Secrets` in the namespace. Let's issue them from the CA above:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: etcd-server-cert
  namespace: demo
spec:
  secretName: etcd-server-cert
  commonName: etcd
  privateKey:
    rotationPolicy: Always
  usages:
    - digital signature
    - key encipherment
    - server auth
    - client auth
  dnsNames:
    - etcd
    - etcd.demo
    - etcd.demo.svc
    - etcd.demo.svc.cluster.local
    - etcd-pods.demo
    - etcd-pods.demo.svc
    - etcd-pods.demo.svc.cluster.local
    - "*.etcd-pods.demo.svc"
    - "*.etcd-pods.demo.svc.cluster.local"
    - localhost
  ipAddresses:
    - 127.0.0.1
  issuerRef:
    name: etcd-ca-issuer
    kind: Issuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: etcd-peer-cert
  namespace: demo
spec:
  secretName: etcd-peer-cert
  commonName: etcd
  privateKey:
    rotationPolicy: Always
  usages:
    - digital signature
    - key encipherment
    - server auth
    - client auth
  dnsNames:
    - etcd
    - etcd.demo
    - etcd.demo.svc
    - etcd.demo.svc.cluster.local
    - etcd-pods.demo
    - etcd-pods.demo.svc
    - etcd-pods.demo.svc.cluster.local
    - "*.etcd-pods.demo.svc"
    - "*.etcd-pods.demo.svc.cluster.local"
    - localhost
  ipAddresses:
    - 127.0.0.1
  issuerRef:
    name: etcd-ca-issuer
    kind: Issuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: etcd-client-cert
  namespace: demo
spec:
  secretName: etcd-client-cert
  commonName: root
  privateKey:
    rotationPolicy: Always
  usages:
    - digital signature
    - key encipherment
    - client auth
  issuerRef:
    name: etcd-ca-issuer
    kind: Issuer
    group: cert-manager.io
```

```bash
$ kubectl apply -f etcd-certificates.yaml
certificate.cert-manager.io/etcd-server-cert created
certificate.cert-manager.io/etcd-peer-cert created
certificate.cert-manager.io/etcd-client-cert created
```

**Create Etcd CR:**

Below is the YAML of a sample `Etcd` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 1
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: etcd-ca-issuer
  deletionPolicy: WipeOut
```

Create the above `Etcd` CR,

```bash
$ kubectl apply -f etcd.yaml
etcd.kubedb.com/etcd created
```

KubeDB will deploy an `Etcd` database according to the above specification. It will also create the necessary `Secrets`, `Services` and an `AppBinding` to access the database. Let's wait until the database is ready to use,

```bash
$ kubectl wait -n demo --for=condition=Ready etcd/etcd --timeout=5m
etcd.kubedb.com/etcd condition met

$ kubectl get appbinding -n demo etcd
NAME   TYPE               VERSION   AGE
etcd   kubedb.com/etcd    3.6.4     3m21s
```

## Deploy Vault using KubeVault

We're going to use Kubernetes secret to store the unseal-keys & root-token. A sample `VaultServer` with Etcd backend manifest file may look like this:

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: sigilr-2.6.1.2
  replicas: 3
  allowedSecretEngines:
    namespaces:
      from: All
  backend:
    etcd:
      databaseRef:
        name: etcd
        namespace: demo
      etcdApi: "v3"
      path: "/vault"
      haEnabled: true
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```

The `databaseRef` points to the `AppBinding` that KubeDB created for the database. The operator takes the etcd address, the credentials, the TLS CA and the client certificate from it, so we don't need to set them here. With `haEnabled: true`, Vault uses etcd leases for leader election, so the `VaultServer` can run with more than one replica: one pod is the active node and the others are standby.

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
    backend: etcd
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
vault-1   2/2     Running   0          91s
vault-2   2/2     Running   0          65s
```

```bash
$ kubectl get vaultserver -n demo
NAME    REPLICAS   VERSION          STATUS   AGE
vault   3          sigilr-2.6.1.2   Ready    2m50s
```

At this stage, we've successfully deployed `Vault` using `KubeVault` operator with Etcd as the storage backend.

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
Storage Type    etcd
Cluster Name    vault-cluster-3f8a1c92
Cluster ID      56c3a681-f799-2310-9813-d67877d25450
HA Enabled      true
HA Cluster      https://vault.demo.svc:8201
HA Mode         active
Active Since    2026-10-12T11:41:37.13040169Z
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
$ kubectl delete etcd -n demo etcd
$ kubectl delete -f etcd-certificates.yaml
$ kubectl delete issuer -n demo etcd-ca-issuer
$ kubectl delete secret -n demo etcd-ca
$ kubectl delete ns demo
```

Delete the `VaultServer` before the `Etcd` database. With `terminationPolicy: WipeOut`, the operator connects to etcd to remove the keys Vault stored there, so the database has to be reachable until the `VaultServer` is gone.

To see all the fields you can set under `spec.backend.etcd`, please check the [Etcd storage backend](/docs/concepts/vault-server-crds/storage/etcd.md) concept.

For more details on how to interact with the vault server, please check the [Vault Server guide](/docs/guides/vault-server/vault-server.md).
