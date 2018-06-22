# Deploying vault for etcd backend and unsealing it using kubernetes secret

## Deploying etcd cluster

Here, we will be using [coreos etcd operator](https://github.com/coreos/etcd-operator) to deploy etcd cluster.

### Deploy etcd operator

See [here](https://github.com/coreos/etcd-operator/blob/master/doc/user/install_guide.md) for instructions on how to install coreos etcd operator

### Deploy etcd cluster

We will deploy etcd cluster `my-etcd-cluster` on `default` namespace.

```yaml
apiVersion: "etcd.database.coreos.com/v1beta2"
kind: "EtcdCluster"
metadata:
  name: my-etcd-cluster
  namespace: default
spec:
  size: 3
  version: "3.2.13"
  TLS:
    static:
      member:
        peerSecret: etcd-peer-tls
        serverSecret: etcd-server-tls
      operatorSecret: etcd-client-tls
```

For deploying secure etcd cluster we need to generate some certificates. See [here](https://github.com/coreos/etcd-operator/blob/master/doc/user/cluster_tls.md) for information on etcd cluster TLS policy. 

Create etcd peer tls secret:
```console
$ tree .
.
├── peer-ca.crt
├── peer.crt
└── peer.key

$ kubectl create secret generic etcd-peer-tls --from-file=peer-ca.crt --from-file=peer.crt --from-file=peer.key
secret "etcd-peer-tls" created

```

Create etcd server tls secret:
```console
$ tree .
.
├── server-ca.crt
├── server.crt
└── server.key

$ kubectl create secret generic etcd-server-tls --from-file=server-ca.crt --from-file=server.crt --from-file=server.key
secret "etcd-server-tls" created

```

Create etcd client tls secret:
```console
$ tree .
.
├── etcd-client-ca.crt
├── etcd-client.crt
└── etcd-client.key

$ kubectl create secret generic etcd-client-tls --from-file=etcd-client-ca.crt --from-file=etcd-client.crt --from-file=etcd-client.key
secret "etcd-client-tls" created

```
> Note: In this example, all certificates issued by same CA.

Create etcd cluster:
```console
$ cat etcd_cluster.yaml 
apiVersion: "etcd.database.coreos.com/v1beta2"
kind: "EtcdCluster"
metadata:
  name: my-etcd-cluster
  namespace: default
spec:
  size: 3
  version: "3.2.13"
  TLS:
    static:
      member:
        peerSecret: etcd-peer-tls
        serverSecret: etcd-server-tls
      operatorSecret: etcd-client-tls

$ kubectl apply -f etcd_cluster.yaml
etcdcluster "my-etcd-cluster" created

$ kubectl get pods -n default
NAME                             READY     STATUS    RESTARTS   AGE
etcd-operator-79579db6bf-9qznh   1/1       Running   0          1h
my-etcd-cluster-bk5qqwqxbp       1/1       Running   0          5m
my-etcd-cluster-kw2vqd57kc       1/1       Running   0          4m
my-etcd-cluster-sx7rjx5ksn       1/1       Running   0          5m

$ kubectl get svc -n default
NAME                     TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
kubernetes               ClusterIP   10.96.0.1      <none>        443/TCP             1h
my-etcd-cluster          ClusterIP   None           <none>        2379/TCP,2380/TCP   19m
my-etcd-cluster-client   ClusterIP   10.99.253.86   <none>        2379/TCP            19m

```

Ectd can be accessed using `my-etcd-cluster-client` service.

## Deploy vault


### Deploy vault operator

See here.

### Deploy vault

We will deploy `my-vault` on `default` namespace. We will configure it for etcd storage backend which is already running on kubernetes cluster. We will use `kubernetes secret` for auto initializing and unsealing. 

```yaml
apiVersion: "core.kubevault.com/v1alpha1"
kind: "VaultServer"
metadata:
  name: "my-vault"
spec:
  nodes: 1
  version: "0.10.0"
  baseImage: "vault"
  backendStorage:
    etcd:
      address: "https://my-etcd-cluster-client.default.svc:2379"
      etcdApi: "v3"
      tlsSecretName: "vault-etcd-tls"
  unsealer:
    secretShares: 4
    secretThreshold: 2
    retryPeriodSeconds: 15
    insecureTLS: true
    mode:
      kubernetesSecret:
        secretName: vault-unseal-keys
```

Create `vault-etcd-tls` secret containing ca cert to verify etcd server, client cert and client key to use communication with etcd server.
```console
$ tree .
.
├── etcd-ca.crt
├── etcd-client.crt
└── etcd-client.key

$ kubectl create secret generic vault-etcd-tls --from-file=etcd-ca.crt --from-file=etcd-client.crt --from-file=etcd-client.key
secret "vault-etcd-tls" created

```
Create vault server:
```console
$ cat vault-crd.yaml
apiVersion: "core.kubevault.com/v1alpha1"
kind: "VaultServer"
metadata:
  name: "my-vault"
spec:
  nodes: 1
  version: "0.10.0"
  baseImage: "vault"
  backendStorage:
    etcd:
      address: "https://my-etcd-cluster-client.default.svc:2379"
      etcdApi: "v3"
      tlsSecretName: "vault-etcd-tls"
  unsealer:
    secretShares: 4
    secretThreshold: 2
    retryPeriodSeconds: 15
    insecureTLS: true
    mode:
      kubernetesSecret:
        secretName: vault-unseal-keys

$ kubectl apply -f vault-crd.yaml
vaultserver "my-vault" created

$ kubectl get vaultservers/my-vault -o yaml
apiVersion: core.kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: my-vault
  namespace: default
  ...
spec:
  backendStorage:
    etcd:
      address: https://my-etcd-cluster-client.default.svc:2379
      etcdApi: v3
      tlsSecretName: vault-etcd-tls
  baseImage: vault
  nodes: 1
  unsealer:
    insecureTLS: true
    mode:
      kubernetesSecret:
        secretName: vault-unseal-keys
    retryPeriodSeconds: 15
    secretShares: 4
    secretThreshold: 2
  version: 0.10.0
status:
  clientPort: 8200
  initialized: true
  phase: Running
  serviceName: my-vault
  updatedNodes:
  - my-vault-f99498d45-2fdmm
  vaultStatus:
    active: my-vault-f99498d45-2fdmm
    unsealed:
    - my-vault-f99498d45-2fdmm

```

Vault operator create an service with same name as vault server. In this example, vault can be accessed using `my-vault` service.

Check vault is unsealed:
```console
$ kubectl port-forward my-vault-f99498d45-2fdmm 8200:8200
Forwarding from 127.0.0.1:8200 -> 8200

# run following commands on another terminal
$ export VAULT_SKIP_VERIFY="true"

$ export VAULT_ADDR='http://127.0.0.1:8200'

$ vault status
Key             Value
---             -----
Seal Type       shamir
Sealed          false
Total Shares    4
Threshold       2
Version         0.10.0
Cluster Name    vault-cluster-3295f03f
Cluster ID      28200213-4da7-a906-b303-d52b933d8f14
HA Enabled      false
```

We can see vault unseal keys and root token from `vault-unseal-keys` secret
```console
$ kubectl get secrets/vault-unseal-keys -o yaml
apiVersion: v1
data:
  vault-root: ZGYxNjA2ZTgtYjljMy0wYTRiLTY2MTAtNGNhMDNmMTI4Mjhj
  vault-unseal-0: YWVjNGRhOTA1YWFiOWI2YjVlNDRmMGQ5MjJjZTM1ZDAwYTE2M2RmMDdmOWU5ZWI1NGIzNDJlMmYxNjQ3OWIxMjc1
  vault-unseal-1: NjAyOTcyZjFlZjg1NzRlZmYzNTBlZTM3ZWEyNjI5Njc2NDc4MTVhNGJkNzE2NzIzOTFhOTRjOTBhZDk1YjgwMzQz
  vault-unseal-2: MDgwNmRkZDkyMjA5YTYzMTRlMTg4MjE2OWUxYmI4NTc3YTY2ODU3ZWVkYTMwMGE5YjFmNDZhNzlhNTZjZmY1OWJm
  vault-unseal-3: MjZkMTZhYzczMTZjNzZlNDc5MmVhZjQ4Yjk0YzkyNDNmZmUzZTlhOGQxNzNlMTQzYTk4YmZkNzFhMzY4MDZlZmZl
kind: Secret
metadata:
  creationTimestamp: 2018-06-08T11:31:33Z
  name: vault-unseal-keys
  namespace: default
  resourceVersion: "43582"
  selfLink: /api/v1/namespaces/default/secrets/vault-unseal-keys
  uid: 7f4fd9ed-6b0f-11e8-b5d1-0800276cd133
type: Opaque

```
We can see the cofig that used when deploying vault. The config is stored in configMap named `{metadata.name}-vault-config`. For this example, it is `my-vault-vault-config`.
```console
$ kubectl get configMaps/my-vault-vault-config -o yaml
apiVersion: v1
data:
  vault.hcl: |2

    listener "tcp" {
      address = "0.0.0.0:8200"
      cluster_address = "0.0.0.0:8201"
      tls_cert_file = "/etc/vault/tls/server.crt"
      tls_key_file  = "/etc/vault/tls/server.key"
    }

    storage "etcd" {
    address = "https://my-etcd-cluster-client.default.svc:2379"
    etcd_api = "v3"
    ha_enable = "false"
    sync = "false"
    tls_ca_file = "/etc/vault/storage/etcd/tls/etcd-ca.crt"
    tls_cert_file = "/etc/vault/storage/etcd/tls/etcd-client.crt"
    tls_key_file = "/etc/vault/storage/etcd/tls/etcd-client.key"
    }
kind: ConfigMap
metadata:
  labels:
    app: vault
    vault_cluster: my-vault
  name: my-vault-vault-config
  namespace: default
  ...         
```