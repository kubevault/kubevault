# Vault Policy Management

You can easily manage Vault [Policy](https://www.vaultproject.io/docs/concepts/policies.html) in Kubernetes way using Vault operator.

You should be familiar with the following CRD:

- [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md)
- [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md)
- [AppBinding](/docs/concepts/appbinding-crds/appbinding.md)

Before you begin:

- Install Vault operator in your cluster following the steps [here](/docs/setup/operator/install.md).

- Deploy Vault. It could be in the Kubernetes cluster or external.


To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```


## VaultPolicy

Using VaultPolicy, you can create, update and delete policy in Vault. In this tutorial, we are going to create `demo-policy` in `demo` namespace. 

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: demo-policy
  namespace: demo
spec:
  vaultAppRef:
    name: vault
    namespace: demo
  policy: |
    path "secret/*" {
      capabilities = ["create", "read", "update", "delete", "list"]
    }
```

We already have deployed Vault using Vault operator. For this tutorial, we deployed Vault with [inmem](inmem.md) backend storage and unsealed with [kubernetes secret](kubernetes_secret.md) mode.

```console
$ kubectl get vaultserver -n demo
NAME      NODES     VERSION   STATUS    AGE
vault     1         1.0.0     Running   46s
```

```console
$ kubectl get vaultserver/vault -n demo -o yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  backend:
    inmem: {}
  nodes: 1
  unsealer:
    mode:
      kubernetesSecret:
        secretName: vault-keys
    secretShares: 4
    secretThreshold: 2
  version: 1.0.0
status:
  clientPort: 8200
  initialized: true
  observedGeneration: 1$6208915667192219204
  phase: Running
  serviceName: vault
  updatedNodes:
  - vault-69cc8bfb45-jggnz
  vaultStatus:
    active: vault-69cc8bfb45-jggnz
    unsealed:
    - vault-69cc8bfb45-jggnz
```

When Vault is deployed by Vault operator using [VaultServer](docs/concepts/vault-server-crds/vaultserver.md). It does following things:

- It enables and configures the kubernetes auth in Vault.

- It also create a policy in Vault that has policy create, update and delete permission. Also create a role in Vault that binds a serviceaccount with policy. This serviceaccount name and role name are specified in AppBinding that is also created by Vault operator.

For VaultServer, AppBinding is created in same namespace of the VaultServer CRD and also name is the same as VaultServer CRD's name.

```cosole
$ kubectl get appbindings/vault -n demo -o yaml
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
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9ERXlNall3TmpBeU1qaGFGdzB5T0RFeU1qTXdOakF5TWpoYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBd3kxaFk0eTlrWDNPc0M1eHJJTXlzdzFyCkFMKzY0S1YwSXhNM0ltYlh3b3B2eEFjQVNXdi9PWXJTZ2VPaE9sUnM2aUVLVUNGZjA5blRvVjdlenR0bVVtSzgKUXBCYldxVFd2NHlhbGxpMTFSMUc1VXdIQzlIRUdjeUc4QkRtTTBWL1R2YW94aHkvUkFTZGJFRXltVVZJdSt6Zwp3TVovSWdPQ3ZyRFNiZ291OExFdkZjYTVvQmxic2YzbWNRRUpSVnFnaUx4bjV0MG1VcVROZ2Nzc0pVcWtnL0o4CnpZVlhaTTN5d1NWZ1NwcnIzMFExa1FPYjNzalNSd3diVnlHRG5LMTN0WExkSEJla3JVYjZub0ZPQWJDVWtKUm0KK25ZREQ4TmZ3U1RwRVlzem1CR0dzYkU0UVJoMFVlNGd3a01UMGU4QkxOeVVIelJ5U21QQjh1OHk1V1RPRFFJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFQSjVUK2phaXhTZ3cxS2tSYjdybys5U3VmbHJKZVhaNzZNMktkck9nMTBXUlFFajEKRG5xakdRSStVY0NYbUpEMHRvSXFlNmgvUmtkblVYZlA3WElUVHdlZnR0T0toM0VTbTZGbk9XMFJPSno5eDRTbQpQVFBiLzl5NjIvNE9OZ1RNU1V0MG1RWVBOUm15a0VkZ2tNOTZqRmgwZkRtUGc4dUxkZG5RNlZsaVBmZitUcnQ0CldDNWVhVmdFcWVQSkdGcFR0SHRSblNNR3pRNDRMWjlyYlN4Wk1RV1krYnJJTm9tRHZ6cW1DWjA5Rit0ZFpCVWcKOWRBUllwR28zV2h3enEyNisvS0NSZnNMY01zZ2htZmVxNFlzUW5VelhBL240QVV2bG40MDF2TjhYQmkxbUFtSApoNm1oWlh1YWVlT0tCVTllc2dlMk8wSmJkTW5HZkZjVWlwbmRNUT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    service:
      name: vault
      port: 8200
      scheme: HTTPS
  parameters:
    apiVersion: config.kubevault.com/v1alpha1
    kind: VaultServerConfiguration
    authPath: kubernetes
    serviceAccountName: vault
    policyControllerRole: vault-policy-controller
    authMethodControllerRole: k8s.-.demo.vault-auth-method-controller
    tokenReviewerServiceAccountName: vault-k8s-token-reviewer
    usePodServiceAccountForCsiDriver: true
```

Here, `vault-policy-controller` role binds the serviceaccount `vault` with the policy. So, this serviceaccount `vault` has policy create, update and delete permission in Vault.

In this tutorial, we are going to use this AppBinding to authenticate against the Vault. See [here](/docs/concepts/appbinding-crds/vault-authentication-using-appbinding.md) for Vault authentication using AppBinding in Vault operator.

Now, we are going to create VaultPolicy.

```console
$ cat examples/guides/policy-management/demo-policy.yaml 
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: demo-policy
  namespace: demo
spec:
  vaultAppRef:
    name: vault
    namespace: demo
  policy: |
    path "secret/*" {
      capabilities = ["create", "read", "update", "delete", "list"]
    }
    
$ kubectl apply -f examples/guides/policy-management/demo-policy.yaml 
vaultpolicy.policy.kubevault.com/demo-policy created
``` 

Check whether the VaultPolicy is successful.

```cosole
$ kubectl get vaultpolicies -n demo
NAME                           STATUS    AGE
demo-policy                      Success   1m
```

Check whether the policy is created in Vault. To resolve the naming conflict, name of policy in Vault will follow this format: `k8s.{spec.clusterName or -}.{spec.namespace}.{spec.name}`. For this case, it is `k8s.-.demo.demo-policy`.

```console
$ vault policy list
k8s.-.demo.demo-policy

$ vault policy read k8s.-.demo.demo-policy
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

If we delete VaultPolicy `demo-policy`, then respective policy will also be deleted from Vault.

```cosole
$ kubectl delete vaultpolicies/demo-policy -n demo
vaultpolicy.policy.kubevault.com "demo-policy" deleted
```

Check whether the policy is deleted in Vault.

```console
$ vault policy read k8s.-.demo.demo-policy
No policy named: k8s.-.demo.demo-policy

$ vault policy list
default
k8s.-.demo.vault-auth-method-controller
vault-policy-controller
root
```

## VaultPolicyBinding

Using VaultPolicyBinding, you can bind serviceaccount with the policy created using VaultPolicy. Vault operator will create Vault kuberenetes [role](https://www.vaultproject.io/api/auth/kubernetes/index.html#create-role) when VaultPolicyBinding is created. In this tutorial, we are going to create `demo-role` VaultPolicyBinding in `demo` namespace.

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: demo-role
  namespace: demo
spec:
  roleName: "demo-role"
  policies : ["demo-policy"] # name of the VaultPolicies
  serviceAccountNames: ["demo-sa"]
  serviceAccountNamespaces: ["demo"]
  TTL: "1000"
  maxTTL: "2000"
  Period: "1000"
```

Here, `demo-sa` in `demo` namespace will have the permission that is specified in `demo-policy` VaultPolicy.

Now, we are going to create VaultPolicyBinding `demo-role`.

```cosole
$ cat examples/guides/policy-management/demo-role.yaml 
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: demo-role
  namespace: demo
spec:
  roleName: "demo-role"
  policies : ["demo-policy"]
  serviceAccountNames: ["demo-sa"]
  serviceAccountNamespaces: ["demo"]
  TTL: "1000"
  maxTTL: "2000"
  Period: "1000"

$ kubectl apply -f examples/guides/policy-management/demo-role.yaml 
vaultpolicybinding.policy.kubevault.com/demo-role created
```

Check whether the `demo-role` is successful.

```console
$ kubectl get vaultpolicybindings -n demo
NAME                           STATUS    AGE
demo-role                      Success   2m
```

Check whether role is created in Vault. 

```console
$ vault list auth/kubernetes/role
Keys
----
demo-role

$ vault read auth/kubernetes/role/demo-role
Key                                 Value
---                                 -----
bound_cidrs                         []
bound_service_account_names         [demo-sa]
bound_service_account_namespaces    [demo]
max_ttl                             33m20s
num_uses                            0
period                              0s
policies                            [k8s.-.demo.demo-policy]
ttl                                 16m40s

```

Now, we are going to login using serviceaccount `demo-sa` and create secret data in Vault.

```console
# get the serviceaccount token secret name
$ kubectl get serviceaccount/demo-sa -n demo -o json | jq '.secrets'
[
  {
    "name": "demo-sa-token-ddk7c"
  }
]

# get serviceaccount token
$ kubectl get secrets/demo-sa-token-ddk7c -n demo -o json | jq -r '.data.token' | base64 -d
eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZW1vIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlbW8tc2EtdG9rZW4tZGRrN2MiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVtby1zYSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjQ2NDQ4YzhhLTA4ZjYtMTFlOS1iMWYyLTA4MDAyN2I3MTg5ZCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZW1vOmRlbW8tc2EifQ.a604tfmrqtE0Dsaslenqc8LSBIA8CLGXY7xvwORdIGV5dXZHCvyDM-qZTUso_tQasyHduFZd5c0kQFFg6jJHfp3jfVOReGFFzeqOV3TWq-eL79Up8YxZ4Jt8INiTbaGWXF8sU7_kM7aJ_PcXuZ508q62QXoL6_GWIy1I2xjdYfgb3tHKLVZ10GiyMcve6Go8tPbeY2emaSBfVWZGxcJrMuy6qCuSFowfjbpzQJkfKxqXMZN1SNyUuXVkZnpWIqUBDAqBY_NCKRDKt_iwxIDsklXTl8ANsGIc_8FZXVJUKOX5pccs4KARu5_gLxqq14fKvAXV7bwsbSM-Xe03obWyMA

# login with serviceaccount token
$ vault write auth/kubernetes/login role=demo-role jwt="eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZW1vIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlbW8tc2EtdG9rZW4tZGRrN2MiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVtby1zYSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjQ2NDQ4YzhhLTA4ZjYtMTFlOS1iMWYyLTA4MDAyN2I3MTg5ZCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZW1vOmRlbW8tc2EifQ.a604tfmrqtE0Dsaslenqc8LSBIA8CLGXY7xvwORdIGV5dXZHCvyDM-qZTUso_tQasyHduFZd5c0kQFFg6jJHfp3jfVOReGFFzeqOV3TWq-eL79Up8YxZ4Jt8INiTbaGWXF8sU7_kM7aJ_PcXuZ508q62QXoL6_GWIy1I2xjdYfgb3tHKLVZ10GiyMcve6Go8tPbeY2emaSBfVWZGxcJrMuy6qCuSFowfjbpzQJkfKxqXMZN1SNyUuXVkZnpWIqUBDAqBY_NCKRDKt_iwxIDsklXTl8ANsGIc_8FZXVJUKOX5pccs4KARu5_gLxqq14fKvAXV7bwsbSM-Xe03obWyMA"
Key                                       Value
---                                       -----
token                                     s.8eeuo1Cu5xCsIWKQ5WcNnTPn
token_accessor                            2czXQjEdRz2EAio2VJyrLAyD
token_duration                            16m40s
token_renewable                           true
token_policies                            ["default" "k8s.-.demo.demo-policy"]
identity_policies                         []
policies                                  ["default" "k8s.-.demo.demo-policy"]
token_meta_service_account_name           demo-sa
token_meta_service_account_namespace      demo
token_meta_service_account_secret_name    demo-sa-token-ddk7c
token_meta_service_account_uid            46448c8a-08f6-11e9-b1f2-080027b7189d
token_meta_role                           demo-role

$ export VAULT_TOKEN='s.8eeuo1Cu5xCsIWKQ5WcNnTPn'
$ vault kv put secret/foo A=B
Success! Data written to: secret/foo

$ vault kv get secret/foo
== Data ==
Key    Value
---    -----
A      B

$ vault auth list
Error listing enabled authentications: Error making API request.

URL: GET https://127.0.0.1:8200/v1/sys/auth
Code: 403. Errors:

* 1 error occurred:
	* permission denied

```

If we delete VaultPolicyBinding, then respective role will be deleted from Vault.

```console
$ kubectl delete vaultpolicybindings/demo-role -n demo 
vaultpolicybinding.policy.kubevault.com "demo-role" deleted
```

Check whether role is deleted from Vault.

```console
$ vault read auth/kubernetes/role/demo-role
No value found at auth/kubernetes/role/demo-role

$ vault list auth/kubernetes/role
Keys
----
k8s.-.demo.vault-auth-method-controller
vault-policy-controller
```
