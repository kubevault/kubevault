---
title: Vault Server OpenStack Swift Backend
menu:
  docs_{{ .version }}:
    identifier: vault-server-swift-backend
    name: OpenStack Swift
    parent: storage-backends-guides
    weight: 180
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Deploy VaultServer with OpenStack Swift Backend

You can easily deploy and manage [OpenBao](https://openbao.org/) in the Kubernetes cluster using KubeVault operator. In this tutorial, we are going to deploy Vault on the Kubernetes cluster using KubeVault operator, with [OpenStack Swift](https://docs.openstack.org/swift/latest/) as the storage backend.

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy VaultServer with OpenStack Swift Backend

> **Create beforehand:** The Swift container. Vault does not create it.

### Prepare the Swift Container

Use an existing OpenStack Swift container and the credentials of an OpenStack user that can use it. This example uses the container `vault-container`. The container must exist before Vault starts. Vault does not create it. You need the Keystone v3 authentication endpoint (`OS_AUTH_URL`), the project (`OS_PROJECT_NAME`), the region (`OS_REGION_NAME`) and the user and project domains of your account, which you can find in its OpenStack (`openrc`) credentials.

**Create Credential Secret:**

The credential secret holds the OpenStack login in the keys `username` and `password`. It must also contain the key `auth_token`. The token is only used for pre-authenticated access together with `storageURL`, so for a plain username and password login any placeholder value will do.

```bash
$ kubectl create secret generic swift-cred -n demo \
    --from-literal=username=<openstack-username> \
    --from-literal=password=<openstack-password> \
    --from-literal=auth_token=unused
secret/swift-cred created
```

## Deploy Vault using KubeVault

We're going to use Kubernetes secret to store the unseal-keys & root-token. A sample `VaultServer` with Swift backend manifest file may look like this:

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
    swift:
      authURL: "https://<keystone-host>/identity/v3"
      container: "vault-container"
      tenant: "<project-name>"
      region: "RegionOne"
      domain: "default"
      projectDomain: "default"
      credentialSecretRef:
        name: swift-cred
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```

Replace `authURL`, `tenant` (the OpenStack project), `region` and the domains with the values of your Swift account. `tenantID`, `trustID` and `storageURL` are also available, see the [Swift storage backend](/docs/concepts/vault-server-crds/storage/swift.md) concept. The Swift storage backend does not support high availability, so the `VaultServer` runs with a single replica.

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
    backend: swift
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

At this stage, we've successfully deployed `Vault` using `KubeVault` operator with Swift as the storage backend.

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
Storage Type    swift
Cluster Name    vault-cluster-482e469c
Cluster ID      af400967-737d-28ef-6d5f-4e0147674aa6
HA Enabled      false
```

We can see the currently enabled list of secret engines.

```bash
$ vault secrets list
Path                                                              Type         Accessor              Description
----                                                              ----         --------              -----------
cubbyhole/                                                        cubbyhole    cubbyhole_7f632162    per-token private secret storage
identity/                                                         identity     identity_0c7e4178     identity store
k8s.62e86b47-6e7c-4849-89fb-88062a22f451.kv.demo.vault-health/    kv           kv_9a8b45d0           n/a
sys/                                                              system       system_fd52bf25       system endpoints used for control, policy and debugging
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
$ kubectl delete secret -n demo swift-cred
$ kubectl delete ns demo
```

With `terminationPolicy: WipeOut`, the operator removes the objects Vault stored in the container when the `VaultServer` is deleted. Delete the `VaultServer` before the credential secret, because the operator uses it to reach the container. The container itself is not deleted, so remove it from your Swift account when you no longer need it.

To see all the fields you can set under `spec.backend.swift`, please check the [Swift storage backend](/docs/concepts/vault-server-crds/storage/swift.md) concept.

For more details on how to interact with the vault server, please check the [Vault Server guide](/docs/guides/vault-server/vault-server.md).
