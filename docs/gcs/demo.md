# Deploying vault for gcs backend and unsealing it using Google Kms Gcs

## Provision GKE cluster
We will need a [IAM service account](https://cloud.google.com/iam/docs/service-accounts) with access to the gcs bucket and kms encryption key. This encryption key will be used for the purpose of encryption and decryption of unseal keys and root token. Assuming we have gcs bucket `vault-test-bucket` and crypto key `vault-init` under `tigerworks-kube` project.

Create service account `vault-sa`:
```console
gcloud iam service-accounts create vault-sa \
  --display-name "vault service account" \
  --project tigerworks-kube
```

Grant access to bucket:
```console
gsutil iam ch \
  serviceAccount:vault-sa@tigerworks-kube.iam.gserviceaccount.com:objectAdmin \
  gs://vault-test-bucket
```
```console
gsutil iam ch \
  serviceAccount:vault-sa@tigerworks-kube.iam.gserviceaccount.com:legacyBucketReader \
  gs://vault-test-bucket
```
Grant access to the crypto key: 
```console
gcloud kms keys add-iam-policy-binding \
  vault-init \
  --location global \
  --keyring vault-key-ring \
  --member serviceAccount:vault-sa@tigerworks-kube.iam.gserviceaccount.com \
  --role roles/cloudkms.cryptoKeyEncrypterDecrypter \
  --project tigerworks-kube
```
Create GKE cluster `vault`:
```console
gcloud container clusters create vault \
  --enable-autorepair \
  --cluster-version 1.9.6-gke.1 \
  --machine-type n1-standard-2 \
  --service-account vault-sa@tigerworks-kube.iam.gserviceaccount.com \
  --num-nodes 1 \
  --zone us-west1-c \
  --project tigerworks-kube
```

## Deploy vault

### Deploy vault operator

See here.

### Deploy vault

We will deploy `my-vault` on `default` namespace. We will configure it for gcs backend which is already created. We will use `google kms gcs` for auto initializing and unsealing. 

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
    gcs:
      bucket: "vault-test-bucket"
  unsealer:
    secretShares: 4
    secretThreshold: 2
    retryPeriodSeconds: 15
    insecureTLS: true
    mode:
      googleKmsGcs:
        bucket: "vault-test-bucket"
        kmsProject: "tigerworks-kube
    "
        kmsLocation: "global"
        kmsKeyRing: "vault-key-ring"
        kmsCryptoKey: "vault-init"
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
    gcs:
      bucket: "vault-test-bucket"
  unsealer:
    secretShares: 4
    secretThreshold: 2
    retryPeriodSeconds: 15
    insecureTLS: true
    mode:
      googleKmsGcs:
        bucket: "vault-test-bucket"
        kmsProject: "tigerworks-kube
    "
        kmsLocation: "global"
        kmsKeyRing: "vault-key-ring"
        kmsCryptoKey: "vault-init"

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
    gcs:
      bucket: vault-test-bucket
  baseImage: vault
  nodes: 1
  unsealer:
    insecureTLS: true
    mode:
      googleKmsGcs:
        Bucket: vault-test-bucket
        kmsCryptoKey: vault-init
        kmsKeyRing: vault-key-ring
        kmsLocation: global
        kmsProject: tigerworks-kube
    
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
  - my-vault-5568f99b5d-d9vs8
  vaultStatus:
    active: my-vault-5568f99b5d-d9vs8
    unsealed:
    - my-vault-5568f99b5d-d9vs8

```
Vault operator create an service with same name as vault server. In this example, vault can be accessed using `my-vault` service. 

Check vault is unsealed:
```console
$ kubectl port-forward my-vault-5568f99b5d-d9vs8 8200:8200
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
Version         0.10.0
Cluster Name    vault-cluster-366fcfb8
Cluster ID      ce076044-bc10-f358-fa3b-671449ae34ff
HA Enabled      false
```

We can see vault unseal keys and root token in `vault-test-bucket`:

![unseal-keys](bucket.png)

Download and decrypt the root token:
```console
$ export VAULT_TOKEN=$(gsutil cat gs://vault-test-bucket/vault-root | \
  gcloud kms decrypt \
    --project tigerworks-kube
 \
    --location global \
    --keyring vault-key-ring \
    --key vault-init \
    --ciphertext-file - \
    --plaintext-file - )

$ echo $VAULT_TOKEN
ec3a2374-38cc-14fe-1b61-2ab60f8763b0
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

    storage "gcs" {
    bucket = "vault-test-bucket"
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
