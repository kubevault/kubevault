---
title: Mount Key/Value Secrets using CSI Driver
menu:
  docs_{{ .version }}:
    identifier: csi-driver-kv
    name: CSI Driver
    parent: kv-secret-engines
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="KubeVault's built-in CSI driver has been removed in favor of [Secrets Store CSI driver for Kubernetes secrets](https://github.com/kubernetes-sigs/secrets-store-csi-driver)." >}}

# Mount Key/Value Secrets using CSI Driver

At first, you need to have a Kubernetes 1.14 or later cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/). To check the version of your cluster, run:

```console
$ kubectl version --short
Client Version: v1.16.2
Server Version: v1.14.0
```

Before you begin:

- Install KubeVault operator in your cluster from [here](/docs/setup/operator/install.md).
- Install Secrets Store CSI driver for Kubernetes secrets in your cluster from [here](https://secrets-store-csi-driver.sigs.k8s.io/getting-started/installation.html).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial stored in [examples](/docs/examples/guides/secret-engines/kv) folder in GitHub repository [KubeVault/docs](https://github.com/kubevault/kubevault)

## Vault Server

If you don't have a Vault Server, you can deploy it by using the KubeVault operator.

- [Deploy Vault Server](/docs/guides/vault-server/vault-server.md)

The KubeVault operator can manage policies and secret engines of Vault servers which are not provisioned by the KubeVault operator. You need to configure both the Vault server and the cluster so that the KubeVault operator can communicate with your Vault server.

- [Configure cluster and Vault server](/docs/guides/vault-server/external-vault-sever.md#configuration)

Now, we have the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains connection and authentication information about the Vault server. And we also have the service account that the Vault server can authenticate.

```console
$ kubectl get serviceaccounts -n demo
NAME                       SECRETS   AGE
vault                      1         20h

$ kubectl get appbinding -n demo
NAME    AGE
vault   50m

$ kubectl get appbinding -n demo vault -o yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9URXhNVEl3T1RFMU5EQmFGdzB5T1RFeE1Ea3dPVEUxTkRCYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdFZFZmtic2c2T085dnM2d1Z6bTlPQ1FYClBtYzBYTjlCWjNMbXZRTG0zdzZGaWF2aUlSS3VDVk1hN1NRSGo2L2YvOHZPeWhqNEpMcHhCM0hCYVFPZ3RrM2QKeEFDbHppU1lEd3dDbGEwSThxdklGVENLWndreXQzdHVQb0xybkppRFdTS2xJait6aFZDTHZ0enB4MDE3SEZadApmZEdhUUtlSXREUVdyNUV1QWlCMjhhSVF4WXREaVN6Y0h3OUdEMnkrblRMUEd4UXlxUlhua0d1UlIvR1B3R3lLClJ5cTQ5NmpFTmFjOE8wVERYRkIydWJQSFNza2xOU1VwSUN3S1IvR3BobnhGak1rWm4yRGJFZW9GWDE5UnhzUmcKSW94TFBhWDkrRVZxZU5jMlczN2MwQlhBSGwyMHVJUWQrVytIWDhnOVBVVXRVZW9uYnlHMDMvampvNERJRHdJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFabHRFN0M3a3ZCeTNzeldHY0J0SkpBTHZXY3ZFeUdxYUdCYmFUbGlVbWJHTW9QWXoKbnVqMUVrY1I1Qlg2YnkxZk15M0ZtZkJXL2E0NU9HcDU3U0RMWTVuc2w0S1RlUDdGZkFYZFBNZGxrV0lQZGpnNAptOVlyOUxnTThkOGVrWUJmN0paUkNzcEorYkpDU1A2a2p1V3l6MUtlYzBOdCtIU0psaTF3dXIrMWVyMUprRUdWClBQMzFoeTQ2RTJKeFlvbnRQc0d5akxlQ1NhTlk0UWdWK3ZneWJmSlFEMVYxbDZ4UlVlMzk2YkJ3aS94VGkzN0oKNWxTVklmb1kxcUlBaGJPbjBUWHp2YzBRRXBKUExaRDM2VDBZcEtJSVhjZUVGYXNxZzVWb1pINGx1Uk50SStBUAp0blg4S1JZU0xGOWlCNEJXd0N0aGFhZzZFZVFqYWpQNWlxZnZoUT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    service:
      name: vault
      port: 8200
      scheme: HTTPS
  parameters:
    apiVersion: config.kubevault.com/v1alpha1
    kind: VaultServerConfiguration
    path: kubernetes
    vaultRole: vault-policy-controller
    kubernetes:
      serviceAccountName: vault
      tokenReviewerServiceAccountName: vault-k8s-token-reviewer
      usePodServiceAccountForCSIDriver: true
```

## Enable and Configure KV Secret Engine

We will use the [Vault CLI](https://www.vaultproject.io/docs/commands/#vault-commands-cli-) throughout the tutorial to [enable and configure](https://www.vaultproject.io/docs/secrets/kv/kv-v1.html#setup) the KV secret engine.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

### Enable KV Secret Engine

Enable the KV secret engine:

```console
$ vault secrets enable -version=1 kv
Success! Enabled the kv secrets engine at: kv/
```

### Write KV Secret

Write arbitrary key-value pairs:

```console
$ vault kv put kv/my-secret my-value=s3cr3t
Success! Data written to: kv/my-secret
```

### Read KV Secret

Read a specific key-value pair:

```console
$ vault kv get kv/my-secret
====== Data ======
Key         Value
---         -----
my-value    s3cr3t
```

## Update Vault Policy

Since Pod's service account will be used by the CSI driver to perform [Kubernetes authentication](https://www.vaultproject.io/docs/auth/kubernetes.html) to the Vault server, the auth method role must have the permission to read secret at `kv/*` path.

During the Vault server configuration, we have created a Kubernetes service account and registered an auth method role at the Vault server. If you have used the KubeVault operator to deploy the Vault server, then the operator has performed these tasks for you.

So, we have the service account that will be referenced from the Pod.

```console
kubectl get serviceaccounts -n demo
NAME                       SECRETS   AGE
vault                      1         7h23m
```

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

You can find the name of the auth method role in the AppBinding's `spec.parameters.vaultRole`. Let's list the token policies assigned for `vault` service account:

```console
$ vault read auth/kubernetes/role/vault-policy-controller
Key                                 Value
---                                 -----
bound_service_account_names         [vault]
bound_service_account_namespaces    [demo]
token_bound_cidrs                   []
token_explicit_max_ttl              0s
token_max_ttl                       24h
token_no_default_policy             false
token_num_uses                      0
token_period                        24h
token_policies                      [default vault-policy-controller]
token_ttl                           24h
token_type                          default
```

Now, we will update the Vault policy `vault-policy-controller` and add the permission to read at `kv/*` path with existing permissions.

`kv-readonly-policy.hcl:`

```yaml
path "kv/*" {
    capabilities = ["read"]
}
```

Update the `vault-policy-controller` policy:

```console
# write existing polices to a file
$ vault policy read vault-policy-controller > examples/guides/secret-engines/kv/policy.hcl

# append the kv-readonly-policy at the end of the existing policies
$ cat examples/guides/secret-engines/kv/kv-readonly-policy.hcl >> examples/guides/secret-engines/kv/policy.hcl

# write the update policy to Vault
$ vault policy write vault-policy-controller examples/guides/secret-engines/kv/policy.hcl
Success! Uploaded policy: vault-policy-controller

# read updated policy
$ vault policy read vault-policy-controller
... ...
... ...
path "kv/*" {
    capabilities = ["read"]
}
```

So, we have updated the policy successfully and ready to mount the secrets into Kubernetes pods.

## Mount secrets into a Kubernetes pod

Since Kubernetes 1.14, `storage.k8s.io/v1beta1` `CSINode` and `CSIDriver` objects were introduced. Let's check [CSIDriver](https://kubernetes-csi.github.io/docs/csi-driver-object.html) and [CSINode](https://kubernetes-csi.github.io/docs/csi-node-object.html) are available or not.

```console
$ kubectl get csidrivers
NAME                        CREATED AT
secrets.csi.kubevault.com   2019-12-09T04:32:50Z

$ kubectl get csinodes
NAME             CREATED AT
2gb-pool-57jj7   2019-12-09T04:32:52Z
2gb-pool-jrvtj   2019-12-09T04:32:58Z
```

So, we can create `StorageClass` now.

### Create StorageClass

Create `StorageClass` object with the following content:

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: vault-kv-storage
  annotations:
    storageclass.kubernetes.io/is-default-class: "false"
provisioner: secrets.csi.kubevault.com
parameters:
  ref: demo/vault # namespace/AppBinding, we created during vault server configuration
  engine: KV # vault engine name
  secret: my-secret # secret name on vault which you want get access
  path: kv # specify the secret engine path, default is kv
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/kv/storageClass.yaml
storageclass.storage.k8s.io/vault-kv-storage created
```

## Test & Verify

Let's create a separate namespace called `trial` for testing purpose.

```console
$ kubectl create ns trial
namespace/trail created
```

### Create PVC

Create a `PersistentVolumeClaim` with the following data. This makes sure a volume will be created and provisioned on your behalf.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: csi-pvc-kv
  namespace: trial
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Mi
  storageClassName: vault-kv-storage
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/kv/pvc.yaml
persistentvolumeclaim/csi-pvc-kv created
```

### Create VaultPolicy and VaultPolicyBinding for Pod's Service Account

Let's say pod's service account name is `pod-sa` located in `trial` namespace. We need to create a [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) and a [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md) so that the pod has access to read secrets from the Vault server.

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: kv-se-policy
  namespace: demo
spec:
  vaultRef:
    name: vault
  # Here, kv secret engine is enabled at "kv".
  # If the path was "demo-se", policy should be like
  # path "demo-se/*" {}.
  policyDocument: |
    path "kv/*" {
      capabilities = ["create", "read"]
    }
---
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: kv-se-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  policies:
  - ref: kv-se-policy
  subjectRef:
    kubernetes:
      serviceAccountNames:
      - "pod-sa"
      serviceAccountNamespaces:
      - "trial"
```

Let's create VaultPolicy and VaultPolicyBinding:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/kv/vaultPolicy.yaml
vaultpolicy.policy.kubevault.com/kv-se-policy created

$ kubectl apply -f docs/examples/guides/secret-engines/kv/vaultPolicyBinding.yaml
vaultpolicybinding.policy.kubevault.com/kv-se-role created
```

Check if the VaultPolicy and the VaultPolicyBinding are successfully registered to the Vault server:

```console
$ kubectl get vaultpolicy -n demo
NAME                           STATUS    AGE
kv-se-policy                  Success   8s

$ kubectl get vaultpolicybinding -n demo
NAME                           STATUS    AGE
kv-se-role                    Success   10s
```

### Create Service Account for Pod

Let's create the service account `pod-sa` which was used in VaultPolicyBinding. When a VaultPolicyBinding object is created, the KubeVault operator create an auth role in the Vault server. The role name is generated by the following naming format: `k8s.(clusterName or -).namespace.name`. Here, it is `k8s.-.demo.kv-se-role`. We need to provide the auth role name as service account `annotations` while creating the service account. If the annotation `secrets.csi.kubevault.com/vault-role` is not provided, the CSI driver will not be able to perform authentication to the Vault.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pod-sa
  namespace: trial
  annotations:
    secrets.csi.kubevault.com/vault-role: k8s.-.demo.kv-se-role
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/kv/podServiceAccount.yaml
serviceaccount/pod-sa created
```

### Create Pod

Now we can create a Pod which refers to this volume. When the Pod is created, the volume will be attached, formatted and mounted to the specific container.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod
  namespace: trial
spec:
  containers:
    - name: mypod
      image: busybox
      command:
      - sleep
      - "3600"
      volumeMounts:
      - name: my-vault-volume
        mountPath: "/etc/kv"
        readOnly: true
  serviceAccountName: pod-sa # service account that was created
  volumes:
  - name: my-vault-volume
    persistentVolumeClaim:
      claimName: csi-pvc-kv
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/kv/pod.yaml
pod/mypod created
```

Check if the Pod is running successfully, by running:

```console
$ kubectl get pods -n trial
NAME                    READY   STATUS    RESTARTS   AGE
mypod                   1/1     Running   0          11s
```

### Verify Secret

If the Pod is running successfully, then check inside the app container by running

```console
$ kubectl exec -it -n trial  mypod sh
/ # ls /etc/kv/
my-value

/ # cat /etc/kv/my-value
s3cr3t

/ # exit
```

So, we can see that the secret `my-secret` is mounted into the pod, where the secret key is mounted as file and value is the content of that file.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted

$ kubectl delete ns trial
namespace "trial" deleted
```
