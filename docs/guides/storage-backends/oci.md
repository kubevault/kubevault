---
title: Vault Server OCI Object Storage Backend
menu:
  docs_{{ .version }}:
    identifier: vault-server-oci-backend
    name: OCI Object Storage
    parent: storage-backends-guides
    weight: 170
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Deploy VaultServer with OCI Object Storage Backend

You can easily deploy and manage [OpenBao](https://openbao.org/) in the Kubernetes cluster using KubeVault operator. In this tutorial, we are going to deploy Vault on the Kubernetes cluster using KubeVault operator, with [Oracle Cloud Infrastructure (OCI) Object Storage](https://www.oracle.com/cloud/storage/object-storage/) as the storage backend.

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy VaultServer with OCI Object Storage Backend

> **Create beforehand:** The data bucket and the lock bucket. Vault does not create them.

### Prepare the OCI Buckets

Use an existing OCI Object Storage setup. You need:

- the Object Storage namespace of your tenancy,
- a bucket for the Vault data and a second bucket for the HA lock objects. This example uses `vault-bucket` and `vault-lock-bucket`. Both must exist before Vault starts. Vault does not create them.
- an API signing key for a user that can read and write objects in both buckets.

Let's create the API signing key Vault uses, in the OCI console,

- Open the user that Vault should act as, go to *API keys* and click *Add API key*.
- Generate a key pair and download the private key.
- Click *Add*. The console shows a configuration snippet with the `user`, `fingerprint`, `tenancy` and `region` values. Keep it for the secret below.
- Give the user, through a group policy, permission to manage the objects in both buckets.

**Create Credential Secret:**

With `authTypeAPIKey: true`, Vault reads the standard OCI configuration. The credential secret holds this configuration in the key `config`. Fill in the values of your API signing key:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: oci-cred
  namespace: demo
type: Opaque
stringData:
  config: |
    [DEFAULT]
    user=ocid1.user.oc1..<unique-id>
    fingerprint=<api-key-fingerprint>
    tenancy=ocid1.tenancy.oc1..<unique-id>
    region=us-phoenix-1
    key_content=-----BEGIN RSA PRIVATE KEY-----\n<private-key-body>\n-----END RSA PRIVATE KEY-----\n
```

```bash
$ kubectl apply -f oci-cred.yaml
secret/oci-cred created
```

> If Vault runs on an OCI compute instance, it can authenticate with the instance principal instead. In that case leave out `authTypeAPIKey` and `credentialSecretRef`.

## Deploy Vault using KubeVault

We're going to use Kubernetes secret to store the unseal-keys & root-token. A sample `VaultServer` with OCI Object Storage backend manifest file may look like this:

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
    oci:
      bucketName: "vault-bucket"
      namespaceName: "<tenancy-namespace>"
      region: "us-phoenix-1"
      authTypeAPIKey: true
      credentialSecretRef:
        name: oci-cred
      haEnabled: true
      lockBucketName: "vault-lock-bucket"
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```

Replace `<tenancy-namespace>` with the Object Storage namespace of your tenancy. With `haEnabled: true`, Vault keeps its leader lock in `lockBucketName`, which is required in that case, so the `VaultServer` can run with more than one replica: one pod is the active node and the others are standby.

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
    backend: oci
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

At this stage, we've successfully deployed `Vault` using `KubeVault` operator with OCI Object Storage as the storage backend.

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
Storage Type    oci
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
$ kubectl delete secret -n demo oci-cred
$ kubectl delete ns demo
```

With `terminationPolicy: WipeOut`, the operator removes the objects Vault stored in the buckets when the `VaultServer` is deleted. The buckets themselves are not deleted.

To see all the fields you can set under `spec.backend.oci`, please check the [OCI Object Storage storage backend](/docs/concepts/vault-server-crds/storage/oci.md) concept.

For more details on how to interact with the vault server, please check the [Vault Server guide](/docs/guides/vault-server/vault-server.md).
