---
title: Mount PKI(certificates) Secrets into Kubernetes pod using CSI Driver
menu:
  docs_{{ .version }}:
    identifier: csi-driver-pki
    name: CSI Driver
    parent: pki-secret-engines
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Mount PKI(certificates) Secrets using CSI Driver

At first, you need to have a Kubernetes 1.16 or later cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/). To check the version of your cluster, run:

```console
$ kubectl version --short
Client Version: v1.21.2
Server Version: v1.21.1
```

Before you begin:

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Install Secrets Store CSI driver for Kubernetes secrets in your cluster from [here](https://secrets-store-csi-driver.sigs.k8s.io/getting-started/installation.html).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial stored in [examples](/docs/examples/guides/secret-engines/pki) folder in GitHub repository [KubeVault/docs](https://github.com/kubevault/kubevault)

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

## Enable and Configure PKI Secret Engine

We will use the [Vault CLI](https://www.vaultproject.io/docs/commands/#vault-commands-cli-) throughout the tutorial to [enable and configure](https://www.vaultproject.io/docs/secrets/pki/index.html#setup) the PKI secret engine.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

To use secret from `PKI` secret engine, you have to perform the following steps.

### Enable PKI Secret Engine

To enable `PKI` secret engine run the following command.

```console
$ vault secrets enable pki
Success! Enabled the pki secrets engine at: pki/
```

Increase the TTL by tuning the secrets engine. The default value of 30 days may be too short, so increase it to 1 year:

```console
$ vault secrets tune -max-lease-ttl=8760h pki
Success! Tuned the secrets engine at: pki/
```

### Configure CA Certificate and Private Key

Configure a CA certificate and private key. Vault can accept an existing key pair, or it can generate its own self-signed root.

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

### Configure a PKI Role

We need to configure a role that maps a name in vault to a procedure for generating certificate. When users of machines generate credentials, they are generated agains this role:

```console
$ vault write pki/roles/example-dot-com \
                          allowed_domains=my-website.com \
                          allow_subdomains=true \
                          max_ttl=72h
Success! Data written to: pki/roles/example-dot-com
```

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
$ kubectl apply -f docs/examples/guides/secret-engines/pki/serviceaccount.yaml
serviceaccount/test-user-account created
```

### Create VaultPolicy and VaultPolicyBinding for Pod's Service Account
When a VaultPolicyBinding object is created, the KubeVault operator create an auth role in the Vault server. The role name is generated by the following naming format: `k8s.(clusterName or -).namespace.name`. Here, it is `k8s.-.demo.postgres-reader-role`. We need to provide the auth role name as service account `annotations` while creating the service account. If the annotation `secrets.csi.kubevault.com/vault-role` is not provided, the CSI driver will not be able to perform authentication to the Vault.

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: pki-se-policy
  namespace: demo
spec:
  vaultRef:
    name: vault
  policyDocument: |
    path "pki/issue/*" {
      capabilities = ["update"]
    }
---
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: pki-se-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  policies:
    - ref: pki-se-policy
  subjectRef:
    kubernetes:
      serviceAccountNames:
        - "test-user-account"
      serviceAccountNamespaces:
        - "demo"
```

Let's create VaultPolicy and VaultPolicyBinding:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/pki/policy.yaml
vaultpolicy.policy.kubevault.com/pki-se-policy created

$ kubectl apply -f docs/examples/guides/secret-engines/pki/policybinding.yaml
vaultpolicybinding.policy.kubevault.com/pki-se-role created
```

Check if the VaultPolicy and the VaultPolicyBinding are successfully registered to the Vault server:

```console
$ kubectl get vaultpolicy -n demo
NAME                           STATUS    AGE
pki-se-policy                  Success   8s

$ kubectl get vaultpolicybinding -n demo
NAME                          STATUS    AGE
pki-se-role                   Success   10s
```

## Mount Certificates into a Kubernetes Pod

So, we can create `SecretProviderClass` now.

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
    roleName: "k8s.-.demo.pki-se-role"
    objects: |
      - objectName: "certificate"
        secretPath: "pki/issue/example-dot-com"
        secretKey: "certificate"
        secretArgs:
          common_name: "www.my-website.com"
          ttl: 24h
        method: "POST"

      - objectName: "issuing_ca"
        secretPath: "pki/issue/example-dot-com"
        secretKey: "issuing_ca"
        secretArgs:
          common_name: "www.my-website.com"
          ttl: 24h
        method: "POST"

      - objectName: "private_key"
        secretPath: "pki/issue/example-dot-com"
        secretKey: "private_key"
        secretArgs:
          common_name: "www.my-website.com"
          ttl: 24h
        method: "POST"

      - objectName: "private_key_type"
        secretPath: "pki/issue/example-dot-com"
        secretKey: "private_key_type"
        secretArgs:
          common_name: "www.my-website.com"
          ttl: 24h
        method: "POST"
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/pki/secretproviderclass.yaml
secretproviderclass.secrets-store.csi.x-k8s.io/vault-db-provider created
```

Here, you can also pass the following parameters optionally to issue the certificate

- `common_name` (string: <required>) – Specifies the requested CN for the certificate. If the CN is allowed by role policy, it will be issued.

- `alt_names` (string: "") – Specifies requested Subject Alternative Names, in a comma-delimited list. These can be host names or email addresses; they will be parsed into their respective fields. If any requested names do not match role policy, the entire request will be denied.

- `ip_sans` (string: "") – Specifies requested IP Subject Alternative Names, in a comma-delimited list. Only valid if the role allows IP SANs (which is the default).

- `uri_sans` (string: "") – Specifies the requested URI Subject Alternative Names, in a comma-delimited list.

- `other_sans` (string: "") – Specifies custom OID/UTF8-string SANs. These must match values specified on the role in allowed_other_sans (globbing allowed). The format is the same as OpenSSL: <oid>;<type>:<value> where the only current valid type is UTF8. This can be a comma-delimited list or a JSON string slice.

- `ttl` (string: "") – Specifies requested Time To Live. Cannot be greater than the role's max_ttl value. If not provided, the role's ttl value will be used. Note that the role values default to system values if not explicitly set.

- `format` (string: "") – Specifies the format for returned data. Can be pem, der, or pem_bundle; defaults to pem. If der, the output is base64 encoded. If pem_bundle, the certificate field will contain the private key and certificate, concatenated; if the issuing CA is not a Vault-derived self-signed root, this will be included as well.

- `private_key_format` (string: "") – Specifies the format for marshaling the private key. Defaults to der which will return either base64-encoded DER or PEM-encoded DER, depending on the value of format. The other option is pkcs8 which will return the key marshalled as PEM-encoded PKCS8.

- `exclude_cn_from_sans` (bool: false) – If true, the given common_name will not be included in DNS or Email Subject Alternate Names (as appropriate). Useful if the CN is not a hostname or email address, but is instead some human-readable identifier.

### Create Pod

Now we can create a Pod which refers to this volume. When the Pod is created, the volume will be attached, formatted and mounted to the specific container.

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
          mountPath: "/secrets-store/pki-assets"
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
$ kubectl apply -f docs/examples/guides/secret-engines/pki/pod.yaml
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
$ kubectl exec -it -n test pod/demo-app -- /bin/sh

/ # ls /secrets-store/pki-assets
certificate       issuing_ca        private_key       private_key_type

/ # cat /secrets-store/pki-assets/certificate
-----BEGIN CERTIFICATE-----
MIIDVjCCAj6gAwIBAgIUNjTBC3qR7Zaj0XrzUc3QEbE+EhgwDQYJKoZIhvcNAQEL
BQAwGTEXMBUGA1UEAxMObXktd2Vic2l0ZS5jb20wHhcNMTkxMjEzMTExNDIwWhcN
..... .... .... .... .... .... .... .... .... .... .... .... ...
bo901cITjNyCTbAF2401pYFZ4rSlxhcuAvc7c27uqvKEh2/ctMGRkvPVygbPdvB8
LfCskfX0sk8PQiEznlmYlChK3KNsEp+xSCyjU+pDEw8AcDXwE6vVFft/fRX0oiHH
KIzTZ7R/QKUkLisloMUHStINISAehglLZTJjo79jB7GN66wyqP+E8iRLEYFAAsb0
aZ5wuSTYEpqOuP6G1tOdhiE7iptFu9Wg9dKtmXkZnc0iTBL60xMUUapH
-----END CERTIFICATE-----
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted

```
