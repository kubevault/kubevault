---
title: Configure TLS/SSL for VaultServer | KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: vault-tls-concepts
    name: TLS/SSL Configuration
    parent: vault-server-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Configure TLS/SSL for VaultServer

`KubeVault` provides support for TLS/SSL for `VaultServer`. This tutorial will show you how to use `KubeVault` to deploy a `VaultServer` with TLS/SSL configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.4.0 or later to your cluster to manage your SSL/TLS certificates.

- Install `KubeVault` operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

### Deploy VaultServer with TLS/SSL configuration

As pre-requisite, at first, we are going to create an Issuer/ClusterIssuer. This Issuer/ClusterIssuer is used to create certificates. Then we are going to deploy a VaultServer with TLS/SSL configuration.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=vault/O=kubevault"
```

- create a secret using the certificate files we have just generated,

```bash
$ kubectl create secret tls vault-ca --cert=ca.crt  --key=ca.key --namespace=demo 
secret/vault-ca created
```

Now, we are going to create an `Issuer` using the `vault-ca` secret that contains the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: vault-issuer
 namespace: demo
spec:
 ca:
   secretName: vault-ca
```

Let’s create the `Issuer` cr we have shown above,

```bash
kubectl apply -f issuer.yaml
issuer.cert-manager.io/vault-issuer created
```

### Deploy VaultServer with TLS/SSL configuration

Here, our issuer `vault-issuer`  is ready to deploy a `VaultServer` Cluster with TLS/SSL configuration. Below is the YAML for VaultServer that we are going to create,

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: vault-issuer
  allowedSecretEngines:
    namespaces:
      from: All
    secretEngines:
      - mysql
  version: 1.8.2
  replicas: 3
  backend:
    raft:
      path: "/vault/data"
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
  terminationPolicy: DoNotTerminate
```

Here,

- `spec.tls.issuerRef` refers to the `vault-issuer` issuer.
