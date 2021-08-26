---
title: Mount PostgreSQL Secrets using CSI Driver
menu:
  docs_{{ .version }}:
    identifier: csi-driver-postgresql
    name: CSI Driver
    parent: postgresql-secret-engines
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Mount PostgreSQL Secrets using CSI Driver
## Kubernetes Secrets Store CSI Driver
![Secrets-store CSI architecture](/docs/guides/secret-engines/csi_architecture.svg)

When the `Pod` is created through the K8s API, it’s scheduled on to a node. The `kubelet` process on the node looks at the pod spec & see if there's any `volumeMount` request. The `kubelet` issues an `RPC` to the `CSI driver` to mount the volume. The `CSI driver` creates & mounts `tmpfs` into the pod. Then the `CSI driver` issues a request to the `Provider`. The provider talks to the external secrets store to fetch the secrets & write them to the pod volume as files. At this point, volume is successfully mounted & the pod starts running.

You can read more about the Kubernetes Secrets Store CSI Driver [here](https://secrets-store-csi-driver.sigs.k8s.io/).

## Consuming Secrets
At first, you need to have a Kubernetes 1.16 or later cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/). To check the version of your cluster, run:

```console
$ kubectl version --short
Client Version: v1.21.2
Server Version: v1.21.1
```

Before you begin:

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Install Secrets Store CSI driver for Kubernetes secrets in your cluster from [here](https://secrets-store-csi-driver.sigs.k8s.io/getting-started/installation.html).
- Install Vault Specific CSI provider from [here](https://github.com/hashicorp/vault-csi-provider)

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial stored in [examples](/docs/examples/guides/secret-engines/postgres) folder in GitHub repository [KubeVault/docs](https://github.com/kubevault/kubevault)

## Vault Server

If you don't have a Vault Server, you can deploy it by using the KubeVault operator.

- [Deploy Vault Server](/docs/guides/vault-server/vault-server.md)

The KubeVault operator can manage policies and secret engines of Vault servers which are not provisioned by the KubeVault operator. You need to configure both the Vault server and the cluster so that the KubeVault operator can communicate with your Vault server.

- [Configure cluster and Vault server](/docs/guides/vault-server/external-vault-sever.md#configuration)

Now, we have the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains connection and authentication information about the Vault server. And we also have the service account that the Vault server can authenticate.

```console
$ kubectl get appbinding -n demo
NAME    AGE
vault   50m

$ kubectl get appbinding -n demo vault -o yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  creationTimestamp: "2021-08-16T08:23:38Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: vault
    app.kubernetes.io/managed-by: kubevault.com
    app.kubernetes.io/name: vaultservers.kubevault.com
  name: vault
  namespace: demo
  ownerReferences:
  - apiVersion: kubevault.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: VaultServer
    name: vault
    uid: 6b405147-93da-41ff-aad3-29ae9f415d0a
  resourceVersion: "602898"
  uid: b54873fd-0f34-42f7-bdf3-4e667edb4659
spec:
  clientConfig:
    service:
      name: vault
      port: 8200
      scheme: http
  parameters:
    apiVersion: config.kubevault.com/v1alpha1
    kind: VaultServerConfiguration
    kubernetes:
      serviceAccountName: vault
      tokenReviewerServiceAccountName: vault-k8s-token-reviewer
      usePodServiceAccountForCSIDriver: true
    path: kubernetes
    vaultRole: vault-policy-controller
```

## Enable & Configure PostgreSQL SecretEngine

### Enable PostgreSQL SecretEngine
```console
$ kubectl apply -f docs/examples/guides/secret-engines/postgres/secretengine.yaml
secretengine.engine.kubevault.com/postgres-engine created
```

### Create PostgreSQLRole
```console
$ kubectl apply -f docs/examples/guides/secret-engines/postgres/secretenginerole.yaml
postgresrole.engine.kubevault.com/postgres-superuser-role created
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

```console
$ kubectl apply -f docs/examples/guides/secret-engines/postgres/serviceaccount.yaml
serviceaccount/test-user-account created

$ kubectl get serviceaccount -n demo
NAME                SECRETS   AGE
test-user-account   1         4h10m
```

### Create VaultPolicy and VaultPolicyBinding for Pod's Service Account
When a VaultPolicyBinding object is created, the KubeVault operator create an auth role in the Vault server. The role name is generated by the following naming format: `k8s.(clusterName or -).namespace.name`. Here, it is `k8s.-.demo.postgres-reader-role`.

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: postgres-reader-policy
  namespace: demo
spec:
  vaultRef:
    name: vault
  policyDocument: |
    path "your-database-path/creds/k8s.-.demo.postgres-superuser-role" {
      capabilities = ["read"]
    }
---
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: postgres-reader-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  policies:
    - ref: postgres-reader-policy
  subjectRef:
    kubernetes:
      serviceAccountNames:
        - "test-user-account"
      serviceAccountNamespaces:
        - "demo"
```

Let's create VaultPolicy and VaultPolicyBinding:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/postgres/policy.yaml
vaultpolicy.policy.kubevault.com/postgres-reader-policy created

$ kubectl apply -f docs/examples/guides/secret-engines/postgres/policybinding.yaml
vaultpolicybinding.policy.kubevault.com/postgres-reader-role created
```

Check if the VaultPolicy and the VaultPolicyBinding are successfully registered to the Vault server:

```console
$ kubectl get vaultpolicy -n demo
NAME                                    STATUS    AGE
postgres-reader-policy                  Success   8s

$ kubectl get vaultpolicybinding -n demo
NAME                                    STATUS    AGE
postgres-reader-role                    Success   10s
```

## Mount secrets into a Kubernetes pod

So, we can create `SecretProviderClass` now. You can read more about `SecretProviderClass` [here](https://secrets-store-csi-driver.sigs.k8s.io/concepts.html#secretproviderclass).

### Create SecretProviderClass

Create `SecretProviderClass` object with the following content:

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1alpha1
kind: SecretProviderClass
metadata:
  name: vault-db-provider
  namespace: demo
spec:
  provider: vault
  parameters:
    vaultAddress: "http://vault.demo:8200"
    roleName: "k8s.-.demo.postgres-reader-role"
    objects: |
      - objectName: "postgres-creds-username"
        secretPath: "your-database-path/creds/k8s.-.demo.postgres-superuser-role"
        secretKey: "username"
      - objectName: "postgres-creds-password"
        secretPath: "your-database-path/creds/k8s.-.demo.postgres-superuser-role"
        secretKey: "password"

```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/postgres/secretproviderclass.yaml
secretproviderclass.secrets-store.csi.x-k8s.io/vault-db-provider created
```
NOTE: The `SecretProviderClass` needs to be created in the same namespace as the pod.

### Create Pod

Now we can create a `Pod` to consume the `Postgres` secrets. When the `Pod` is created, the `Provider` fetches the secret and writes them to Pod's volume as files. At this point, the volume is successfully mounted and the `Pod` starts running.

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
          mountPath: "/secrets-store/postgres-creds"
          readOnly: true
  volumes:
    - name: secrets-store-inline
      csi:
        driver: secrets-store.csi.k8s.io
        readOnly: true
        volumeAttributes:
          secretProviderClass: "vault-db-provider"

```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/postgres/pod.yaml
pod/demo-app created
```
## Test & Verify

Check if the Pod is running successfully, by running:

```console
$ kubectl get pods -n demo
NAME                       READY   STATUS    RESTARTS   AGE
demo-app                   1/1     Running   0          11s
```

### Verify Secret

If the Pod is running successfully, then check inside the app container by running

```console
$ kubectl exec -it -n demo pod/demo-app -- /bin/sh
/ # ls /secrets-store/postgres-creds
postgres-creds-password  postgres-creds-username

/ # cat /secrets-store/postgres-creds/postgres-creds-password
TAu2Zvg1WYE07W8Uf-nW

/ # cat /secrets-store/postgres-creds/postgres-creds-username
v-kubernetes-test-k8s.-.demo.postgres-s-iPkxiH80Ollq2QgF82Ab-1629178048

/ # exit
```

So, we can see that the secret `db-username` and `db-password` is mounted into the pod, where the secret key is mounted as file and value is the content of that file.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted

```
