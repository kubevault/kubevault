---
title: Manage External Vault using KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: external-vault-platform
    name: External Vault
    parent: platform-guides
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage External Vault using KubeVault operator

The KubeVault operator can manage policies and secret engines of Vault servers which are not provisioned by the KubeVault operator. These Vault servers can be running outside a Kubernetes cluster or running inside a Kubernetes cluster but provisioned using a Helm chart.

The KubeVault operator can perform the following operations for externally provisioned Vault servers:

- Manage Vault [policy](https://www.vaultproject.io/docs/concepts/policies.html) using [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) and [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md). Guides can be found [here](/docs/guides/policy-management/overview.md).

In this tutorial, we are going to show how we can use KubeVault operator for Vault which is not provisioned by KubeVault operator.

## Connecting with Vault

We have a Vault running which can be accessible by the address `http://vault.default.svc:8200` from Kubernetes cluster. KubeVault operator use [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) to communicate with Vault. [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) provides a way of specifying Vault connection information and credential. Following authentication methods are currently supported by KubeVault operator using AppBinding:

- [Token Auth Method](https://www.vaultproject.io/docs/auth/token.html#token-auth-method)
- [Kubernetes Auth Method](https://www.vaultproject.io/docs/auth/kubernetes.html)
- [AWS IAM Auth Method](https://www.vaultproject.io/docs/auth/aws.html#iam-auth-method)
- [Userpass Auth Method](https://www.vaultproject.io/docs/auth/userpass.html)
- [TLS Certificates Auth Method](https://www.vaultproject.io/docs/auth/cert.html)

Vault authentication using AppBinding can be found in [here](/docs/concepts/vault-server-crds/auth-methods/overview.md).

In this tutorial, we are going to use [Kubernetes Auth Method](https://www.vaultproject.io/docs/auth/kubernetes.html).

Now, we are going to enable and configure [Kubenetes auth](https://www.vaultproject.io/docs/auth/kubernetes.html) in Vault.

- Create a service account and cluster role bindings that allow that service account to authenticate with the review token API.

  ```console
  $ cat docs/examples/guides/provider/external-vault/token-reviewer-sa.yaml
  apiVersion: v1
  kind: ServiceAccount
  metadata:
    name: token-reviewer
    namespace: demo

  $ cat docs/examples/guides/provider/external-vault/token-review-binding.yaml
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
    namespace: demo⏎

  $ kubectl apply -f docs/examples/guides/provider/external-vault/token-reviewer-sa.yaml
  serviceaccount/token-reviewer created

  $ kubectl apply -f docs/examples/guides/provider/external-vault/token-review-binding.yaml
  clusterrolebinding.rbac.authorization.k8s.io/role-tokenreview-binding created
  ```

- Enable Kubernetes auth in Vault.

  ```console
  $ vault auth enable kubernetes
  Success! Enabled Kubernetes auth method at: kubernetes/
  ```

- Configure Kubernetes auth in Vault.

  ```console
  $ kubectl get sa token-reviewer -n demo -o jsonpath="{.secrets[*]['name']}"
  token-reviewer-token-fvqsv

  $ export SA_JWT_TOKEN=$(kubectl get secret token-reviewer-token-fvqsv -n demo -o jsonpath="{.data.token}" | base64 --decode; echo)

  $ export SA_CA_CRT=$(kubectl get secret token-reviewer-token-fvqsv -n demo -o jsonpath="{.data['ca\.crt']}" | base64 --decode; echo)

  $ vault write auth/kubernetes/config \
      token_reviewer_jwt="$SA_JWT_TOKEN" \
      kubernetes_host="https://192.168.99.100:8443" \
      kubernetes_ca_cert="$SA_CA_CRT"
  Success! Data written to: auth/kubernetes/config
  ```

We are going to create a Vault [policy](https://www.vaultproject.io/docs/concepts/policies.html). It has permission to manage policy and Kubernetes role in Vault.

```console
$ cat docs/examples/guides/provider/external-vault/policy-admin.hcl
path "sys/policy/*" {
  capabilities = ["create", "update", "read", "delete", "list"]
}

path "sys/policy" {
  capabilities = ["read", "list"]
}

path "auth/kubernetes/role" {
  capabilities = ["read", "list"]
}

path "auth/kubernetes/role/*" {
  capabilities = ["create", "update", "read", "delete", "list"]
}

$ vault policy write policy-admin docs/examples/guides/provider/external-vault/policy-admin.hcl
Success! Uploaded policy: policy-admin
```

We are going to assign the above policy to a service account `policy-admin` so that we can use that service account to manage policy and Kubernetes role.

```console
$ cat docs/examples/guides/provider/external-vault/policy-admin-sa.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: policy-admin
  namespace: demo

$ kubectl apply -f docs/examples/guides/provider/external-vault/policy-admin-sa.yaml
serviceaccount/policy-admin created

$ vault write auth/kubernetes/role/policy-admin-role \
    bound_service_account_names=policy-admin \
    bound_service_account_namespaces=demo \
    policies=policy-admin \
    ttl=24h
Success! Data written to: auth/kubernetes/role/policy-admin-role
```

Now, we are going create AppBinding that will contain Vault information. For authentication, service account `policy-admin` and Kubernetes role `policy-admin-role` will be used.

```console
$ cat docs/examples/guides/provider/external-vault/vault-app.yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  clientConfig:
    url: http://vault.default.svc:8200
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9ERXlNamN3TkRVNU1qVmFGdzB5T0RFeU1qUXdORFU1TWpWYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMVhid2wyQ1NNc2VQTU5RRzhMd3dUVWVOCkI1T05oSTlDNzFtdUoyZEZjTTlUc1VDQnlRRk1weUc5dWFvV3J1ZDhtSWpwMVl3MmVIUW5udmoybXRmWGcrWFcKSThCYkJUaUFKMWxMMFE5MlV0a1BLczlXWEt6dTN0SjJUR1hRRDhhbHZhZ0JrR1ViOFJYaUNqK2pnc1p6TDRvQQpNRWszSU9jS0xnMm9ldFZNQ0hwNktpWTBnQkZiUWdJZ1A1TnFwbksrbU02ZTc1ZW5hWEdBK2V1d09FT0YwV0Z2CmxGQmgzSEY5QlBGdTJKbkZQUlpHVDJKajBRR1FNeUxodEY5Tk1pZTdkQnhiTWhRVitvUXp2d1EvaXk1Q2pndXQKeDc3d29HQ2JtM0o4cXRybUg2Tjl6Tlc3WlR0YTdLd05PTmFoSUFEMSsrQm5rc3JvYi9BYWRKT0tMN2dLYndJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFXeWFsdUt3Wk1COWtZOEU5WkdJcHJkZFQyZnFTd0lEOUQzVjN5anBlaDVCOUZHN1UKSS8wNmpuRVcyaWpESXNHNkFDZzJKOXdyaSttZ2VIa2Y2WFFNWjFwZHRWeDZLVWplWTVnZStzcGdCRTEyR2NPdwpxMUhJb0NrekVBMk5HOGRNRGM4dkQ5WHBQWGwxdW5veWN4Y0VMeFVRSC9PRlc4eHJxNU9vcXVYUkxMMnlKcXNGCmlvM2lJV3EvU09Yajc4MVp6MW5BV1JSNCtSYW1KWjlOcUNjb1Z3b3R6VzI1UWJKWWJ3QzJOSkNENEFwOUtXUjUKU2w2blk3NVMybEdSRENsQkNnN2VRdzcwU25seW5mb3RaTUpKdmFzbStrOWR3U0xtSDh2RDNMMGNGOW5SOENTSgpiTjBiZzczeVlWRHgyY3JRYk0zcko4dUJnY3BsWlRpUy91SXJ2QT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
  parameters:
    apiVersion: config.kubevault.com/v1alpha1
    kind: VaultServerConfiguration
    path: kubernetes
    vaultRole: policy-admin-role
    kubernetes:
      serviceAccountName: policy-admin

$ kubectl apply -f docs/examples/guides/provider/external-vault/vault-app.yaml
appbinding.appcatalog.appscode.com/vault-app created
```

If KubeVault operator uses the above AppBinding `vault-app`, then it will have the permission that is given to service account `policy-admin` by `policy-admin-role` role. Now, we are going to create [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) using `vault-app` AppBinding.

```console
$ cat docs/examples/guides/provider/external-vault/demo-policy.yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: demo-policy
  namespace: demo
spec:
  ref:
    name: vault-app
    namespace: demo
  policyDocument: |
    path "secret/*" {
      capabilities = ["create", "read", "update", "delete", "list"]
    }

$ kubectl apply -f  docs/examples/guides/provider/external-vault/demo-policy.yaml
vaultpolicy.policy.kubevault.com/demo-policy created

$ kubectl get vaultpolicies -n demo
NAME          STATUS    AGE
demo-policy   Success   3s

# To resolve the naming conflict, name of policy in Vault will follow this format: 'k8s.{clusterName}.{metadata.namespace}.{metadata.name}'. For this case, it is 'k8s.-.demo.demo-policy'.
$ vault policy list
default
k8s.-.demo.demo-policy
policy-admin
root

$ vault policy read k8s.-.demo.demo-policy
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```
