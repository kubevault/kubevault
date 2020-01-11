---
title: Vault Policy Management
menu:
  docs_{{ .version }}:
    identifier: overview-policy-management
    name: Overview
    parent: policy-management-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Vault Policy Management

You can easily manage the Vault [policies](https://www.vaultproject.io/docs/concepts/policies.html) in Kubernetes native way using the KubeVault operator. The operator also provides functionality to create auth method roles that bind policies.

You should be familiar with the following CRD:

- [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md)
- [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md)
- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

## Before you begin

- Install KubeVault operator in your cluster following the steps [here](/docs/setup/operator/install.md).

- Deploy the Vault server or configure an existing one.
  - [Setup Vault Server](/docs/guides/vault-server/overview.md#setup-vault-server)

Now, you have the AppBinding that holds the connection information of the Vault server.

```console
$ kubectl get appbinding -n demo vault -o yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  labels:
    app: vault
    vault_cluster: vault
  name: vault
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9URXhNRGN3TXpVNE1qZGFGdzB5T1RFeE1EUXdNelU0TWpkYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdVFaUWZNU1pZek13djUzQjlKcmJlZkc0ClREYmtuaERGamZNMmJrdU90ak83cElFdG5ZSzRrVEdDRkRGd1RlTUNoczhiNFBQcGN5YzBZZ3BSdFFYMW9VTGUKdTFCOE0ralBtMXdhYys4S0JGR3BJdjVpS2dzMjI1MWczTThoY0lqK0ZFQ3hMVTN1bHZDazlUSXlJYzNLSGlDcwpFUmg2VXA2V1hxSVJYb3loNlViWmFrd2tsUTd5SGdUY0ZQMzNzNlBVVXVZaFNtUTJBNmxPU2NPSFRaVytHVDNrCjdPSzUxQ3g5RUptcjdZY1J2N3RiNXI0bGYycy81ZGg0SnYwS2UySkNCUExoK0dBeVh2cHlMd1dmc0w3cWduZ28KZSt0SXVadXFsSENwWDN1eERILzBncGRtK0NHTkN0aUlZUzdGOEFQODlsVVdrc2JOcVdKL1RnTlIrYmp6WHdJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFGb0QvMGtZSHk5OW96MU9xSEVOOC9UVjVXM0dIN1lZSUdqVHZ5Mmc2NFpsajNDbjQKdHdlejd3VGhtMGllazFSR1VXM3luNzg2cEswZVFqYWM3OWZpUG5iTEx0WkRVTzFkSVM5Rnp1MmFiNmpITEpxYQpteUswcHJ1NGNDbmtZTmtOdWhKclQwSkl4Rjc5cXBEanlhSjVuOUxXbDFQZmZwZDZrUW5vdjZrNGtQL2JoTXhECk9oWVBvaVcydkZxSEI4NCtPVHhLbHRkWnNwTkR3bEh1NjZ6a1c0Z0xiRVBnTWUzN1p3NUFISWZld2hqWG93VHYKNVZpZnZYUjVreVNZTTdCMHlqcnpzUzNvK1ZGdS9wMzdmWThwaHMwZzdCWjd3TlhCcWJBT010K2s0Vy9CU0xUagpOelB2STB0SGFCTklnRFdHbm1waEIxN1BpQnBPc1Y2RHdiOVRYdz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
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

## VaultPolicy

![Vault Policy](/docs/images/guides/policy-management/vault_policy.svg)

Using VaultPolicy, you can create, update and delete policy in Vault. In this tutorial, we are going to create `read-only-policy` in `demo` namespace with the permissions to read and list Vault policies.

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: read-only-policy
  namespace: demo
spec:
  vaultRef:
    name: vault
  policyDocument: |
    path "sys/policy" {
      capabilities = ["list"]
    }

    path "sys/policy/*" {
      capabilities = ["read"]
    }
```

Now, we are going to create VaultPolicy.

```console
$ kubectl apply -f docs/examples/guides/policy-management/read-only-policy.yaml
vaultpolicy.policy.kubevault.com/read-only-policy created
```

Check whether the VaultPolicy is successful.

```cosole
$ kubectl get vaultpolicy -n demo
NAME                           PHASE     AGE
read-only-policy               Success   15s
```

Check whether the policy is created in the Vault server. To resolve the naming conflict, name of policy in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`. In this case, it is `k8s.-.demo.read-only-policy`.

> Don't have Vault CLI? Enable Vault CLI from [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli).

```console
$ vault list sys/policy
Keys
----
k8s.-.demo.read-only-policy
... ...

$ vault read sys/policy/k8s.-.demo.read-only-policy
Key      Value
---      -----
name     k8s.-.demo.read-only-policy
rules    path "sys/policy" {
  capabilities = ["list"]
}

path "sys/policy/*" {
  capabilities = ["read"]
}
```

If we delete VaultPolicy `read-only-policy`, then the respective Vault policy will also be deleted from Vault.

```cosole
$ kubectl delete vaultpolicies/read-only-policy -n demo
vaultpolicy.policy.kubevault.com "read-only-policy" deleted
```

Check whether the policy is deleted in Vault.

```console
$ vault read sys/policy/k8s.-.demo.read-only-policy
No value found at sys/policy/k8s.-.demo.read-only-policy

$ vault list sys/policy
Keys
----
default
k8s.-.demo.vault-auth-method-controller
root
vault-policy-controller
```

## VaultPolicyBinding

![Vault Policy](/docs/images/guides/policy-management/vault_policy_binding.svg)

Using VaultPolicyBinding, you can create an auth method role that binds the Vault policies to users or service accounts.

Currently supported auth methods for creating role:

- [Kubernetes Auth Method](https://www.vaultproject.io/docs/auth/kubernetes.html)

In this tutorial, we are going to create a `policy-reader-role` VaultPolicyBinding in `demo` namespace.

Create a service account in the demo namespace:

```console
$ kubectl create serviceaccount -n demo  demo-sa
serviceaccount/demo-sa created
```

Get JWT token of the `demo-sa` service account:

```console
$ kubectl get secret -n demo demo-sa-token-jz7x5 -o jsonpath="{.data.token}" | base64 -d
eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZW1vIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlbW8tc2EtdG9rZW4tano3eDUiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVtby1zYSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjQxMGE4M2ViLWJhMDMtNDY1OS1iOTVjLTlmM2Y3ODdmZTM0OCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZW1vOmRlbW8tc2EifQ.w2H7cUXxAjeY4ZGJYVuTK8XrhpCXZeqUPQhFAyTndWhcevXOJFnK7jtyceYWaN0zy6TkBxHeAzVQdyLaFrNgecUKTzCZGaWHAoXlJOMY4Q49mHzEf3iGOBM7m1ckTTP9ABcOsVjD7OvlKslse_NnMDxVtuiughtMcrIhK5pbngQbJRpGkHaiOjgzIpHR3ybLmak7a24CXif0ZAqZd_y5l7bKi8eLr2Sidgq1R1sOMtOpnrj7qQCownw_KRrSPqhSCVSmaNDEYeqA9Jbw-JWVb3SW-FodjTPJsKj_qv791dZZE910CMBcsuJMPuAvNlX0cpOqO-7cdJzNG5y7IoYiSA
```

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: policy-reader-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  policies:
  - ref: read-only-policy
  subjectRef:
    kubernetes:
      serviceAccountNames:
        - "demo-sa"
      serviceAccountNamespaces:
        - "demo"
```

Here, a Kubernetes auth method role will be created that binds the `read-only-policy` policy to the service account `demo-sa` in `demo` namespace.

Let's create `read-only-policy` by using VaultPolicy:

```console
$ kubectl apply -f docs/examples/guides/policy-management/demo-policy.yaml
vaultpolicy.policy.kubevault.com/read-only-policy created
```

Check status:

```console
$ kubectl get vaultpolicy -n demo
NAME                           PHASE     AGE
read-only-policy               Success   15s
```

Now, we are going to create VaultPolicyBinding `policy-reader-role`.

```cosole
$ kubectl apply -f docs/examples/guides/policy-management/policy-reader-role.yaml
vaultpolicybinding.policy.kubevault.com/policy-reader-role created
```

Check whether the `policy-reader-role` is successful.

```console
$ kubectl get vaultpolicybinding -n demo
NAME                           PHASE     AGE
policy-reader-role             Success   11s
```

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/policy-management/../vault-server/vault-server.md#enable-vault-cli).

Check whether the Kubernetes auth role is created in Vault. To resolve the naming conflict,name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`. In this case, it is `k8s.-.demo.policy-reader-role`.

```console
$ vault list auth/kubernetes/role
Keys
----
k8s.-.demo.policy-reader-role
k8s.-.demo.vault-auth-method-controller
vault-policy-controller


$ vault read auth/kubernetes/role/k8s.-.demo.policy-reader-role
Key                                 Value
---                                 -----
bound_service_account_names         [demo-sa]
bound_service_account_namespaces    [demo]
policies                            [k8s.-.demo.read-only-policy]
token_bound_cidrs                   []
token_explicit_max_ttl              0s
token_max_ttl                       0s
token_no_default_policy             false
token_num_uses                      0
token_period                        0s
token_policies                      [k8s.-.demo.read-only-policy]
token_ttl                           0s
token_type                          default
```

Now, we are going to perform authentication to the Vault using `demo-sa`'s JWT token. In response to successful authentication, the Vault will provide us a token that will have permissions of the `read-only-policy` policy.

```console
$ vault write auth/kubernetes/login \
    role=k8s.-.demo.policy-reader-role \
    jwt=eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZW1vIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlbW8tc2EtdG9rZW4tano3eDUiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVtby1zYSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjQxMGE4M2ViLWJhMDMtNDY1OS1iOTVjLTlmM2Y3ODdmZTM0OCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZW1vOmRlbW8tc2EifQ.w2H7cUXxAjeY4ZGJYVuTK8XrhpCXZeqUPQhFAyTndWhcevXOJFnK7jtyceYWaN0zy6TkBxHeAzVQdyLaFrNgecUKTzCZGaWHAoXlJOMY4Q49mHzEf3iGOBM7m1ckTTP9ABcOsVjD7OvlKslse_NnMDxVtuiughtMcrIhK5pbngQbJRpGkHaiOjgzIpHR3ybLmak7a24CXif0ZAqZd_y5l7bKi8eLr2Sidgq1R1sOMtOpnrj7qQCownw_KRrSPqhSCVSmaNDEYeqA9Jbw-JWVb3SW-FodjTPJsKj_qv791dZZE910CMBcsuJMPuAvNlX0cpOqO-7cdJzNG5y7IoYiSA
Key                                       Value
---                                       -----
token                                     s.0of6p1q8SrcN3OscDaBlmWuI
token_accessor                            lvddBe41uXe0DK5xGoaPKsN7
token_duration                            768h
token_renewable                           true
token_policies                            ["default" "k8s.-.demo.read-only-policy"]
identity_policies                         []
policies                                  ["default" "k8s.-.demo.read-only-policy"]
token_meta_role                           k8s.-.demo.policy-reader-role
token_meta_service_account_name           demo-sa
token_meta_service_account_namespace      demo
token_meta_service_account_secret_name    demo-sa-token-jz7x5
token_meta_service_account_uid            410a83eb-ba03-4659-b95c-9f3f787fe348
```

Grab the token and export it as env  to check its behavior:

```console
$ export VAULT_TOKEN=s.0of6p1q8SrcN3OscDaBlmWuI

$ vault list sys/policy
Keys
----
default
k8s.-.demo.read-only-policy
k8s.-.demo.vault-auth-method-controller
root
vault-policy-controller

$ vault read sys/policy/k8s.-.demo.read-only-policy
Key      Value
---      -----
name     k8s.-.demo.read-only-policy
rules    path "sys/policy" {
  capabilities = ["list"]
}

path "sys/policy/*" {
  capabilities = ["read"]
}

$ vault delete sys/policy/k8s.-.demo.read-only-policy
Error deleting sys/policy/k8s.-.demo.read-only-policy: Error making API request.

URL: DELETE https://127.0.0.1:8200/v1/sys/policy/k8s.-.demo.read-only-policy
Code: 403. Errors:

* 1 error occurred:
  * permission denied

$ vault list auth/kubernetes/role
Error listing auth/kubernetes/role/: Error making API request.

URL: GET https://127.0.0.1:8200/v1/auth/kubernetes/role?list=true
Code: 403. Errors:

* 1 error occurred:
  * permission denied

```

Here, we can see that we don't have the permission to do anything but list and read the policies. So, the VaultPolicy `read-only-policy` and the VaultPolicyBinding `policy-reader-role` are working perfectly.

If we delete VaultPolicyBinding, then the respective role will be deleted from Vault.

```console
$ kubectl delete vaultpolicybinding policy-reader-role -n demo
vaultpolicybinding.policy.kubevault.com "policy-reader-role" deleted
```

Check whether the role is deleted from Vault.

```console
$ vault read auth/kubernetes/role/k8s.-.demo.policy-reader-role
No value found at auth/kubernetes/role/k8s.-.demo.policy-reader-role

$ vault list auth/kubernetes/role
Keys
----
k8s.-.demo.vault-auth-method-controller
vault-policy-controller
```
