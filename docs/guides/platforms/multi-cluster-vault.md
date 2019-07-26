---
title: Use a Vault Server with Multiple Kubernetes Clusters
menu:
  docs_0.2.0:
    identifier: multi-cluster-platform
    name: Multi-Cluster
    parent: platform-guides
    weight: 100
menu_name: docs_0.2.0
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Use a Vault Server with Multiple Kubernetes Clusters

In this tutorial, we are going to show how to use Vault operator in multiple Kubernetes clusters with a single  Vault server.

To being with, we have created two GKE clusters.

![cluser image](/docs/images/guides/provider/multi-cluster/gke-cluster.png)

We are going to install Vault operator in `demo-cluster-1`. We are going to set `--cluster-name` flag. This flag value will be used by Vault operator when creating resources in Vault.

```console
$ kubectl config current-context
gke_ackube_us-central1-a_demo-cluster-1

$ curl -fsSL https://github.com/kubevault/operator/raw/0.2.0/hack/deploy/install.sh \
    | bash -s -- --cluster-name=demo-cluster-1

$ kubectl get pods -n kube-system
NAME                                                       READY   STATUS    RESTARTS   AGE
vault-operator-5fc7666575-8v6ft                            1/1     Running   0          1h
```

We are going to deploy Vault in `demo-cluster-1` using Vault operator. Guides to deploy Vault in GKE can be found [here](/docs/guides/platforms/gke.md).

```console
$ kubectl get vaultserverversions/1.0.0 -o yaml
apiVersion: catalog.kubevault.com/v1alpha1
kind: VaultServerVersion
metadata:
  name: 1.0.0
spec:
  deprecated: false
  exporter:
    image: kubevault/vault-exporter:0.1.0
  unsealer:
    image: kubevault/vault-unsealer:0.2.0
  vault:
    image: vault:1.0.0
  version: 1.0.0

$ cat examples/guides/provider/multi-cluster/my-vault.yaml
cat examples/guides/provider/multi-cluster/my-vault.yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: my-vault
  namespace: demo
spec:
  nodes: 1
  version: "1.0.0"
  backend:
    gcs:
      bucket: "demo-vault"
      credentialSecret: "google-cred"
  serviceTemplate:
    spec:
      type: LoadBalancer
      loadBalancerIP: 104.155.177.205
  tls:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9UQXhNRGN3T1RNNE1UaGFGdzB5T1RBeE1EUXdPVE00TVRoYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdXZxWFJrMGZrMHNWMFpoMDQwd0FaVTBhCkhlRW9vUnlVMlpaaGtjS3dPS201N2pUWkJaMEkvMjg2dTNpUVFpc2tMTFNjYUtvaHp0c012RXFCU0JpNU5MNEMKVXVQbm5CZklIVVo1UDhwQWNOUXJ5SURETGxXZTFBTEVKU0N0L3daRG5mMkRPdXZGSVMybFJVZDV2WFp1dlVjWgptdml5T3VUOW9CclNwNkh5YUpRYkUrZk1qQTRvZ0ZoQWZmN1djMm1DVk1jam8wU3htK2lrVWxVZWhXdWd4T3M0Cm5GUG5pWmt3a1h0KzFweU45WjltclhwUTlZM3FvdGlmdk1aUnVhVS9hbjUxOUZqSWdzVUZtRGVoZ3c3blJwYkkKZ0NGeUNPSlc5ZTNDczNRVTViVTJYUk1leDBlTDZReTZJY2dDdGZIRmpBeGhjUHAzeEJjRW5XSEdONDM2dHdJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFUMzFuQUNpMmNySW0rWCtGZUNocjJJcnhZeGFDSlpLNW92Y1Jqem5TZFR0N2JadWoKcTVZQW5jbitESDlxUURkczFVTEdjR1ZISlpiS3RORU9GVVlJbDVXYUZBVnNBMTJoaURCZnJXc24ydUV6K0pUVwovcStLSVE0OW1LUWV1TG80bkVoQnRJYjJzaXBKMmxmUEVyUXhHQllrZ3lOT05zOTN5NEdPVXU4dVdBaUFqZ21oCmM4a1QzTVV0ZVRNUHczQ3JKU2ZtbGUxQkk0bkNPNXEreW54Zk56SXZqZU5PYnNvVTdOOVhoZTdCQjZSSzQ1akgKVW41bzBTZkkrR0dYU0l5eWFxVHlJRUJkK3Nub1pMRWQzMm9DeWJFOC9QajE3T3BDZHhhd0ZnZWJGUnNEQ2pkRQo0bUI2RmVPMi9md3VFUk5lTlZnczBHMjBPc2t4cDJoTWJRYUlJZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
    tlsSecret: vault-tls
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      googleKmsGcs:
        bucket: "demo-vault"
        kmsProject: "ackube"
        kmsLocation: "global"
        kmsKeyRing: "vault"
        kmsCryptoKey: "vault-key"
        credentialSecret: "google-cred"

$ kubectl apply -f examples/guides/provider/multi-cluster/my-vault.yaml
vaultserver.kubevault.com/my-vault created

$ kubectl get vaultserver my-vault -n demo
NAME       NODES   VERSION   STATUS    AGE
my-vault   1       1.0.0     Running   1m

$ kubectl get services -n demo
NAME       TYPE           CLUSTER-IP     EXTERNAL-IP       PORT(S)                                        AGE
my-vault   LoadBalancer   10.3.251.241   104.155.177.205   8200:31542/TCP,8201:31390/TCP,9102:30911/TCP   2m
```

