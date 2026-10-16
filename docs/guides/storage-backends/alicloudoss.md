---
title: Vault Server Alicloud OSS Backend
menu:
  docs_{{ .version }}:
    identifier: vault-server-alicloudoss-backend
    name: Alicloud OSS
    parent: storage-backends-guides
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Deploy VaultServer with Alicloud OSS Backend

You can easily deploy and manage [OpenBao](https://openbao.org/) in the Kubernetes cluster using KubeVault operator. In this tutorial, we are going to deploy Vault on the Kubernetes cluster using KubeVault operator, with [Alibaba Cloud Object Storage Service (OSS)](https://www.alibabacloud.com/product/object-storage-service) as the storage backend.

![Vault Server](/docs/images/guides/vault-server/vault_server_guide.svg)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy VaultServer with Alicloud OSS Backend

### Prepare the OSS Bucket

Vault stores its data as objects in an Alibaba Cloud OSS bucket. There is no local emulator for this backend, so you need an Alibaba Cloud account with OSS activated. The bucket must exist before Vault starts. Vault does not create it.

Let's create a bucket and a RAM user that can use it in the Alibaba Cloud console,

- In the OSS console, create a private bucket with Standard storage in a region close to your cluster. Bucket names are unique across all of Alibaba Cloud, so choose your own. This tutorial uses the bucket `kubevault-storage-backend` in the Singapore region (`ap-southeast-1`).
- In RAM, create a user with permanent AccessKey access and grant it permission to list, read, write and delete the objects in the bucket, for example with the `AliyunOSSFullAccess` system policy or a custom policy limited to this bucket. Save the AccessKey ID and Secret, because the secret is shown only once.
- Note the public endpoint of the bucket's region. It has the form `https://oss-<region-id>.aliyuncs.com`, for example `https://oss-ap-southeast-1.aliyuncs.com` for Singapore, and it must be the endpoint of the region the bucket is in.

**Create Credential Secret:**

The credential secret holds the AccessKey pair in the keys `access_key` and `secret_key`.

```bash
$ kubectl create secret generic alicloud-cred -n demo \
    --from-literal=access_key=<AccessKey-ID> \
    --from-literal=secret_key=<AccessKey-Secret>
secret/alicloud-cred created
```

## Deploy Vault using KubeVault

We're going to use Kubernetes secret to store the unseal-keys & root-token. A sample `VaultServer` with Alicloud OSS backend manifest file may look like this:

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
    alicloudoss:
      endpoint: "https://oss-ap-southeast-1.aliyuncs.com"
      bucket: "kubevault-storage-backend"
      credentialSecretRef:
        name: alicloud-cred
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```

The Alicloud OSS storage backend does not support high availability, so the `VaultServer` runs with a single replica.

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
    backend: alicloudoss
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

At this stage, we've successfully deployed `Vault` using `KubeVault` operator with Alicloud OSS as the storage backend.

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
Version         2.6.1-sigilr.2
Storage Type    alicloudoss
Cluster Name    vault-cluster-12850424
Cluster ID      4e463f1c-3501-e520-41a7-be8979c4d2d4
HA Enabled      false
```

We can see the currently enabled list of secret engines.

```bash
$ vault secrets list
Path                                                              Type         Accessor              Description
----                                                              ----         --------              -----------
cubbyhole/                                                        cubbyhole    cubbyhole_614ec5cc    per-token private secret storage
identity/                                                         identity     identity_efb24c71     identity store
k8s.62e86b47-6e7c-4849-89fb-88062a22f451.kv.demo.vault-health/    kv           kv_72632299           n/a
sys/                                                              system       system_08abe22a       system endpoints used for control, policy and debugging
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
$ kubectl delete secret -n demo alicloud-cred
$ kubectl delete ns demo
```

With `terminationPolicy: WipeOut`, the operator removes the objects Vault stored in the bucket when the `VaultServer` is deleted. Delete the `VaultServer` before the credential secret, because the operator uses it to reach the bucket. The bucket itself and the RAM user's AccessKey are not deleted, so remove them in the Alibaba Cloud console when you no longer need them.

To see all the fields you can set under `spec.backend.alicloudoss`, please check the [Alicloud OSS storage backend](/docs/concepts/vault-server-crds/storage/alicloudoss.md) concept.

For more details on how to interact with the vault server, please check the [Vault Server guide](/docs/guides/vault-server/vault-server.md).
