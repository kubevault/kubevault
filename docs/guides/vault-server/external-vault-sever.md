---
title: External Vault Server
menu:
  docs_{{ .version }}:
    identifier: external-vault-server
    name: External Vault Server
    parent: vault-server-guides
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# External Vault Server

In this tutorial, we are going to demonstrate how the KubeVault operator works with external Vault servers (i.e. not provisioned by the KubeVault operator). To do so, we need to configure both the cluster and the Vault server.
Later we will create a [Vault policy](https://www.vaultproject.io/docs/concepts/policies.html)
using [VaultPolicy CRD](/docs/concepts/policy-crds/vaultpolicy.md) in Vault to check whether it is working or not.

## Before you begin

- Install KubeVault operator in your cluster following the steps [here](/docs/setup/operator/install.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

## Configuration

To communicate with Vault, the KubeVault operator needs to perform authentication to Vault sever.
The Vault server will issue a `token` in the response of successful authentication.
Then the KubeVault operator will perform the rest of the tasks using that `token`.
Hence, the `token` must have the path-permissions that we want to access from KubeVault operator over API call.

We will use [Kubernetes auth method](https://www.vaultproject.io/docs/auth/kubernetes.html) throughout the tutorial,
you can use any from the below list:

- [AWS IAM Auth Method](/docs/concepts/vault-server-crds/auth-methods/aws-iam.md)
- [Kubernetes Auth Method](/docs/concepts/vault-server-crds/auth-methods/kubernetes.md)
- [TLS Certificates Auth Method](/docs/concepts/vault-server-crds/auth-methods/tls.md)
- [Token Auth Method](/docs/concepts/vault-server-crds/auth-methods/token.md)
- [Userpass Auth Method](/docs/concepts/vault-server-crds/auth-methods/userpass.md)
- [GCP IAM Auth Method](/docs/concepts/vault-server-crds/auth-methods/gcp-iam.md)
- [Azure Auth Method](/docs/concepts/vault-server-crds/auth-methods/azure.md)

The whole configuration process can be divided into two parts:

- `Cluster configuration`: Create  an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) which holds
  connection and authentication information of Vault. Also create necessary Kubernetes resources (i.e. `secret`, `service account`, `ClusterRole`, `ClusterRoleBinding`, etc.) based on the requirements of the AppBinding.

- `Vault configuration`: Enable and Configure the auth method in Vault. Create [Vault policy](https://www.vaultproject.io/docs/concepts/policies.html) with necessary path-permissions which will be required by the KubeVault operator. Create a `user` or a `role` under the auth method mentioning the vault policies. This role name will be referenced by the AppBinding while performing authentication to Vault and the Vault will issue `token` in the response of successful authentication with assigned policies.

### Cluster Configuration

Since we are using the Kubernetes auth method, we need to create two Kubernetes `service accounts`.
One of them will be used by the Vault to verify Kubernetes authentication. The other one will be used by the AppBinding
to perform authentication to Vault.

#### Create Token Reviewer Service Account

The [Kubernetes auth method](https://www.vaultproject.io/docs/auth/kubernetes.html) can be used to authenticate with Vault using a Kubernetes Service Account Token. This auth method accesses the Kubernetes `TokenReview API` to validate the provided JWT is still valid. The service account used in this auth method will need to have access to the `TokenReview API`. If Kubernetes is configured to use RBAC roles, the Service Account should be granted permission to access this API.

Let's name token reviewer service account as `token-reviewer` and create it:

```console
$ kubectl create serviceaccount -n demo token-reviewer
serviceaccount/token-reviewer created
```

`ClusterRoleBinding` for the token reviewer service account:

```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: role-tokenreview-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- kind: ServiceAccount
  name: token-reviewer
  namespace: demo
```

Create `ClusterRoleBinding`:

```console
$ kubectl apply -f docs/examples/guides/vault-server/clusterRoleBinding.yaml
clusterrolebinding.rbac.authorization.k8s.io/role-tokenreview-binding created
```

Get the service account JWT token which will be used while configuring Vault:

```console
$ kubectl get secrets -n demo token-reviewer-token-s9hrs -o=jsonpath='{.data.token}' | base64 -d
eyJhbGciOiJSUzI1NiIsImtp...
```

#### Create AppBinding Service Account

The KubeVault operator will use the AppBinding that holds a reference to this service account to
perform authentication.  

Let's name the service account `vault` and create it:

```console
$ kubectl create serviceaccount -n demo vault
serviceaccount/vault created
```

#### Create AppBinding

The [AppBinding CRD](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) provides a way to specify connection information, credential and
parameters that are necessary for communicating with Vault.

Access vault server using `url`:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault
  namespace: demo
spec:
  clientConfig:
    url: https://demo-vault-server.com ## remote vault server url
    caBundle: eyJtc2ciOiJleGFtcGxlIn0= ## base64 encoded vault server ca.crt
  parameters:
    apiVersion: config.kubevault.com/v1alpha1
    kind: VaultServerConfiguration
    path: kubernetes ## Kubernetes auth is enabled in this path
    vaultRole: vault-role ## auth-role name against which login will be done
    kubernetes:
      serviceAccountName: vault  ## service account name
      usePodServiceAccountForCSIDriver: true ##  required while using CSI driver
```

Access vault server using Kubernetes service by replacing `spec.clientConfig`:

```yaml
spec:
  clientConfig:
    service:
      name: vault
      port: 8200
      scheme: HTTPS
    caBundle: eyJtc2ciOiJleGFtcGxlIn0= ## base64 encoded vault server ca.crt
```

Create AppBinding:

```console
$ kubectl apply -f docs/examples/guides/vault-server/appBinding.yaml
appbinding.appcatalog.appscode.com/vault created
```

### Vault Configuration

We will use Vault CLI to configure Vault.

1. Create [Vault policy](https://www.vaultproject.io/docs/concepts/policies.html) which contains a list of `path` along with
    `capacities`. For more details visit Vault [official doc](https://www.vaultproject.io/docs/concepts/policies.html#creating-policies).

    Create `vault.hcl` file:

    ```hcl
    path "sys/mounts" {
      capabilities = ["read", "list"]
    }
    path "sys/mounts/*" {
      capabilities = ["create", "read", "update", "delete"]
    }
    path "sys/leases/revoke/*" {
        capabilities = ["update"]
    }
    path "sys/policy/*" {
        capabilities = ["create", "update", "read", "delete", "list"]
    }
    path "sys/policy" {
        capabilities = ["read", "list"]
    }
    path "sys/policies" {
        capabilities = ["read", "list"]
    }
    path "sys/policies/*" {
        capabilities = ["create", "update", "read", "delete", "list"]
    }
    path "auth/kubernetes/role" {
        capabilities = ["read", "list"]
    }
    path "auth/kubernetes/role/*" {
        capabilities = ["create", "update", "read", "delete", "list"]
    }
    ```

    Create vault policy:

    ```console
    $ vault policy write vault-policy examples/guides/vault-server/vault.hcl
    Success! Uploaded policy: vault-policy
    ```

    List policies to check:

    ```console
    $ vault list sys/policy
    Keys
    ----
    default
    root
    vault-policy
    ```

2. Enable and configure the Kubernetes auth method (if not already enabled). For more details visit Vault
    [official doc](https://www.vaultproject.io/docs/auth/kubernetes.html#configuration).

    Enable Kubernetes auth:

    ```console
    $ vault auth enable kubernetes
    Success! Enabled kubernetes auth method at: kubernetes/
    ```

    Configure Kubernetes auth with `token-reviewer` service account JWT token:

    ```console
    $ vault write auth/kubernetes/config \
         token_reviewer_jwt="eyJhbGciOiJSUzI1N..." \
         kubernetes_host=https://127.0.0.1:40969\
         kubernetes_ca_cert=@examples/guides/vault-server/ca.crt
    Success! Data written to: auth/kubernetes/config
    ```

    You can find `kubernetes_host` and `kubernetes_ca_cert` in your cluster's `kubeconfig` file.

3. Create an auth method role which includes Vault policies.
   The KubeVault operator will perform authentication under this role and will have permission
   mentioned by the policies.

   Create Kubernetes auth method role:

   ```console
   $ vault write auth/kubernetes/role/vault-role \
           bound_service_account_names=vault \
           bound_service_account_namespaces=demo \
           policies=vault-policy \
           ttl=1h
    Success! Data written to: auth/kubernetes/role/vault-role
    ```

## Testing

We will create a Vault policy using [VaultPolicy CRD](/docs/concepts/policy-crds/vaultpolicy.md) to check whether our configuration worked or not.

Deploy `secret-policy.yaml`:

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: secret-admin
  namespace: demo
spec:
  vaultRef:
    name: vault
  vaultPolicyName: secret-admin
  policy:
    path:
      secret/*:
        capabilities:
        - create
        - read
        - update
        - delete
        - list
```

```console
$ kubectl apply -f docs/examples/guides/vault-server/secret-policy.yaml
vaultpolicy.policy.kubevault.com/secret-admin created
```

Now you can check from Vault:

```console
$ vault list sys/policy
Keys
----
default
root
secret-admin
vault-policy
```

So, we can see, `secret-admin` policy is already on the list.

Now delete the VaultPolicy crd:

```console
$ kubectl delete vaultpolicy secret-admin -n demo
vaultpolicy.policy.kubevault.com "secret-admin" deleted
````

Deleting VaultPolicy crd will also delete the policy from Vault.

Updated list:

```console
$ vault list sys/policy
Keys
----
default
root
vault-policy
```