Now we are going to create `demo-policy-secret-admin` [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) in `demo-cluster-1`. Guides to manage policy in Vault can be found [here](/docs/guides/policy-management/overview.md).

```console
$ cat examples/guides/provider/multi-cluster/demo-policy-secret-admin.yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: demo-policy-secret-admin
  namespace: demo
spec:
  ref:
    name: my-vault
    namespace: demo
  policyDocument: |
    path "secret/*" {
      capabilities = ["create", "read", "update", "delete", "list"]
    }
$ kubectl apply -f examples/guides/provider/multi-cluster/demo-policy-secret-admin.yaml
vaultpolicy.policy.kubevault.com/demo-policy-secret-admin created

$ kubectl get vaultpolicies -n demo
NAME                              STATUS    AGE
demo-policy-secret-admin          Success   1m
```

Check the created `demo-policy-secret-admin` [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) in Vault. To resolve the naming conflict, name of policy in Vault will follow this format: `k8s.{spec.clusterName or -}.{spec.namespace}.{spec.name}`. For this case, it is `k8s.demo-cluster-1.demo.demo-policy-secret-admin`.

```console
$ export VAULT_ADDR='https://104.155.177.205:31542'

$ export VAULT_CACERT="cert/ca.crt"

$ export VAULT_TOKEN="s.KLJFDIUJLKDFDLKFJ"

$ vault policy list
default
k8s.demo-cluster-1.demo.demo-policy-secret-admin
my-vault-policy-controller
root

```

We are going to install Vault operator in `demo-cluster-2`.  We are going to set `--cluster-name`, this flag value will be used by Vault operator when creating resource in Vault.

```console
$ kubectl config current-context
gke_ackube_us-central1-a_demo-cluster-2

$ curl -fsSL https://github.com/kubevault/operator/raw/0.2.0/hack/deploy/install.sh \
    | bash -s -- --cluster-name=demo-cluster-2

$ kubectl get pods -n kube-system
NAME                                                       READY   STATUS    RESTARTS   AGE
vault-operator-5fc7666575-8v6ft                            1/1     Running   0          1h
```

