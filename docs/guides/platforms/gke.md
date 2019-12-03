---
title: Deploy Vault on Google Kubernetes Engine (GKE)
menu:
  docs_{{ .version }}:
    identifier: gke-platform
    name: GKE
    parent: platform-guides
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Deploy Vault on Google Kubernetes Engine (GKE)

Here, we are going to deploy Vault in GKE using Vault operator. We are going to use [GCS bucket](https://cloud.google.com/storage/docs/) as Vault backend and `googleKmsGcs` unsealer mode for automatic unsealing the Vault.

## Before You Begin

At first, you need to have a GKE cluster. If you don't already have a cluster, create one from [here](https://cloud.google.com/kubernetes-engine/).

- Install Vault operator in your cluster following the steps [here](/docs/setup/operator/install.md).

- You should be familiar with the following CRD:
  - [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md)
  - [Unsealer](/docs/concepts/vault-server-crds/unsealer/unsealer.md)
  - [googleKmsGcs](/docs/concepts/vault-server-crds/unsealer/google_kms_gcs.md)

- You will need a [GCS bucket](https://cloud.google.com/storage/docs/) to use it as Vault backend storage. In this tutorial, we are going to use `demo-vault` GCS bucket.

- You will need a [Google KMS](https://cloud.google.com/kms/) crypto key to use it for Vault unsealer. In this tutorial, we are going to use key `vault-key` int `vault` key ring.


### Provision Cluster

We are going to use [gcloud](https://cloud.google.com/sdk/gcloud/) to provision a cluster.

```console
$ gcloud container clusters create vault \
      --enable-autorepair \
      --cluster-version 1.11.4-gke.13 \
      --machine-type n1-standard-2 \
      --num-nodes 1 \
      --zone us-east1-b \
      --project ackube
```

![gke instance](/docs/images/guides/provider/gke/gke-cluster.png)

Now, we are going to create service account and set access permission to this service account.

```console
$ gcloud iam service-accounts create vault-sa \
      --display-name "vault service account" \
      --project ackube
Created service account [vault-sa].
```

Grant access to bucket:

```console
$ gsutil iam ch \
      serviceAccount:vault-sa@ackube.iam.gserviceaccount.com:objectAdmin \
      gs://demo-vault
```

```console
$ gsutil iam ch \
      serviceAccount:vault-sa@ackube.iam.gserviceaccount.com:legacyBucketReader \
      gs://demo-vault
```

Grant access to the crypto key:

```console
$ gcloud kms keys add-iam-policy-binding \
      vault-key \
      --location global \
      --keyring vault \
      --member serviceAccount:vault-sa@ackube.iam.gserviceaccount.com \
      --role roles/cloudkms.cryptoKeyEncrypterDecrypter \
      --project ackube
```

### Install Vault operator

See [here](/docs/setup/operator/install.md).

```console
$ kubectl get pods -n kube-system
NAME                                              READY   STATUS    RESTARTS   AGE
vault-operator-7cc8cdf7f6-jmhg4                   1/1     Running   6          8m
```

### Deploy Vault

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

We will deploy `my-vault` on `demo` namespace. We will configure it for GCS backend. We will use `googleKmsGcs` for auto initializing and unsealing. We already created a GCS bucket `demo-vault`.

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: my-vault
  namespace: demo
spec:
  replicas: 1
  version: "1.0.0"
  backend:
    gcs:
      bucket: "demo-vault"
      credentialSecret: "google-cred"
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      googleKmsGcs:
        bucket: "demo-vault"
        kmsProject: "ackube"
        kmsLocation: "global"
        kmsKeyRing: "vault"
        kmsCryptoKey: "vault-key"
        credentialSecret: "google-cred"
```

Here, `spec.version` specifies the name of the [VaultServerVersion](docs/concepts/vault-server-crds/vaultserverversion.md) CRD. If that does not exist, then create one.

```console
$ kubectl get vaultserverversions
NAME     VERSION   VAULT_IMAGE    DEPRECATED   AGE
1.0.0    1.0.0     vault:1.0.0    false        1m

$ kubectl get vaultserverversions/1.0.0 -o yaml
apiVersion: catalog.kubevault.com/v1alpha1
kind: VaultServerVersion
metadata:
  name: 1.0.0
spec:
  deprecated: false
  exporter:
    image: kubevault/vault-exporter:0.1.0
  unsealer:
    image: kubevault/vault-unsealer:0.2.0
  vault:
    image: vault:1.0.0
  version: 1.0.0
```

`spec.backend.gcs.credentialSecret` and `spec.unsealer.mode.googleKmsGcs.credentialSecret` specifies the name of the Kubernetes secret containing `vault-sa@ackube.iam.gserviceaccount.com` credential.

```console
$ kubectl get secrets/google-cred -n demo -o yaml
apiVersion: v1
data:
  sa.json: ewogICJ0eXBlIjogIn...
kind: Secret
metadata:
  name: google-cred
  namespace: demo
type: Opaque

```

Now, we are going to create `my-vault` in `demo` namespace.

```console
$ cat examples/guides/provider/gke/my-vault.yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: my-vault
  namespace: demo
spec:
  replicas: 1
  version: "1.0.0"
  backend:
    gcs:
      bucket: "demo-vault"
      credentialSecret: "google-cred"
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      googleKmsGcs:
        bucket: "demo-vault"
        kmsProject: "ackube"
        kmsLocation: "global"
        kmsKeyRing: "vault"
        kmsCryptoKey: "vault-key"
        credentialSecret: "google-cred"

$ kubectl apply -f examples/guides/provider/gke/my-vault.yaml
vaultserver.kubevault.com/my-vault created
```

Check the `my-vault` status. It may take some time to reach `Running` stage.

```console
$ kubectl get vaultserver/my-vault -n demo
NAME       NODES   VERSION   STATUS    AGE
my-vault   1       1.0.0     Running   2m
```

`status` field in `my-vault` will show more detail information.

```console
$ kubectl get vaultserver/my-vault -n demo -o json | jq '.status'
{
  "clientPort": 8200,
  "initialized": true,
  "observedGeneration": "1$6208915667192219204",
  "phase": "Running",
  "serviceName": "my-vault",
  "updatedNodes": [
    "my-vault-75b6f87dbb-kq4tp"
  ],
  "vaultStatus": {
    "active": "my-vault-75b6f87dbb-kq4tp",
    "unsealed": [
      "my-vault-75b6f87dbb-kq4tp"
    ]
  }
}

```

Vault operator will create a service `{metadata.name}` for `my-vault` in the same namespace. For this case, service name is `my-vault`. You can specify service configuration in [spec.serviceTemplate](/docs/concepts/vault-server-crds/vaultserver.md#specservicetemplate). Vault operator will use that configuration to create service.

```console
$ kubectl get services -n demo
NAME       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
my-vault   ClusterIP   10.3.244.122   <none>        8200/TCP,8201/TCP,9102/TCP   4m
```

The configuration used to run Vault can be found in `{metadata.name}-vault-config` configMap. For this case, it is `my-vault-vault-config`. Confidential data are omitted in this configMap.

```console
$ kubectl get configmaps -n demo
NAME                    DATA      AGE
my-vault-vault-config   1         49m

$ kubectl get configmaps/my-vault-vault-config -n demo -o yaml
apiVersion: v1
data:
  vault.hcl: |2-

    listener "tcp" {
      address = "0.0.0.0:8200"
      cluster_address = "0.0.0.0:8201"
      tls_cert_file = "/etc/vault/tls/server.crt"
      tls_key_file  = "/etc/vault/tls/server.key"
    }

    storage "gcs" {
      bucket = "demo-vault"
    }

    telemetry {
      statsd_address = "0.0.0.0:9125"
    }
kind: ConfigMap
metadata:
  name: my-vault-vault-config
  namespace: demo
```

In this `my-vault`, Vault operator will use self-signed certificates for Vault and also will create `{metadata.name}-vault-tls` secret containing certificates. You can optionally specify certificates in [spec.tls](/docs/concepts/vault-server-crds/vaultserver.md#spectls).

```console
$ kubectl get secrets -n demo
NAME                                      TYPE                                  DATA      AGE
my-vault-vault-tls                        Opaque                                3         1h
```

We can see unseal keys and root token in `demo-vault` bucket.

![unseal keys](/docs/images/guides/provider/gke/gke-unseal-keys.png)

### Using Vault

Download and decrypt the root token:

```console
$ export VAULT_TOKEN=$(gsutil cat gs://demo-vault/vault-root-token | \
  gcloud kms decrypt \
    --project ackube \
    --location global \
    --keyring vault \
    --key vault-key \
    --ciphertext-file - \
    --plaintext-file - )

$ echo $VAULT_TOKEN
s.5DEELd1OiRmwfnrqfqQeguug
```

> Note: Make sure you have the permission to do above operation. Also we highly recommend not to use root token for using vault.

For testing purpose, we are going to port forward the active vault pod, since the service we exposed for Vault is ClusterIP type. Make sure Vault cli is installed.

```console
$ kubectl port-forward my-vault-75b6f87dbb-kq4tp -n demo 8200:8200
Forwarding from 127.0.0.1:8200 -> 8200

# run following commands on another terminal
$ export VAULT_SKIP_VERIFY="true"

$ export VAULT_ADDR='https://127.0.0.1:8200'

$ vault status
Key             Value
---             -----
Seal Type       shamir
Sealed          false
Total Shares    4
Threshold       2
Version         1.0.0
Cluster Name    vault-cluster-84d6b1b0
Cluster ID      bb6487bb-0deb-9e95-144e-e85c9ebd07eb
HA Enabled      false

```

Set Vault token for further use. In this case, we are going to use root token(not recommended).

```console
$  export VAULT_TOKEN='s.5DEELd1OiRmwfnrqfqQeguug'

$ vault secrets list
Path          Type         Accessor              Description
----          ----         --------              -----------
cubbyhole/    cubbyhole    cubbyhole_9ce16bb9    per-token private secret storage
identity/     identity     identity_45904875     identity store
secret/       kv           kv_22970276           key/value secret storage
sys/          system       system_51cd4d05       system endpoints used for control, policy and debugging

```

We are going to write,read and delete a secret in Vault

```console
$ vault kv put secret/foo A=B
Success! Data written to: secret/foo

# see written secret data
$ vault kv get secret/foo
== Data ==
Key    Value
---    -----
A      B

# delete the secret
$ vault kv delete secret/foo
Success! Data deleted (if it existed) at: secret/foo

# check the secret whether it is exist or not
$ vault kv get secret/foo
No value found at secret/foo

```
