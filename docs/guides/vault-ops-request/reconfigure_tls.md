---
title: Vault Ops Request Reconfigure TLS
menu:
  docs_{{ .version }}:
    identifier: reconfigure-tls-ops-request-guides
    name: Reconfigure TLS
    parent: ops-request-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/README.md).

# Reconfigure VaultServer TLS/SSL

`KubeVault` supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing `VaultServer` via a `VaultOpsRequest`. This tutorial will show you how to use `KubeVault` to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeVault cli on your workstation and KubeVault operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/guides/vault-ops-request](https://github.com/kubevault/kubevault/tree/{{< param "info.version" >}}/docs/examples/guides/vault-ops-request) folder in GitHub repository [kubevault/kubevault](https://github.com/kubevault/kubevault).

## Add TLS to a VaultServer

Here, We are going to create a `VaultServer` without TLS and then reconfigure the `VaultServer` to use TLS.

### Deploy VaultServer without TLS

In this section, we are going to deploy a VaultServer without TLS. In the next few sections we will reconfigure TLS using `VaultOpsRequest` CRD. Below is the YAML of the `VaultServer` CR that we are going to create,

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: 1.10.3
  replicas: 3
  allowedSecretEngines:
    namespaces:
      from: All
    secretEngines:
      - gcp
  backend:
    raft:
      storage:
        storageClassName: "standard"
        resources:
          requests:
            storage: 1Gi
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  monitor:
    agent: prometheus.io
    prometheus:
      exporter:
        resources: {}
  terminationPolicy: WipeOut

```

Let's create the `VaultServer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/vault-ops-request/vaultserver.yaml
vaultserver.kubevault.com/vault created
```


Now, wait until `VaultServer` has status `Ready`. i.e,

```bash
$ kubectl get vs -n demo
NAME    REPLICAS   VERSION   STATUS   AGE
vault   3          1.12.1    Ready    128m
```

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in VaultServer. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=vault/O=kubevault"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls vault-ca --cert=ca.crt --key=ca.key --namespace=demo

secret/vault-ca created
```

Now, Let's create an `Issuer` using the `vault-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: issuer
  namespace: demo
spec:
  ca:
    secretName: vault-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/vault-ops-request/issuer.yaml
issuer.cert-manager.io/issuer created
```

### Create VaultOpsRequest

In order to add TLS to the VaultServer, we have to create a `VaultOpsRequest` CRO with our created issuer. Below is the YAML of the `VaultOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubevault.com/v1alpha1
kind: VaultOpsRequest
metadata:
  name: vault-ops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  vaultRef:
    name: vault
  tls:
    issuerRef:
      name: issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - appscode
          organizationalUnits:
            - client
```

Here,

- `spec.vaultRef.name` specifies that we are performing reconfigure TLS operation on `vault` VaultServer.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our VaultServer.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates.

Let's create the `VaultOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/vault-ops-request/vault-ops-add-tls.yaml
vaultopsrequest.ops.kubevault.com/vault-ops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `VaultOpsRequest` to be `Successful`.  Run the following command to watch `VaultOpsRequest` CRO,

```bash
$ kubectl get vaultopsrequest -n demo
Every 2.0s: kubectl get vaultopsrequest -n demo
NAME               TYPE             STATUS        AGE
vault-ops-add-tls  ReconfigureTLS   Successful    91s
```

## Rotate Certificate

Now we are going to rotate the certificate of this VaultServer. First let's check the current expiration date of the certificate.

```bash
$ kubectl exec -it -n demo vault-0 -- bin/sh
/ # cd etc/vault/tls/server
/etc/vault/tls/server # cat tls.crt
-----BEGIN CERTIFICATE-----
MIID2DCCAsCgAwIBAgIQL1rqn4OHpvchiFRI3DPXIjANBgkqhkiG9w0BAQsFADAk
...
XJRRwl5psqcyp5ZJI1ar5JP1JCGQa3QTArwstw==
-----END CERTIFICATE-----
```

Copy & paste the certificate in any certificates decoding tool like [certlogic](https://certlogik.com/decoder/) & check it's expiry date.

### Create VaultOpsRequest

Now we are going to increase it using a VaultOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubevault.com/v1alpha1
kind: VaultOpsRequest
metadata:
  name: vault-ops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  vaultRef:
    name: vault
  tls:
    rotateCertificates: true
```

Here,

- `spec.vaultRef.name` specifies that we are performing reconfigure TLS operation on `vault` VaultServer.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our VaultServer.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this VaultServer.

Let's create the `VaultOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/vault-ops-request/vault-ops-rotate.yaml
vaultopsrequest.ops.kubevault.com/vault-ops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `VaultOpsRequest` to be `Successful`.  Run the following command to watch `VaultOpsRequest` CRO,

```bash
$ kubectl get vaultopsrequest -n demo
Every 2.0s: kubectl get vaultopsrequest -n demo
NAME                TYPE             STATUS        AGE
vault-ops-rotate    ReconfigureTLS   Successful    112
```

Now, let's check the expiration date of the certificate again, it should be updated.

```bash
$ kubectl exec -it -n demo vault-0 -- bin/sh
/ # cd etc/vault/tls/server
/etc/vault/tls/server # cat tls.crt
-----BEGIN CERTIFICATE-----
MIID2DCCAsCgAwIBAgIQL1rqn4OHpvchiFRI3DPXIjANBgkqhkiG9w0BAQsFADAk
...
XJRRwl5psqcyp5ZJI1ar5JP1JCGQa3QTArwstw==
-----END CERTIFICATE-----
```

## Change Issuer/ClusterIssuer

Now, we are going to change the issuer of this VaultServer.

- Let's create a new ca certificate and key using a different subject `CN=ca-updated,O=kubevault-updated`.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca-updated/O=kubevault-updated"
Generating a RSA private key
..............................................................+++++
......................................................................................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a new ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls vault-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/vault-new-ca created
```

Now, Let's create a new `Issuer` using the `vault-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: new-issuer
  namespace: demo
spec:
  ca:
    secretName: vault-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/vault-ops-request/new-issuer.yaml
issuer.cert-manager.io/new-issuer created
```


### Create VaultOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `VaultOpsRequest` CRO with the newly created issuer. Below is the YAML of the `VaultOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubevault.com/v1alpha1
kind: VaultOpsRequest
metadata:
  name: vault-ops-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  vaultRef:
    name: vault
  tls:
    issuerRef:
      name: new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.vaultRef.name` specifies that we are performing reconfigure TLS operation on `vault` VaultServer.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our VaultServer.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `VaultOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/vault-ops-request/vault-ops-change-issuer.yaml
vaultopsrequest.ops.kubevault.com/vault-ops-change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `VaultOpsRequest` to be `Successful`.  Run the following command to watch `VaultOpsRequest` CRO,

```bash
$ kubectl get vaultopsrequest -n demo
Every 2.0s: kubectl get vaultopsrequest -n demo
NAME                       TYPE             STATUS        AGE
vault-ops-change-issuer    ReconfigureTLS   Successful    105s
```

## Remove TLS from the VaultServer

Now, we are going to remove TLS from this VaultServer using a VaultOpsRequest.

### Create VaultOpsRequest

Below is the YAML of the `VaultOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubevault.com/v1alpha1
kind: VaultOpsRequest
metadata:
  name: vault-ops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  vaultRef:
    name: vault
  tls:
    remove: true
```

Here,

- `spec.vaultRef.name` specifies that we are performing reconfigure TLS operation on `vault` VaultServer.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our VaultServer.
- `spec.tls.remove` specifies that we want to remove tls from this VaultServer.

Let's create the `VaultOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/vault-ops-request/vault-ops-remove.yaml
vaultopsrequest.ops.kubeavult.com/vault-ops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `VaultOpsRequest` to be `Successful`.  Run the following command to watch `VaultOpsRequest` CRO,

```bash
$ kubectl get vaultopsrequest -n demo
Every 2.0s: kubectl get vaultopsrequest -n demo
NAME               TYPE             STATUS        AGE
vault-ops-remove   ReconfigureTLS   Successful    105s
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete vaultserver -n demo vault
kubectl delete issuer -n demo issuer new-issuer
kubectl delete vaultopsrequest vault-ops-add-tls vault-ops-remove vault-ops-rotate vault-ops-change-issuer
kubectl delete ns demo
```
