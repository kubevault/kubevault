---
title: Manage PKI(certificates) secrets using the Vault Operator
menu:
  docs_0.2.0:
    identifier: overview-pki
    name: Overview
    parent: pki-secret-engines
    weight: 10
menu_name: docs_0.2.0
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
    image: kubevault/vault-exporter:0.2.0
  unsealer:
    image: kubevault/vault-unsealer:0.2.0
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

## Create Policy for PKI secrets

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


## Read/Write secrets into Vault

From your local machine check the Vault server is running with following command:

```console
$ kubectl get pods -l app=vault -n demo
NAME                     READY   STATUS    RESTARTS   AGE
vault-848797ffdf-xdnn8   3/3     Running   0          8m44s
```

Vault server is running on port 8200. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Vault server from local machine. Run following commands on a separate terminal,

```console
$ kubectl port-forward -n vault-848797ffdf-xdnn8 8200
Forwarding from 127.0.0.1:8200 -> 8200
Forwarding from [::1]:8200 -> 8200
```

Now, you can access the Vault server at https://127.0.0.1:8200.

To retrieve `CACert` of Vault server run following command:

```console
$ kubectl get pods vault-848797ffdf-xdnn8 -n demo -o jsonpath='{.spec.containers[?(@.name=="vault-unsealer")].args}'
[run --v=3 --secret-shares=4 --secret-threshold=2 --vault.ca-cert=-----BEGIN CERTIFICATE-----
MIICuDCCAaCgAwIBAgIBADANBgkqhkiG9w0BAQsFADANMQswCQYDVQQDEwJjYTAe
Fw0xOTAzMDEwNTI1MzlaFw0yOTAyMjYwNTI1MzlaMA0xCzAJBgNVBAMTAmNhMIIB
IjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsd+CM6/GpA13afJpIjnL+2B7
kjP6EnkIkrOaVm2tSf61r4HknjZYmENLvuiByCAwUIcWa+qa6LXgAQ+bV2EYOpyA
uU1oJAphR2ARJsAjrzKDPtOLLu00/gCY6fJ4ueelwV2HlPIqjKTKZQHm6/yFCbp3
mnTmGSf0kYGefcuf1BfZsA3wWKy9uetom8OHkUe+ufWGcbSVEVuGTV5jfbVZ/uo+
AiNuR+qc4N1hIIVdVJxc98I2FTiII1vMYk7GjwubDcudxXzuKAYbJpY8No89Y8OT
YDCl5YCILZyssMlRSa31S65nMJsZjkjKRtxMqIDCcpWcCO5Ij/qfoUexqNgQZQID
AQABoyMwITAOBgNVHQ8BAf8EBAMCAqQwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG
9w0BAQsFAAOCAQEAkBidD4V8vWHwnBip4psyMLdHG08H1KzcsDOfZ1n2q957Pfb9
f2A0aq6My0/TdSEQaEHeOSruQonQDvnUOSZ0JDhjf5aKssggwzbHCmS6JqVEm+eb
RW0OHJepJj382umj5qP/dUKBRM+cM56S1aheVw+H9cR4ltX2kRw98nPi7Ilwbd8a
pjrlb2brUszDjaR0DGVOoeiSPFw7qv1EQA8xhu7a+K6woKwwd8a/VrRgfDQkeTLH
R1RqEYe1Uk5t2sGIQ2q1ymWQfl2218P4Hh1TpAF4gzDrc5t3VOThy3ZLS9i2/XIu
89cKx7f8pb6/ybfCARI96S+WcNbZHv+SKEL2MQ==
-----END CERTIFICATE-----
 --auth.k8s-host=https://10.96.0.1:443 ...
```

From the output grep `vault.ca-cert` key and store the value int `ca.crt` file.

We need to configure following environment variable.

```console
export VAULT_ADDR=https://127.0.0.1:8200
export VAULT_TOKEN=$(kubectl get secrets -n demo vault-keys -o jsonpath="{.data.vault-root-token}" | base64 --decode; echo)

export VAULT_CACERT=ca.crt #Put the path where you stored ca.crt

$ kubectl get secrets -n demo vault-vault-tls -o jsonpath="{.data.tls\.crt}" | base64 -d>tls.crt
export VAULT_CLIENT_CERT=tls.crt

$ kubectl get secrets -n demo vault-vault-tls -o jsonpath="{.data.tls\.key}" | base64 -d>tls.key
export VAULT_CLIENT_KEY=tls.key

```

Check whether Vault server can be accessed

```console
$ vault status
Key             Value
---             -----
Seal Type       shamir
Sealed          false
Total Shares    4
Threshold       2
Version         1.0.0
Cluster Name    vault-cluster-1bfbb939
Cluster ID      3db2acdf-28b6-8afb-ed52-fed6cf55379d
HA Enabled      false
```

To enable `PKI` engine, run:

```console
$ vault secrets enable pki
Success! Enabled the pki secrets engine at: pki/
```

#### Write Certificate Data

To write certificate into Vault run following command:

```console
$ vault write pki/root/generate/internal \
    common_name=my-website.com \
    ttl=8760h
```

#### Issue Certificate Data

To issue certificate from Vault run following command:

```console
$ vault write pki/issue/my-role \
    common_name=www.my-website.com

Key                 Value
---                 -----
certificate         -----BEGIN CERTIFICATE-----...
issuing_ca          -----BEGIN CERTIFICATE-----...
private_key         -----BEGIN RSA PRIVATE KEY-----...
private_key_type    rsa
serial_number       1d:2e:c6:06:45:18:60:0e:23:d6:c5:17:43:c0:fe:46:ed:d1:50:be
```

To learn more usages of Vault `PKI(certificates)` secret engine click [this](https://www.vaultproject.io/docs/secrets/pki/index.html#setup).