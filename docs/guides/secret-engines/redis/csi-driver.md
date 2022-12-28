---
title: Mount Redis Secrets using CSI Driver
menu:
    docs_{{ .version }}:
        identifier: csi-driver-redis
        name: CSI Driver
        parent: redis-secret-engines
        weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Mount Redis Secrets using CSI Driver

## Kubernetes Secrets Store CSI Driver

Secrets Store CSI driver for Kubernetes secrets - Integrates secrets stores with Kubernetes via a [Container Storage Interface (CSI)](https://kubernetes-csi.github.io/docs/) volume.

The Secrets Store CSI driver `secrets-store.csi.k8s.io` allows Kubernetes to mount multiple secrets, keys, and certs stored in enterprise-grade external secrets stores into their pods as a volume. Once the Volume is attached, the data in it is mounted into the container’s file system.

![Secrets-store CSI architecture](/docs/guides/secret-engines/csi_architecture.svg)

When the `Pod` is created through the K8s API, it’s scheduled on to a node. The `kubelet` process on the node looks at the pod spec & see if there's any `volumeMount` request. The `kubelet` issues an `RPC` to the `CSI driver` to mount the volume. The `CSI driver` creates & mounts `tmpfs` into the pod. Then the `CSI driver` issues a request to the `Provider`. The provider talks to the external secrets store to fetch the secrets & write them to the pod volume as files. At this point, volume is successfully mounted & the pod starts running.

You can read more about the Kubernetes Secrets Store CSI Driver [here](https://secrets-store-csi-driver.sigs.k8s.io/).

## Consuming Secrets

At first, you need to have a Kubernetes 1.16 or later cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/). To check the version of your cluster, run:

```bash
$ kubectl version --short
Client Version: v1.24.0
Kustomize Version: v4.5.4
Server Version: v1.23.13
```

Before you begin:

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Install Secrets Store CSI driver for Kubernetes secrets in your cluster from [here](https://secrets-store-csi-driver.sigs.k8s.io/getting-started/installation.html).
- Install Vault Specific CSI provider from [here](https://github.com/hashicorp/vault-csi-provider)

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial stored in [examples](/docs/examples/guides/secret-engines/redis) folder in GitHub repository [KubeVault/docs](https://github.com/kubevault/kubevault)

## Vault Server

If you don't have a Vault Server, you can deploy it by using the KubeVault operator. To create a Redis Secret Engine, VaultServer version needs to be 1.12.1+.

- [Deploy Vault Server](/docs/guides/vault-server/vault-server.md)

The KubeVault operator can manage policies and secret engines of Vault servers which are not provisioned by the KubeVault operator. You need to configure both the Vault server and the cluster so that the KubeVault operator can communicate with your Vault server.

- [Configure cluster and Vault server](/docs/guides/vault-server/external-vault-sever.md#configuration)

Now, we have the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains connection and authentication information about the Vault server. And we also have the service account that the Vault server can authenticate.

```bash
$ kubectl get appbinding -n demo
NAME    AGE
vault   50m

$ kubectl get appbinding -n demo vault -o yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  creationTimestamp: "2022-12-27T09:37:31Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: vault
    app.kubernetes.io/managed-by: kubevault.com
    app.kubernetes.io/name: vaultservers.kubevault.com
  name: vault
  namespace: demo
  ownerReferences:
  - apiVersion: kubevault.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: VaultServer
    name: vault
    uid: e32d10cd-aec9-4060-bc9a-098d69bb5d6b
  resourceVersion: "294415"
  uid: 09011421-5a2f-44cf-a8ac-7069565b0f78
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
    backend: raft
    backupTokenSecretRef:
      name: vault-backup-token
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
            value: k8s.kubevault.com.demo.vault
        restoreTask:
          name: vault-restore-1.10.3
          params:
          - name: keyPrefix
            value: k8s.kubevault.com.demo.vault
    unsealer:
      mode:
        kubernetesSecret:
          secretName: vault-keys
      secretShares: 5
      secretThreshold: 3
    vaultRole: vault-policy-controller
```

## Enable & Configure REdis SecretEngine

### Enable Redis SecretEngine

```bash
$ kubectl apply -f docs/examples/guides/secret-engines/redis/secretengine.yaml
secretengine.engine.kubevault.com/redis-secret-engine created
```

### Create RedisRole

```bash
$ kubectl apply -f docs/examples/guides/secret-engines/redis/secretenginerole.yaml
redisrole.engine.kubevault.com/write-read-role created
```

Let's say pod's service account name is `test-user-account` located in `demo` namespace. We need to create a [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) and a [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md) so that the pod has access to read secrets from the Vault server.

### Create Service Account for Pod

Let's create the service account `test-user-account` which will be used in VaultPolicyBinding.
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-user-account
  namespace: demo
```

```bash
$ kubectl apply -f docs/examples/guides/secret-engines/redis/serviceaccount.yaml
serviceaccount/test-user-account created

$ kubectl get serviceaccount -n demo
NAME                SECRETS   AGE
test-user-account   1         4h10m
```

### Create SecretRoleBinding for Pod's Service Account

SecretRoleBinding will create VaultPolicy and VaultPolicyBinding inside vault.
When a VaultPolicyBinding object is created, the KubeVault operator create an auth role in the Vault server. The role name is generated by the following naming format: `k8s.(clusterName or -).namespace.name`. Here, it is `k8s.kubevault.com.demo.write-read-role`.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretRoleBinding
metadata:
  name: secret-role-binding
  namespace: demo
spec:
  roles:
    - kind: RedisRole
      name: write-read-role
  subjects:
    - kind: ServiceAccount
      name: test-user-account
      namespace: demo
```

Let's create SecretRoleBinding:

```bash
$ kubectl apply -f docs/examples/guides/secret-engines/redis/secret-role-binding.yaml
secretrolebinding.engine.kubevault.com/secret-role-binding created
```
Check if the VaultPolicy and the VaultPolicyBinding are successfully registered to the Vault server:

```bash
$ kubectl get vaultpolicy -n demo
NAME                                 STATUS    AGE
srb-demo-secret-role-binding                 Success   8s

$ kubectl get vaultpolicybinding -n demo
NAME                                 STATUS    AGE
srb-demo-secret-role-binding                    Success   10s
```

## Mount secrets into a Kubernetes pod

So, we can create `SecretProviderClass` now. You can read more about `SecretProviderClass` [here](https://secrets-store-csi-driver.sigs.k8s.io/concepts.html#secretproviderclass).

### Create SecretProviderClass

Get `roleName` from VaultPolicyBinding
```bash
$ kubectl get vaultpolicybinding -n demo srb-demo-secret-role-binding  -o=jsonpath="{['spec.vaultRoleName']}"
k8s.kubevault.com.demo.srb-demo-secret-role-binding
```

The `secretPath` can be constructed as `your-data-base-path/creds/your-role-name`. 
Or get secretPath from VaultPolicy
```bash
$ kubectl get vaultpolicy -n demo srb-demo-secret-role-binding -o=jsonpath="{['spec.policyDocument']}"
path "/k8s.kubevault.com.redis.demo.redis-secret-engine/creds/k8s.kubevault.com.demo.write-read-role" {
  capabilities = ["read"]
}
```

The secretPath here is `/k8s.kubevault.com.redis.demo.redis-secret-engine/creds/k8s.kubevault.com.demo.write-read-role`

Create `SecretProviderClass` object with the following content:

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: vault-db-provider
  namespace: demo
spec:
  provider: vault
  parameters:
    vaultAddress: "http://vault.demo:8200"
    roleName: k8s.kubevault.com.demo.srb-demo-secret-role-binding
    objects: |
      - objectName: "redis-creds-username"
        secretPath: "/k8s.kubevault.com.redis.demo.redis-secret-engine/creds/k8s.kubevault.com.demo.write-read-role"
        secretKey: "username"
      - objectName: "redis-creds-password"
        secretPath: "/k8s.kubevault.com.redis.demo.redis-secret-engine/creds/k8s.kubevault.com.demo.write-read-role"
        secretKey: "password"
```

```bash
$ kubectl apply -f docs/examples/guides/secret-engines/redis/secretproviderclass.yaml
secretproviderclass.secrets-store.csi.x-k8s.io/vault-db-provider created
```

NOTE: The `SecretProviderClass` needs to be created in the same namespace as the pod.

### Create Pod

Now we can create a `Pod` to consume the `Redis` secrets. When the `Pod` is created, the `Provider` fetches the secret and writes them to Pod's volume as files. At this point, the volume is successfully mounted and the `Pod` starts running.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-app
  namespace: demo
spec:
  serviceAccountName: test-user-account
  containers:
    - image: jweissig/app:0.0.1
      name: demo-app
      imagePullPolicy: Always
      volumeMounts:
        - name: secrets-store-inline
          mountPath: "/secrets-store/redis-creds"
          readOnly: true
  volumes:
    - name: secrets-store-inline
      csi:
        driver: secrets-store.csi.k8s.io
        readOnly: true
        volumeAttributes:
          secretProviderClass: "vault-db-provider"
```

```bash
$ kubectl apply -f docs/examples/guides/secret-engines/redis/pod.yaml
pod/demo-app created
```
## Test & Verify

Check if the Pod is running successfully, by running:

```bash
$ kubectl get pods -n demo
NAME                       READY   STATUS    RESTARTS   AGE
demo-app                   1/1     Running   0          11s
```

### Verify Secret

If the Pod is running successfully, then check inside the app container by running

```bash
$ kubectl exec -it -n demo pod/demo-app -- /bin/sh
/ # ls /secrets-store/redis-creds
redis-creds-password  redis-creds-username

/ # cat /secrets-store/redis-creds/redis-creds-password
eEip9Orr-yFONjlRGntY

/ # cat /secrets-store/redis-creds/redis-creds-username
V_KUBERNETES-DEMO-TEST-USER-ACCOUNT_K8S.KUBEVAULT.COM.DEMO.WRITE-READ-ROLE_764DXABBDPMUGZP9C6AB_1672/app

/ # exit
```

So, we can see that the secret `db-username` and `db-password` is mounted into the pod, where the secret key is mounted as file and value is the content of that file.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete ns demo
namespace "demo" deleted

```