Now we are going to create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains connection and credential information of the Vault that is deployed in `demo-cluster-1`. In this AppBinding, we are going to use [token auth](https://www.vaultproject.io/docs/auth/token.html#token-auth-method). Guides to Vault authentication using AppBinding can be found [here](/docs/concepts/vault-server-crds/auth-methods/overview.md).

```console
$ cat examples/guides/provider/multi-cluster/vault-app.yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  secret:
    name: vault-token
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9UQXhNRGN3T1RNNE1UaGFGdzB5T1RBeE1EUXdPVE00TVRoYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdXZxWFJrMGZrMHNWMFpoMDQwd0FaVTBhCkhlRW9vUnlVMlpaaGtjS3dPS201N2pUWkJaMEkvMjg2dTNpUVFpc2tMTFNjYUtvaHp0c012RXFCU0JpNU5MNEMKVXVQbm5CZklIVVo1UDhwQWNOUXJ5SURETGxXZTFBTEVKU0N0L3daRG5mMkRPdXZGSVMybFJVZDV2WFp1dlVjWgptdml5T3VUOW9CclNwNkh5YUpRYkUrZk1qQTRvZ0ZoQWZmN1djMm1DVk1jam8wU3htK2lrVWxVZWhXdWd4T3M0Cm5GUG5pWmt3a1h0KzFweU45WjltclhwUTlZM3FvdGlmdk1aUnVhVS9hbjUxOUZqSWdzVUZtRGVoZ3c3blJwYkkKZ0NGeUNPSlc5ZTNDczNRVTViVTJYUk1leDBlTDZReTZJY2dDdGZIRmpBeGhjUHAzeEJjRW5XSEdONDM2dHdJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFUMzFuQUNpMmNySW0rWCtGZUNocjJJcnhZeGFDSlpLNW92Y1Jqem5TZFR0N2JadWoKcTVZQW5jbitESDlxUURkczFVTEdjR1ZISlpiS3RORU9GVVlJbDVXYUZBVnNBMTJoaURCZnJXc24ydUV6K0pUVwovcStLSVE0OW1LUWV1TG80bkVoQnRJYjJzaXBKMmxmUEVyUXhHQllrZ3lOT05zOTN5NEdPVXU4dVdBaUFqZ21oCmM4a1QzTVV0ZVRNUHczQ3JKU2ZtbGUxQkk0bkNPNXEreW54Zk56SXZqZU5PYnNvVTdOOVhoZTdCQjZSSzQ1akgKVW41bzBTZkkrR0dYU0l5eWFxVHlJRUJkK3Nub1pMRWQzMm9DeWJFOC9QajE3T3BDZHhhd0ZnZWJGUnNEQ2pkRQo0bUI2RmVPMi9md3VFUk5lTlZnczBHMjBPc2t4cDJoTWJRYUlJZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
    url: https://104.155.177.205:8200

$ kubectl apply -f examples/guides/provider/multi-cluster/vault-app.yaml
appbinding.appcatalog.appscode.com/vault-app created
```

Now we are going to create `demo-policy-secret-reader` [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) in `demo-cluster-2`.

```console
$ cat examples/guides/provider/multi-cluster/demo-policy-secret-reader.yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: demo-policy-secret-reader
  namespace: demo
spec:
  ref:
    name: vault-app
    namespace: demo
  policyDocument: |
    path "secret/*" {
      capabilities = ["read", "list"]
    }

$ kubectl apply -f examples/guides/provider/multi-cluster/demo-policy-secret-reader.yaml
vaultpolicy.policy.kubevault.com/demo-policy-secret-reader created

$ kubectl get vaultpolicies -n demo
NAME                        STATUS    AGE
demo-policy-secret-reader   Success   1m
```

Check the created `demo-policy-secret-reader` [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) in Vault. To resolve the naming conflict, name of policy in Vault will follow this format: `k8s.{spec.clusterName or -}.{spec.namespace}.{spec.name}`. For this case, it is `k8s.demo-cluster-2.demo.demo-policy-secret-reader`.

```console
$ vault policy list
default
k8s.demo-cluster-1.demo.demo-policy-secret-admin
k8s.demo-cluster-2.demo.demo-policy-secret-reader
my-vault-policy-controller
root

```

This how we can use Vault operator in multiple Kubernetes clusters with a single Vault Server without naming conflict.
