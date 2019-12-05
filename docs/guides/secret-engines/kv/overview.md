---
title: Manage Key/Value Secrets using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-kv
    name: Overview
    parent: kv-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Key/Value Secrets using the KubeVault operator

You can easily manage [KV secret engine](https://www.vaultproject.io/docs/secrets/kv/index.html#kv-secrets-engine) using the KubeVault operator.

You should be familiar with the following CRD:

- [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md)
- [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md)
- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/operator/install.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we are going to demonstrate the use of the KV secret engine.

## Vault Server

If you don't have a Vault Server, you can deploy it by using the KubeVault operator.

- [Deploy Vault Server](/docs/guides/vault-server/vault-server.md)

The KubeVault operator can manage policies and secret engines of Vault servers which are not provisioned by the KubeVault operator. You need to configure both the Vault server and the cluster so that the KubeVault operator can communicate with your Vault server.

- [Configure cluster and Vault server](/docs/guides/vault-server/external-vault-sever.md#configuration)

Now, we have the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains connection and authentication information about the Vault server.

```console
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
    authMethodControllerRole: k8s.-.demo.vault-auth-method-controller
    path: kubernetes
    policyControllerRole: vault-policy-controller
    serviceAccountName: vault
    tokenReviewerServiceAccountName: vault-k8s-token-reviewer
    usePodServiceAccountForCsiDriver: true
```

## Use KV Secret Engine as Root User

Here, we are going to use the Vault root token to perform authentication to the Vault server. We will use the [Vault CLI](https://www.vaultproject.io/docs/commands/#vault-commands-cli-) throughout the tutorial.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

Export the root token as environment variable:

```console
export VAULT_TOKEN=s.lbSCc2GGit1QmqghBgYgjbOG
```

### Enable KV Secret Engine

Enable the KV secret engine:

```console
$ vault secrets enable -version=1 kv
Success! Enabled the kv secrets engine at: kv/
```

### Write KV Secrets

Write arbitrary key-value pairs:

```console
$ vault kv put kv/my-secret my-value=s3cr3t
Success! Data written to: kv/my-secret

$ vault kv put kv/secret-name secret-value=8HI.HFDJK324
Success! Data written to: kv/secret-name

$ vault kv put  kv/key value=sdfkjdslkfjdslj
Success! Data written to: kv/key
```

### List KV Secrets

List key-value pairs:

```console
$ vault kv list kv/
Keys
----
key
my-secret
secret-name
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

### Delete KV Secrets

Delete a specific key-value pair:

```console
$ vault kv delete kv/my-secret
Success! Data deleted (if it existed) at: kv/my-secret
```

## Use KV Secret Engine as Non-root User

Here, we are going to create a Kubernetes service account and give it limited access (i.e only KV secret engine) from the Vault using the VaultPolicy and the VaultPolicyBinding.

### Create Kubernetes Service Account

Create a service account `kv-admin` to the `demo` namespace:

```console
$ kubectl create serviceaccount -n demo kv-admin
serviceaccount/kv-admin created

# get service account JWT token which will be required while performing
# login operation to the Vault
$ kubectl get secret -n demo kv-admin-token-8cgr2 -o jsonpath="{.data.token}" | base64 -d
eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZW1vIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6Imt2LWFkbWluLXRva2VuLThjZ3IyIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6Imt2LWFkbWluIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiMjhiNDdlMWQtMzQyZC00MjYyLWI0NDItMzRjYzViOTFhYThlIiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OmRlbW86a3YtYWRtaW4ifQ.NkAbcuOsziZCtDtUYuxzuCKcAVuywnIbdEHylB1un6yc5K_Qfl_AtsnuKjWbJDZtp1kjc6bwy6dftMPSoPwd6U9FO5kbGbLqoA6vsa3Y_gJ74dhYqZnGHZZg9KpCxLHxvl8phcjIrRMvKW_dn95p334GWSI_AqU1zvGTQnFhjlrb-NRKpeTA7N7Y1JP2x1wB8KdtHha-qqGmLsUMJbc8VebgKnG8zjhc1KfgtO0lMLt4uLthBS4ca10r4fOsz259n66FOkVPfbPnXlUYzeObz-Ng4cFwdZ6xLgdF2wz9e8pTKXhe8NifzTFMk_44TPpE5pBqsog80lfMuq7Tk4O3TQ
```

### Create VaultPolicy and VaultPolicyBinding

A sample VaultPolicy object with necessary path permission for KV secret engine:

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: kv-policy
  namespace: demo
spec:
  vaultRef:
    name: vault
  policyDocument: |
    path "sys/mounts" {
      capabilities = ["read", "list"]
    }

    path "sys/mounts/*" {
      capabilities = ["create", "read", "update", "delete"]
    }

    path "kv/*" {
        capabilities = ["create","list", "read", "update", "delete"]
    }

    path "sys/leases/revoke/*" {
        capabilities = ["update"]
    }
```

Create VaultPolicy and check status:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/kv/policy.yaml
vaultpolicy.policy.kubevault.com/kv-policy created

$ kubectl get vaultpolicy -n demo
NAME                           STATUS    AGE
kv-policy                      Success   8m51s
```

A sample VaultPolicyBinding object that binds the `kv-policy` to the `kv-admin` service account:

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: kv-admin-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  policies:
  - ref: kv-policy
  subjectRef:
    kubernetes:
      serviceAccountNames:
        - "kv-admin"
      serviceAccountNamespaces:
        - "demo"
      ttl: "1000"
      maxTTL: "2000"
      period: "1000"
```

Create VaultPolicyBinding and check status:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/kv/policyBinding.yaml
vaultpolicybinding.policy.kubevault.com/kv-admin-role created

$ kubectl get vaultpolicybindings -n demo
NAME                           STATUS    AGE
kv-admin-role                  Success   4m56s
```

### Login Vault and Use KV Secret Engine

To resolve the naming conflict, name of the policy and role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

List Vault policies and Kubernetes auth roles:

```console
$ vault list sys/policy
Keys
----
k8s.-.demo.kv-policy
k8s.-.demo.vault-auth-method-controller
vault-policy-controller

$ vault list auth/kubernetes/role
Keys
----
k8s.-.demo.kv-admin-role
k8s.-.demo.vault-auth-method-controller
vault-policy-controller

$ vault read auth/kubernetes/role/k8s.-.demo.kv-admin-role
Key                                 Value
---                                 -----
bound_service_account_names         [kv-admin]
bound_service_account_namespaces    [demo]
max_ttl                             33m20s
period                              16m40s
policies                            [k8s.-.demo.kv-policy]
token_bound_cidrs                   []
token_explicit_max_ttl              0s
token_max_ttl                       33m20s
token_no_default_policy             false
token_num_uses                      0
token_period                        16m40s
token_policies                      [k8s.-.demo.kv-policy]
token_ttl                           16m40s
token_type                          default
ttl                                 16m40s
```

So, we can see that the `kv-policy` is added to the `kv-admin-role`.

Now, login to the Vault using `kv-admin`'s JWT token under `kv-admin-role` role.

```console
$ vault write auth/kubernetes/login \
        role=k8s.-.demo.kv-admin-role \
        jwt=eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZW1vIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6Imt2LWFkbWluLXRva2VuLThjZ3IyIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6Imt2LWFkbWluIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiMjhiNDdlMWQtMzQyZC00MjYyLWI0NDItMzRjYzViOTFhYThlIiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OmRlbW86a3YtYWRtaW4ifQ.NkAbcuOsziZCtDtUYuxzuCKcAVuywnIbdEHylB1un6yc5K_Qfl_AtsnuKjWbJDZtp1kjc6bwy6dftMPSoPwd6U9FO5kbGbLqoA6vsa3Y_gJ74dhYqZnGHZZg9KpCxLHxvl8phcjIrRMvKW_dn95p334GWSI_AqU1zvGTQnFhjlrb-NRKpeTA7N7Y1JP2x1wB8KdtHha-qqGmLsUMJbc8VebgKnG8zjhc1KfgtO0lMLt4uLthBS4ca10r4fOsz259n66FOkVPfbPnXlUYzeObz-Ng4cFwdZ6xLgdF2wz9e8pTKXhe8NifzTFMk_44TPpE5pBqsog80lfMuq7Tk4O3TQ
Key                                       Value
---                                       -----
token                                     s.HJ8owGJLrqzlnA8tKuYdrElh
token_accessor                            FHN3pCvTAoyZuq7FZoOe1fSc
token_duration                            16m40s
token_renewable                           true
token_policies                            ["default" "k8s.-.demo.kv-policy"]
identity_policies                         []
policies                                  ["default" "k8s.-.demo.kv-policy"]
token_meta_role                           k8s.-.demo.kv-admin-role
token_meta_service_account_name           kv-admin
token_meta_service_account_namespace      demo
token_meta_service_account_secret_name    kv-admin-token-8cgr2
token_meta_service_account_uid            28b47e1d-342d-4262-b442-34cc5b91aa8e
```

Export the new Vault token as an environment variable:

```console
export VAULT_TOKEN=s.HJ8owGJLrqzlnA8tKuYdrElh
```

Now perform read, write, list and delete operation on KV secret engine:

```console
# Enable KV secret engine
$ vault secrets enable -version=1 kv
Success! Enabled the kv secrets engine at: kv/

# Write KV secret
$ vault kv put kv/my-secret my-value=s3cr3t
Success! Data written to: kv/my-secret

# List KV secrets
$ vault kv list kv/
Keys
----
my-secret

# Read KV secret
$ vault kv get kv/my-secret
====== Data ======
Key         Value
---         -----
my-value    s3cr3t

# Delete KV secret
$ vault kv delete kv/my-secret
Success! Data deleted (if it existed) at: kv/my-secret
```

To learn more usages of Vault `Key/Vaule` secret engine click [this](https://www.vaultproject.io/docs/secrets/kv/kv-v1.html#usage).