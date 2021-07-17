---
title: Vault Server
menu:
  docs_{{ .version }}:
    identifier: vault-server
    name: Vault Server
    parent: vault-server-guides
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Vault Server

You can easily deploy and manage [HashiCorp Vault](https://www.vaultproject.io/) in the Kubernetes cluster using KubeVault operator. In this tutorial, we are going to deploy Vault on the Kubernetes cluster using KubeVault operator.

![Vault Server](/docs/images/guides/vault-server/vault_server_guide.svg)

## Before you begin

- Install KubeVault operator in your cluster following the steps [here](/docs/setup/operator/install.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

## Deploy Vault Server

To start with this tutorial, you need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [VaultServerVersion](/docs/concepts/vault-server-crds/vaultserverversion.md)
- [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md)

### Deploy VaultServerVersion

By installing KubeVault operator, you have already deployed some VaultServerVersion crds named after
the Vault image tag its using. You can list them by using the following command:

```console
$ kubectl get vaultserverversions
NAME    VERSION   VAULT_IMAGE   DEPRECATED   AGE
1.2.0   1.2.0     vault:1.2.0   false        38s
1.2.2   1.2.2     vault:1.2.2   false        38s
1.2.3   1.2.3     vault:1.2.3   false        38s
```

Now you can use them or deploy your own version by yourself:

```yaml
apiVersion: catalog.kubevault.com/v1alpha1
kind: VaultServerVersion
metadata:
  name: 1.2.2
spec:
  exporter:
    image: kubevault/vault-exporter:v0.1.0
  unsealer:
    image: kubevault/vault-unsealer:v0.3.0
  vault:
    image: vault:1.2.2
  version: 1.2.2
```

Deploy VaultServerVersion `1.2.1`:

```console
$ kubectl apply -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/vault-server/vaultserverversion.yaml
vaultserverversion.catalog.kubevault.com/1.2.1 created
```

### Deploy VaultServer

Once you have deployed VaultServerVersion, you are ready to deploy VaultServer.

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  replicas: 1
  version: "1.2.3"
  serviceTemplate:
    spec:
      type: NodePort
  backend:
    inmem: {}
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      kubernetesSecret:
        secretName: vault-keys
```

Here we are using `inmem` backend which will lose data when Vault server pods are restarted. For production setup, use an appropriate backend. For more information about supported **backends** and **unsealer options** visit `VaultServer` CRD [documentation](/docs/concepts/vault-server-crds/vaultserver.md)

Deploy `VaultServer`:

```console
$ kubectl apply -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/vault-server/vaultserver.yaml
vaultserver.kubevault.com/vault created
```

Check VaultServer status:

```console
$ kubectl get vaultserver -n demo
NAME    NODES   VERSION   STATUS       AGE
vault   1       1.2.3     Processing   47s
$ kubectl get vaultserver -n demo
NAME    NODES   VERSION   STATUS   AGE
vault   1       1.2.3     Sealed   54s
$ kubectl get vaultserver -n demo
NAME    NODES   VERSION   STATUS    AGE
vault   1       1.2.3     Running   68s
```

Since the status is `Running` that means you have deployed the Vault server successfully. Now, you are ready to use with this Vault server.

On creation of `VaultServer` object, the KubeVault operator performs the following tasks:

- Creates a `deployment` for Vault named after VaultServer crd

    ```console
    $ kubectl get deployment -n demo
    NAME    READY   UP-TO-DATE   AVAILABLE   AGE
    vault   1/1     1            1           25m
    ```

- Creates a `service` to communicate with vault pod/pods

    ```console
    $ kubectl get services -n demo
    NAME    TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)                         AGE
    vault   NodePort   10.110.35.39   <none>        8200:32580/TCP,8201:30062/TCP   20m
    ```

- Creates an `AppBinding` that holds connection information for this Vault server.

    ```console
    $ kubectl get appbindings -n demo
    NAME    AGE
    vault   30m
    ```

- Creates a `ServiceAccount` which will be used by the AppBinding for performing authentication.

    ```console
    $ kubectl get sa -n demo
    NAME                       SECRETS   AGE
    vault                      1         36m
    ```

- Unseals Vault and stores the Vault root token. For `kubernetesSecret` mode, the operator creates a k8s secret containing root token.

    ```console
    $ kubectl get secrets -n demo
    NAME                                   TYPE                                  DATA   AGE
    vault-keys                             Opaque                                5      42m
    ```

- Enables `Kubernetes auth method` and creates k8s auth role with Vault policies for the `service account`(here 'vault') on Vault.

## Enable Vault CLI

> Don't have the Vault binary? Download from [here](https://www.vaultproject.io/downloads.html).

If you want to communicate with the Vault servers using [Vault (CLI)](https://www.vaultproject.io/docs/commands/), perform the following commands:

Get your desire Vault server pod name:

```console
$ kubectl get pods -n demo -l=app.kubernetes.io/name=vault-operator
NAME                    READY   STATUS    RESTARTS   AGE
vault-8679f4cbf-v78cs   3/3     Running   0          93m
```

Perform port-forwarding:

```console
$ kubectl port-forward -n demo pod/vault-8679f4cbf-v78cs 8200
Forwarding from 127.0.0.1:8200 -> 8200
Forwarding from [::1]:8200 -> 8200
...
```

Now, you can access the Vault server at `https://localhost:8200`.

Retrieve the Vault server CA certificate from the pod `spec` and save the value from `--vault.ca-cert` to a file named `ca.crt`.

```console
$ kubectl get pods vault-8679f4cbf-v78cs -n demo -o jsonpath='{.spec.containers[?(@.name=="vault-unsealer")].args}'
[run --v=3 --secret-shares=4 --secret-threshold=2 --vault.ca-cert=-----BEGIN CERTIFICATE-----
MIICuDCCAaCgAwIBAgIBADANBgkqhkiG9w0BAQsFADANMQswCQYDVQQDEwJjYTAe
Fw0xOTExMDYwOTM2NDhaFw0yOTExMDMwOTM2NDhaMA0xCzAJBgNVBAMTAmNhMIIB
IjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3OfhIHN4VIidwXkf5RTMRl3J
I3+szklt6xw2ICX83OLFKk5N2DmVM1zCLcBBwE3b2PBnP3eDGEVadIHj14T+9xdc
zLjj8WbCjVR824Xn2oDLOIuwso4SFFLD1kgyfmrDw9fs0tzL8bAQqYF/75q2+Pu5
ERVscb0wXwVTE6sEqNToWqG190aUEuLbLE0n2BwqGdX1xHDhe34YgjXwvssdUJS5
tTG83iWsAJilyjFBl1Y5gP6hkgi1IB+R6HTyXY1rzKiNn3WVofp1kEeEMAJElC1Z
q4W087gYrl702MpCDh5OfVq+C4f2lc2BLh0HQ5FU1ksecFyvTo5ohdBaNzs20QID
AQABoyMwITAOBgNVHQ8BAf8EBAMCAqQwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG
9w0BAQsFAAOCAQEAn6YSx7ndvmSU+SH0bFJjnLSGMSOwWtfRAiAnJ8z+0Oea87Rr
nM+fIR4QTW8bo55Q9+fQztoWvpsb9scwfF6dg92/CsMSiOhVFvJLHHASv0Oh6vC0
dbC2N6ZGvMQb99ZPjpt5By5w7Gy5eZG2lBwitYW5M9imtxuAlkZyobrnXzNDCrYI
GDVcajcirb4qI36jjLBE9iYDiUfo3uPcgWO9XnDwRvM09lse2+VRttl7/2fqE7Vh
3mstGC4e50rgshrxvVBx6NFnTo41OpMnG7GUYCtn4/9/W5M0QDEs6rWENj6g064o
JfizhesI4ULH4XBLLJ0VN6Wp6QVJ5tEyxSA5MA==
-----END CERTIFICATE-----
 --auth.k8s-host= ... ... ...
```

Get `vault-tls` from the Kubernetes secret and write it on `tls.crt` and `tls.key` respectively:

```bash
$ kubectl get secrets -n demo vault-vault-tls -o jsonpath="{.data.tls\.crt}" | base64 -d>tls.crt

$ kubectl get secrets -n demo vault-vault-tls -o jsonpath="{.data.tls\.key}" | base64 -d>tls.key
```

List files to check:

```console
$ ls
ca.crt  tls.crt  tls.key
```

Export Vault environment variables:

```bash
$ export VAULT_ADDR=https://127.0.0.1:8200

$ export VAULT_TOKEN=$(kubectl get secrets -n demo vault-keys -o jsonpath="{.data.vault-root-token}" | base64 --decode; echo)

$ export VAULT_CACERT=ca.crt # put ca.crt file directory

$ export VAULT_CLIENT_CERT=tls.crt # put tls.crt file directory

$ export VAULT_CLIENT_KEY=tls.key # put tls.key file directory

```

Now check whether Vault server can be accessed:

```console
$ vault status
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    4
Threshold       2
Version         1.2.3
Cluster Name    vault-cluster-bb64ffd2
Cluster ID      94fcaedb-0e10-8600-21f5-97339509c60b
HA Enabled      false
```

```console
$ vault list sys/policy
Keys
----
default
k8s.-.demo.vault-auth-method-controller
root
vault-policy-controller
```

Vault CLI is ready to use. To learn more about the Vault CLI and its functionality, visit the [official documentation](https://www.vaultproject.io/docs/commands/).
