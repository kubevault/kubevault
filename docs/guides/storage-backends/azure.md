---
title: Vault Server Azure Blob Storage Backend
menu:
  docs_{{ .version }}:
    identifier: vault-server-azure-backend
    name: Azure Blob Storage
    parent: storage-backends-guides
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Deploy VaultServer with Azure Blob Storage Backend

You can easily deploy and manage [OpenBao](https://openbao.org/) in the Kubernetes cluster using KubeVault operator. In this tutorial, we are going to deploy Vault on the Kubernetes cluster using KubeVault operator, with [Azure Blob Storage](https://azure.microsoft.com/products/storage/blobs) as the storage backend.

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy VaultServer with Azure Blob Storage Backend

> **Create beforehand:** The Azure storage account. Vault creates the container in it if the container does not exist.

### Prepare the Azure Storage Account

Vault stores its data in an Azure Blob Storage container. There is no local emulator for this backend, so you need an Azure storage account (Blob Storage, hierarchical namespace disabled) and its access key. Vault creates the container if it does not exist yet.

You can find the key under *Security + networking* > *Access keys* of the storage account in the Azure portal. Use the **Key** value, not the connection string. You can also read it with the Azure CLI:

```bash
$ az storage account keys list \
    --account-name <storage-account-name> \
    --resource-group <resource-group> \
    --query '[0].value' -o tsv
```

**Create Credential Secret:**

The credential secret holds the account key in the key `account_key`.

```bash
$ kubectl create secret generic azure-cred -n demo \
    --from-literal=account_key=<storage-account-key>
secret/azure-cred created
```

## Deploy Vault using KubeVault

We're going to use Kubernetes secret to store the unseal-keys & root-token. A sample `VaultServer` with Azure backend manifest file may look like this:

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
    azure:
      accountName: "<storage-account-name>"
      container: "kubevault-storage-backend"
      credentialSecretRef:
        name: azure-cred
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```

Replace `<storage-account-name>` with the name of your storage account. The Azure storage backend does not support high availability, so the `VaultServer` runs with a single replica.

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
    backend: azure
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

At this stage, we've successfully deployed `Vault` using `KubeVault` operator with Azure as the storage backend.

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
Storage Type    azure
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
$ kubectl delete secret -n demo azure-cred
$ kubectl delete ns demo
```

With `terminationPolicy: WipeOut`, the operator removes the blobs Vault stored in the container when the `VaultServer` is deleted. The storage account and the container itself are not deleted.

To see all the fields you can set under `spec.backend.azure`, please check the [Azure storage backend](/docs/concepts/vault-server-crds/storage/azure.md) concept.

For more details on how to interact with the vault server, please check the [Vault Server guide](/docs/guides/vault-server/vault-server.md).
