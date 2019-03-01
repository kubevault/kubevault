---
title: Manage PKI(certificates) secrets using the Vault Operator
menu:
  docs_0.1.0:
    identifier: overview-pki
    name: Overview
    parent: pki-secret-engines
    weight: 10
menu_name: docs_0.1.0
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage PKI(certificates) secrets using the Vault Operator

You can easily manage [PKI secret engine](https://www.vaultproject.io/docs/secrets/pki/index.html#pki-secrets-engine) using Vault operator.

You should be familiar with the following CRD:

- [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md)
- [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md)
- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)

Before you begin:

- Install Vault operator in your cluster following the steps [here](/docs/setup/operator/install).

- Deploy Vault. It could be in the Kubernetes cluster or external.

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

For this tutorial, we are going to deploy Vault using Vault operator.

```console
$ cat examples/guides/secret-engins/pki/vault.yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  nodes: 1
  version: "1.0.0"
  backend:
    inmem: {}
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      kubernetesSecret:
        secretName: vault-keys

$ cat examples/guides/secret-engins/pki/vaultserverversion.yaml
apiVersion: catalog.kubevault.com/v1alpha1
kind: VaultServerVersion
metadata:
  name: 1.0.0
spec:
  exporter:
    image: kubevault/vault-exporter:0.1.0
  unsealer:
    image: kubevault/vault-unsealer:0.1.0
  vault:
    image: vault:1.0.0
  version: 1.0.0

$ kubectl apply -f docs/examples/guides/secret-engins/pki/vaultserverversion.yaml 
vaultserverversion.catalog.kubevault.com/1.0.0 created

$ kubectl apply -f examples/guides/secret-engins/pki/vault.yaml
vaultserver.kubevault.com/vault created

$ kubectl get vaultserver/vault -n demo
NAME      NODES     VERSION   STATUS    AGE
vault     1         1.0.0     Running   1h
```

## Key/Value Policy create

In this tutorial, we are going to create policy `pki-policy` and policybinding  `pki-role`.

```yaml
$ cat examples/guides/secret-engins/pki/vault-app.yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9ERXlNamN3TkRVNU1qVmFGdzB5T0RFeU1qUXdORFU1TWpWYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMVhid2wyQ1NNc2VQTU5RRzhMd3dUVWVOCkI1T05oSTlDNzFtdUoyZEZjTTlUc1VDQnlRRk1weUc5dWFvV3J1ZDhtSWpwMVl3MmVIUW5udmoybXRmWGcrWFcKSThCYkJUaUFKMWxMMFE5MlV0a1BLczlXWEt6dTN0SjJUR1hRRDhhbHZhZ0JrR1ViOFJYaUNqK2pnc1p6TDRvQQpNRWszSU9jS0xnMm9ldFZNQ0hwNktpWTBnQkZiUWdJZ1A1TnFwbksrbU02ZTc1ZW5hWEdBK2V1d09FT0YwV0Z2CmxGQmgzSEY5QlBGdTJKbkZQUlpHVDJKajBRR1FNeUxodEY5Tk1pZTdkQnhiTWhRVitvUXp2d1EvaXk1Q2pndXQKeDc3d29HQ2JtM0o4cXRybUg2Tjl6Tlc3WlR0YTdLd05PTmFoSUFEMSsrQm5rc3JvYi9BYWRKT0tMN2dLYndJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFXeWFsdUt3Wk1COWtZOEU5WkdJcHJkZFQyZnFTd0lEOUQzVjN5anBlaDVCOUZHN1UKSS8wNmpuRVcyaWpESXNHNkFDZzJKOXdyaSttZ2VIa2Y2WFFNWjFwZHRWeDZLVWplWTVnZStzcGdCRTEyR2NPdwpxMUhJb0NrekVBMk5HOGRNRGM4dkQ5WHBQWGwxdW5veWN4Y0VMeFVRSC9PRlc4eHJxNU9vcXVYUkxMMnlKcXNGCmlvM2lJV3EvU09Yajc4MVp6MW5BV1JSNCtSYW1KWjlOcUNjb1Z3b3R6VzI1UWJKWWJ3QzJOSkNENEFwOUtXUjUKU2w2blk3NVMybEdSRENsQkNnN2VRdzcwU25seW5mb3RaTUpKdmFzbStrOWR3U0xtSDh2RDNMMGNGOW5SOENTSgpiTjBiZzczeVlWRHgyY3JRYk0zcko4dUJnY3BsWlRpUy91SXJ2QT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    service:
      name: vault
      port: 8200
      scheme: HTTPS
  parameters:
    serviceAccountName: demo-sa
    policyControllerRole: pki-role
    authPath: kubernetes

$ kubectl apply -f examples/guides/secret-engins/pki/vault-app.yaml
appbinding.appcatalog.appscode.com/vault-app created
```

You need to create `demo-sa` serviceaccount by running following command:

```console
$ kubectl create serviceaccount -n demo demo-sa
serviceaccount/demo-sa created
```

`demo-sa` serviceaccount in the above AppBinding need to have the policy with following capabilities in Vault.

```hcl
path "sys/mounts" {
  capabilities = ["read", "list"]
}

path "sys/mounts/*" {
  capabilities = ["create", "read", "update", "delete"]
}

path "pki/*" {
        capabilities = ["read"]
}

path "sys/leases/revoke/*" {
    capabilities = ["update"]
}
```

You can manage policy in Vault using Vault operator, see [here](/docs/guides/policy-management/policy-management).

To create policy with above capabilities run following command

```console
$ kubectl apply -f examples/guides/secret-engins/pki/policy.yaml
vaultpolicy.policy.kubevault.com/pki-policy created
vaultpolicybinding.policy.kubevault.com/pki-role created
```
