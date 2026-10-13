---
title: Vault Server CockroachDB Backend
menu:
  docs_{{ .version }}:
    identifier: vault-server-cockroachdb-backend
    name: CockroachDB
    parent: storage-backends-guides
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Deploy VaultServer with CockroachDB Backend

You can easily deploy and manage [HashiCorp Vault](https://www.vaultproject.io/) in the Kubernetes cluster using KubeVault operator. In this tutorial, we are going to deploy Vault on the Kubernetes cluster using KubeVault operator, with [CockroachDB](https://www.cockroachlabs.com/) as the storage backend.

![Vault Server](/docs/images/guides/vault-server/vault_server_guide.svg)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy VaultServer with CockroachDB Backend

> This tutorial uses [cert-manager](https://cert-manager.io/docs/installation/) to issue the CockroachDB certificates. It must be installed in your cluster.

### Deploy Sample CockroachDB

CockroachDB speaks the PostgreSQL wire protocol. We are going to run a single-node CockroachDB secured with TLS, and let Vault log in as the `root` user with a client certificate.

**Create a CA and a cert-manager `Issuer`:**

KubeDB asks [cert-manager](https://cert-manager.io/docs/installation/) to issue the CockroachDB certificates. Let's create a self-signed CA and an `Issuer` backed by it. This CA is only for this tutorial, use your own CA or issuer in production.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout ca.key -out ca.crt -sha256 \
    -subj "/CN=cockroachdb-ca" \
    -addext 'basicConstraints=critical,CA:TRUE' \
    -addext 'keyUsage=critical,keyCertSign,cRLSign'

$ kubectl create secret tls cockroachdb-ca -n demo --cert=ca.crt --key=ca.key
secret/cockroachdb-ca created
```

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: cockroachdb-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: cockroachdb-ca
```

```bash
$ kubectl apply -f cockroachdb-ca-issuer.yaml
issuer.cert-manager.io/cockroachdb-ca-issuer created
```

**Create the Certificates:**

CockroachDB needs a node certificate. Vault needs a client certificate for the `root` user. Both are issued from the CA above:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: cockroachdb-node
  namespace: demo
spec:
  secretName: cockroachdb-node-tls
  commonName: node
  usages:
    - server auth
    - client auth
  dnsNames:
    - localhost
    - cockroachdb
    - cockroachdb.demo
    - cockroachdb.demo.svc
    - cockroachdb.demo.svc.cluster.local
  ipAddresses:
    - 127.0.0.1
  privateKey:
    algorithm: ECDSA
    rotationPolicy: Always
    size: 256
  issuerRef:
    name: cockroachdb-ca-issuer
    kind: Issuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: cockroachdb-client-root
  namespace: demo
spec:
  secretName: cockroachdb-client-tls
  commonName: root
  usages:
    - client auth
  privateKey:
    algorithm: ECDSA
    rotationPolicy: Always
    size: 256
  issuerRef:
    name: cockroachdb-ca-issuer
    kind: Issuer
    group: cert-manager.io
```

```bash
$ kubectl apply -f cockroachdb-certificates.yaml
certificate.cert-manager.io/cockroachdb-node created
certificate.cert-manager.io/cockroachdb-client-root created
```

**Create CockroachDB:**

Below is the YAML of the `Service` and `StatefulSet` that we are going to create for this tutorial. The sample does not use a persistent volume, so its data is gone when the pod is deleted.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: cockroachdb
  namespace: demo
  labels:
    app: cockroachdb
spec:
  ports:
  - port: 26257
    targetPort: 26257
    name: sql
  - port: 8080
    targetPort: 8080
    name: http
  selector:
    app: cockroachdb
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: cockroachdb
  namespace: demo
spec:
  serviceName: "cockroachdb"
  replicas: 1
  selector:
    matchLabels:
      app: cockroachdb
  template:
    metadata:
      labels:
        app: cockroachdb
    spec:
      containers:
      - name: cockroachdb
        image: cockroachdb/cockroach:v24.1.0
        imagePullPolicy: IfNotPresent
        command:
        - "/bin/bash"
        - "-ecx"
        - "exec /cockroach/cockroach start-single-node --certs-dir=/cockroach/cockroach-certs --http-port=8080 --listen-addr=0.0.0.0:26257"
        ports:
        - containerPort: 26257
          name: sql
        - containerPort: 8080
          name: http
        volumeMounts:
        - name: certs
          mountPath: /cockroach/cockroach-certs
          readOnly: true
        resources:
          limits:
            cpu: "1"
            memory: "1Gi"
          requests:
            cpu: "250m"
            memory: "512Mi"
      volumes:
      - name: certs
        secret:
          secretName: cockroachdb-node-tls
          # CockroachDB refuses private keys that are readable by group or others.
          defaultMode: 0400
          items:
          - key: ca.crt
            path: ca.crt
          - key: tls.crt
            path: node.crt
          - key: tls.key
            path: node.key
```

```bash
$ kubectl apply -f cockroachdb.yaml
service/cockroachdb created
statefulset.apps/cockroachdb created

$ kubectl wait -n demo --for=condition=Ready pod/cockroachdb-0 --timeout=5m
pod/cockroachdb-0 condition met
```

**Create Connection Secret:**

The `root` user logs in with its certificate, so the password stays empty. The operator adds the CA and the client certificate (`sslrootcert`, `sslcert` and `sslkey`) to the connection URL, so only `sslmode` is set here.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cockroachdb-cred
  namespace: demo
type: Opaque
stringData:
  username: "root"
  password: ""
  connection_url: "postgresql://root@cockroachdb.demo.svc:26257/defaultdb?sslmode=verify-full"
```

```bash
$ kubectl apply -f cockroachdb-cred.yaml
secret/cockroachdb-cred created
```

## Deploy Vault using KubeVault

We're going to use Kubernetes secret to store the unseal-keys & root-token. A sample `VaultServer` with CockroachDB backend manifest file may look like this:

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: sigilr-2.6.1.2
  replicas: 2
  allowedSecretEngines:
    namespaces:
      from: All
  backend:
    cockroachdb:
      address: "cockroachdb.demo.svc:26257"
      credentialSecretRef:
        name: cockroachdb-cred
      sslMode: verify-full
      # ca.crt, tls.crt and tls.key of the "root" client certificate
      tlsSecretRef:
        name: cockroachdb-client-tls
      haEnabled: true
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```

Vault creates the data table (`openbao_kv_store`) and, because `haEnabled` is `true`, the lock table (`openbao_ha_locks`) on its own. With `haEnabled`, the `VaultServer` can run with more than one replica: one pod is the active node and the others are standby.

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
    backend: cockroachdb
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
```

```bash
$ kubectl get vaultserver -n demo
NAME    REPLICAS   VERSION          STATUS   AGE
vault   2          sigilr-2.6.1.2   Ready    2m50s
```

At this stage, we've successfully deployed `Vault` using `KubeVault` operator with CockroachDB as the storage backend.

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
Storage Type    cockroachdb
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
$ kubectl delete secret -n demo cockroachdb-cred
$ kubectl delete -f cockroachdb.yaml
$ kubectl delete -f cockroachdb-certificates.yaml
$ kubectl delete issuer -n demo cockroachdb-ca-issuer
$ kubectl delete secret -n demo cockroachdb-ca
$ kubectl delete ns demo
```

Delete the `VaultServer` before CockroachDB. With `terminationPolicy: WipeOut`, the operator connects to CockroachDB to remove the data Vault stored there, so the database has to be reachable until the `VaultServer` is gone.

To see all the fields you can set under `spec.backend.cockroachdb`, please check the [CockroachDB storage backend](/docs/concepts/vault-server-crds/storage/cockroachdb.md) concept.

For more details on how to interact with the vault server, please check the [Vault Server guide](/docs/guides/vault-server/vault-server.md).
