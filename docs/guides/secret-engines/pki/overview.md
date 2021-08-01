---
title: Manage PKI(certificates) secrets using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-pki
    name: Overview
    parent: pki-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage PKI(certificates) secrets using the KubeVault operator

The [PKI secrets engine](https://www.vaultproject.io/docs/secrets/pki/index.html) generates dynamic X.509 certificates. With this secrets engine, services can get certificates without going through the usual manual process of generating a private key and CSR, submitting to a CA, and waiting for a verification and signing process to complete. Vault's built-in authentication and authorization mechanisms provide the verification functionality.

You can easily manage [PKI secret engine](https://www.vaultproject.io/docs/secrets/pki/index.html) using the KubeVault operator.

You should be familiar with the following CRD:

- [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md)
- [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md)
- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we are going to demonstrate the use of the PKI secret engine.

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
    path: kubernetes
    vaultRole: vault-policy-controller
    kubernetes:
      serviceAccountName: vault
      tokenReviewerServiceAccountName: vault-k8s-token-reviewer
      usePodServiceAccountForCSIDriver: true
```

## Use PKI Secret Engine as Root User

Here, we are going to use the Vault root token to perform authentication to the Vault server. We will use the [Vault CLI](https://www.vaultproject.io/docs/commands/#vault-commands-cli-) throughout the tutorial.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

Export the root token as environment variable:

```console
export VAULT_TOKEN=s.diWLjSzmfSmF0qUNYV3qOIeX
```

Enable the PKI secrets engine:

```console
$ vault secrets enable pki
Success! Enabled the pki secrets engine at: pki/
```

Increase the TTL by tuning the secrets engine. The default value of 30 days may be too short, so increase it to 1 year:

```console
$ vault secrets tune -max-lease-ttl=8760h pki
Success! Tuned the secrets engine at: pki/
```

Configure a CA certificate and private key:

```console
$ vault write pki/root/generate/internal \
                          common_name=my-website.com \
                          ttl=8760h
Key              Value
---              -----
certificate      -----BEGIN CERTIFICATE-----
MIIDPjCCAiagAwIBAgIUEDmnAmC0siISlrezD3/CeUXTSfswDQYJKoZIhvcNAQEL
... ... ... ... ... ... ... ... ... ... ... ... ... ... ... ...
CsFVu+vfMM9XEMYeKHRWAq9onJFyGuwKGhF0/7RbZ3EunTj6Zph+UMucGoL4xfXj
ITltdU1N4JPvihQq+8Omryay
-----END CERTIFICATE-----
expiration       1606200496
issuing_ca       -----BEGIN CERTIFICATE-----
MIIDPjCCAiagAwIBAgIUEDmnAmC0siISlrezD3/CeUXTSfswDQYJKoZIhvcNAQEL
... ... ... ... ... ... ... ... ... ... ... ... ... ... ... ...
CsFVu+vfMM9XEMYeKHRWAq9onJFyGuwKGhF0/7RbZ3EunTj6Zph+UMucGoL4xfXj
ITltdU1N4JPvihQq+8Omryay
-----END CERTIFICATE-----
serial_number    10:39:a7:02:60:b4:b2:22:12:96:b7:b3:0f:7f:c2:79:45:d3:49:fb
```

Configure a role that maps a name in Vault to a procedure for generating a certificate. When users or machines generate credentials, they are generated against this role:

```console
$ vault write pki/roles/example-dot-com \
                          allowed_domains=my-website.com \
                          allow_subdomains=true \
                          max_ttl=72h
Success! Data written to: pki/roles/example-dot-com
```

Generate a new credential by writing to the /issue endpoint with the name of the role:

```console
$ vault write pki/issue/example-dot-com \
                        common_name=www.my-website.com
Key                 Value
---                 -----
certificate         -----BEGIN CERTIFICATE-----
MIIDVjCCAj6gAwIBAgIUWQhPLW6R/nk/3x3XReHC1Ze4BWUwDQYJKoZIhvcNAQEL
... ... ... ... ... ... ... ... ... ... ... ... ... ... ... ...
TSuguIiSBt5NN0ou4aY01FbeJJOHZhtpj31XdXOCAKR40lPCmWtEUAbcuEhLlkm+
vmhNYxBqkx33jEIMxk95P4eKIYPyr45/8o7bV1jq7G26aBzj1Mjd0JmU
-----END CERTIFICATE-----
expiration          1574924103
issuing_ca          -----BEGIN CERTIFICATE-----
MIIDPjCCAiagAwIBAgIUEDmnAmC0siISlrezD3/CeUXTSfswDQYJKoZIhvcNAQEL
... ... ... ... ... ... ... ... ... ... ... ... ... ... ... ...
CsFVu+vfMM9XEMYeKHRWAq9onJFyGuwKGhF0/7RbZ3EunTj6Zph+UMucGoL4xfXj
ITltdU1N4JPvihQq+8Omryay
-----END CERTIFICATE-----
private_key         -----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAuK7V4GuoHSF8pnlr4hApeU7V3zpuQ2rWt3pXgi9TPBCmIuye
... ... ... ... ... ... ... ... ... ... ... ... ... ... ... ...
+o8HetGW5xWvuQ/ObkiSzdQ8nxMyiQj/whe4riYriOw1fYwPrjZfxTm1jsyEmbbm
gYewhfHP3hOgTCVu3SjhvOXS3pnW7hUP4wtvpLLdRumEUM/fK7pwNg==
-----END RSA PRIVATE KEY-----
private_key_type    rsa
serial_number       59:008:4f:2d:6e:91:fe:79:3f:df:1d:d7:45:e1:c2:d5:97:b8:05:65
```

For more details visit the [official Vault documentation](https://www.vaultproject.io/docs/secrets/pki/index.html#setup).

## Use PKI Secret Engine as Non-root User

Here, we are going to create a Kubernetes service account and give it limited access (i.e only PKI secret engine) from the Vault using the VaultPolicy and the VaultPolicyBinding.

### Create Kubernetes Service Account

Create a service account `pki-admin` to the `demo` namespace:

```console
$  kubectl create serviceaccount -n demo pki-admin
serviceaccount/pki-admin created

# get service account JWT token which will be required while performing
# login operation to the Vault
$ kubectl get secrets -n demo pki-admin-token-26kwb  -o jsonpath="{.data.token}" | base64 --decode;
eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZW1vIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6InBraS1hZG1pbi10b2tlbi0yNmt3YiIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJwa2ktYWRtaW4iLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiJkYmVkZDQ2Ni0yYzc0LTQ0OGItOTBlZS01MDlkNGI4MTJjOTEiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVtbzpwa2ktYWRtaW4ifQ.ce7OqA05nsfBMRsEOiG1Lje_mOBdUZRKALB9Sc9LVqjKIJZHdxvZ7NT4ZKrIyPEe02aItzxlXLAP4Fa8dUMshZuNyuxBYN7p2qHRCwVKHqOuz8LdRQWypKiLozL9v0DHk-vbFWFcm0eye57vJBFtriYyYRUA84WZhxRb9wz-f8z7PSmO2mpjkrICt7wi48j-4FObdhFWk6HAKXFD7bCzL4j3CWUcx2wTIsnOEz9SifjYZuGaog6tpWhnj-guEKpXJzBLAoMBU0Vr3U7Zv_z1qvKFF4ZherUBxSOMo27lL2xbhkpbW2wf_DCAjLx8pScoh9mxv7AK2WJCHeA0JRzrug
```

### Create VaultPolicy and VaultPolicyBinding

A sample VaultPolicy object with necessary path permission for the PKI secret engine:

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: pki-policy
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

    path "pki/*" {
        capabilities = ["read","create", "list", "update", "delete"]
    }

    path "sys/leases/revoke/*" {
        capabilities = ["update"]
    }
```

Create VaultPolicy and check status:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/pki/policy.yaml
vaultpolicy.policy.kubevault.com/pki-policy created

$ kubectl get vaultpolicy -n demo
NAME                           STATUS    AGE
pki-policy                     Success   3m15s
```

A sample VaultPolicyBinding object that binds the `pki-policy` to the `pki-admin` service account:

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: pki-admin-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  policies:
  - ref: pki-policy
  subjectRef:
    kubernetes:
      serviceAccountNames:
        - "pki-admin"
      serviceAccountNamespaces:
        - "demo"
      ttl: "1000"
      maxTTL: "2000"
      period: "1000"
```

Create VaultPolicyBinding and check status:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/pki/policyBinding.yaml
vaultpolicybinding.policy.kubevault.com/pki-admin-role created

$ kubectl get vaultpolicybindings -n demo
NAME                           STATUS    AGE
pki-admin-role                 Success   43m
```

### Login Vault and Use PKI Secret Engine

To resolve the naming conflict, name of the policy and role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

List Vault policies and Kubernetes auth roles:

```console
$ vault list sys/policy
Keys
----
k8s.-.demo.pki-policy

$ vault read sys/policy/k8s.-.demo.pki-policy
Key      Value
---      -----
name     k8s.-.demo.pki-policy
rules    path "sys/mounts" {
  capabilities = ["read", "list"]
}

path "sys/mounts/*" {
  capabilities = ["create", "read", "update", "delete"]
}

path "pki/*" {
    capabilities = ["read","create", "list", "update", "delete"]
}

path "sys/leases/revoke/*" {
    capabilities = ["update"]
}

$ vault list auth/kubernetes/role
Keys
----
k8s.-.demo.pki-admin-role

$ vault read auth/kubernetes/role/k8s.-.demo.pki-admin-role
Key                                 Value
---                                 -----
bound_service_account_names         [pki-admin]
bound_service_account_namespaces    [demo]
max_ttl                             33m20s
period                              16m40s
policies                            [k8s.-.demo.pki-policy]
token_bound_cidrs                   []
token_explicit_max_ttl              0s
token_max_ttl                       33m20s
token_no_default_policy             false
token_num_uses                      0
token_period                        16m40s
token_policies                      [k8s.-.demo.pki-policy]
token_ttl                           16m40s
token_type                          default
ttl                                 16m40s
```

So, we can see that the `pki-policy` is added to the `pki-admin-role`.

Now, login to the Vault using `pki-admin`'s JWT token under `pki-admin-role` role.

```console
$ vault write auth/kubernetes/login \
                       role=k8s.-.demo.pki-admin-role \
                       jwt=eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZW1vIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6InBraS1hZG1pbi10b2tlbi0yNmt3YiIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJwa2ktYWRtaW4iLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiJkYmVkZDQ2Ni0yYzc0LTQ0OGItOTBlZS01MDlkNGI4MTJjOTEiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVtbzpwa2ktYWRtaW4ifQ.ce7OqA05nsfBMRsEOiG1Lje_mOBdUZRKALB9Sc9LVqjKIJZHdxvZ7NT4ZKrIyPEe02aItzxlXLAP4Fa8dUMshZuNyuxBYN7p2qHRCwVKHqOuz8LdRQWypKiLozL9v0DHk-vbFWFcm0eye57vJBFtriYyYRUA84WZhxRb9wz-f8z7PSmO2mpjkrICt7wi48j-4FObdhFWk6HAKXFD7bCzL4j3CWUcx2wTIsnOEz9SifjYZuGaog6tpWhnj-guEKpXJzBLAoMBU0Vr3U7Zv_z1qvKFF4ZherUBxSOMo27lL2xbhkpbW2wf_DCAjLx8pScoh9mxv7AK2WJCHeA0JRzrug

Key                                       Value
---                                       -----
token                                     s.ZPu4zcyaajjpxtS1t8fnh2LV
token_accessor                            5OknOf72h8WnP1v0I1C01626
token_duration                            16m40s
token_renewable                           true
token_policies                            ["default" "k8s.-.demo.pki-policy"]
identity_policies                         []
policies                                  ["default" "k8s.-.demo.pki-policy"]
token_meta_role                           k8s.-.demo.pki-admin-role
token_meta_service_account_name           pki-admin
token_meta_service_account_namespace      demo
token_meta_service_account_secret_name    pki-admin-token-26kwb
token_meta_service_account_uid            dbedd466-2c74-448b-90ee-509d4b812c91
```

Export the new Vault token as an environment variable:

```console
export VAULT_TOKEN=s.ZPu4zcyaajjpxtS1t8fnh2LV
```

Now generate a new certificate using the PKI secret engine:

Enable the PKI secrets engine:

```console
$ vault secrets enable pki
Success! Enabled the pki secrets engine at: pki/
```

Increase the TTL by tuning the secrets engine. The default value of 30 days may be too short, so increase it to 1 year:

```console
$ vault secrets tune -max-lease-ttl=8760h pki
Success! Tuned the secrets engine at: pki/
```

Configure a CA certificate and private key:

```console
$ vault write pki/root/generate/internal \
                          common_name=my-website.com \
                          ttl=8760h
Key              Value
---              -----
certificate      -----BEGIN CERTIFICATE-----
MIIDPjCCAiagAwIBAgIUEDmnAmC0siISlrezD3/CeUXTSfswDQYJKoZIhvcNAQEL
... ... ... ... ... ... ... ... ... ... ... ... ... ... ... ...
CsFVu+vfMM9XEMYeKHRWAq9onJFyGuwKGhF0/7RbZ3EunTj6Zph+UMucGoL4xfXj
ITltdU1N4JPvihQq+8Omryay
-----END CERTIFICATE-----
expiration       1606200496
issuing_ca       -----BEGIN CERTIFICATE-----
MIIDPjCCAiagAwIBAgIUEDmnAmC0siISlrezD3/CeUXTSfswDQYJKoZIhvcNAQEL
... ... ... ... ... ... ... ... ... ... ... ... ... ... ... ...
CsFVu+vfMM9XEMYeKHRWAq9onJFyGuwKGhF0/7RbZ3EunTj6Zph+UMucGoL4xfXj
ITltdU1N4JPvihQq+8Omryay
-----END CERTIFICATE-----
serial_number    10:39:a7:02:60:b4:b2:22:12:96:b7:b3:0f:7f:c2:79:45:d3:49:fb
```

Configure a role that maps a name in Vault to a procedure for generating a certificate. When users or machines generate credentials, they are generated against this role:

```console
$ vault write pki/roles/example-dot-com \
                          allowed_domains=my-website.com \
                          allow_subdomains=true \
                          max_ttl=72h
Success! Data written to: pki/roles/example-dot-com
```

Generate a new credential by writing to the /issue endpoint with the name of the role:

```console
$ vault write pki/issue/example-dot-com \
                        common_name=www.my-website.com
Key                 Value
---                 -----
certificate         -----BEGIN CERTIFICATE-----
MIIDVjCCAj6gAwIBAgIUWQhPLW6R/nk/3x3XReHC1Ze4BWUwDQYJKoZIhvcNAQEL
... ... ... ... ... ... ... ... ... ... ... ... ... ... ... ...
TSuguIiSBt5NN0ou4aY01FbeJJOHZhtpj31XdXOCAKR40lPCmWtEUAbcuEhLlkm+
vmhNYxBqkx33jEIMxk95P4eKIYPyr45/8o7bV1jq7G26aBzj1Mjd0JmU
-----END CERTIFICATE-----
expiration          1574924103
issuing_ca          -----BEGIN CERTIFICATE-----
MIIDPjCCAiagAwIBAgIUEDmnAmC0siISlrezD3/CeUXTSfswDQYJKoZIhvcNAQEL
... ... ... ... ... ... ... ... ... ... ... ... ... ... ... ...
CsFVu+vfMM9XEMYeKHRWAq9onJFyGuwKGhF0/7RbZ3EunTj6Zph+UMucGoL4xfXj
ITltdU1N4JPvihQq+8Omryay
-----END CERTIFICATE-----
private_key         -----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAuK7V4GuoHSF8pnlr4hApeU7V3zpuQ2rWt3pXgi9TPBCmIuye
... ... ... ... ... ... ... ... ... ... ... ... ... ... ... ...
+o8HetGW5xWvuQ/ObkiSzdQ8nxMyiQj/whe4riYriOw1fYwPrjZfxTm1jsyEmbbm
gYewhfHP3hOgTCVu3SjhvOXS3pnW7hUP4wtvpLLdRumEUM/fK7pwNg==
-----END RSA PRIVATE KEY-----
private_key_type    rsa
serial_number       59:008:4f:2d:6e:91:fe:79:3f:df:1d:d7:45:e1:c2:d5:97:b8:05:65
```

For more details visit the [official Vault documentation](https://www.vaultproject.io/docs/secrets/pki/index.html#setup).
