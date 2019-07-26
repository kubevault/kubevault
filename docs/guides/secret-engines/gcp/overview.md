---
title: Manage GCP IAM Secrets using the Vault Operator
menu:
  docs_0.2.0:
    identifier: overview-gcp
    name: Overview
    parent: gcp-secret-engines
    weight: 10
menu_name: docs_0.2.0
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage GCP IAM Secrets using the Vault Operator

You can easily manage [GCP secret engine](https://www.vaultproject.io/docs/secrets/gcp/index.html) using Vault operator.

You need to be familiar with the following CRDs:

- [GCPRole](/docs/concepts/secret-engine-crds/gcprole.md)
- [GCPAccessKeyRequest](/docs/concepts/secret-engine-crds/gcpaccesskeyrequest)
- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)

Before you begin:

- Install Vault operator in your cluster following the steps [here](/docs/setup/operator/install).

- Deploy Vault. It could be in the Kubernetes cluster or external.

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we are going to create a [role](https://www.vaultproject.io/api/secret/gcp/index.html#create-update-roleset) using GCPRole and issue credential using GCPAccessKeyRequest. For this tutorial, we are going to deploy Vault using Vault operator.

```console
$ cat examples/guides/secret-engins/gcp/vaultseverInmem.yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  nodes: 1
  version: "1.0.1"
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

```console
$ kubectl get vaultserverversions/1.0.1 -o yaml
apiVersion: catalog.kubevault.com/v1alpha1
kind: VaultServerVersion
metadata:
  labels:
    app: vault-operator
  name: 1.0.1
spec:
  deprecated: false
  exporter:
    image: kubevault/vault-exporter:0.1.0
  unsealer:
    image: kubevault/vault-unsealer:0.2.0
  vault:
    image: vault:1.0.1
  version: 1.0.1
```

```console
$ kubectl apply -f examples/guides/secret-engins/gcp/vaultseverInmem.yaml
  vaultserver.kubevault.com/vault created

$ kubectl get vaultserver/vault -n demo
  NAME    NODES   VERSION   STATUS    AGE
  vault   1       1.0.1     Running   15m
```

## GCPRole

Using [GCPRole](/docs/concepts/secret-engine-crds/gcprole.md), you can [configure gcp secret backend](https://www.vaultproject.io/docs/secrets/gcp/index.html#setup) and [create gcp roleset](https://www.vaultproject.io/api/secret/gcp/index.html#create-update-roleset).

```console
$ cat examples/guides/secret-engins/gcp/gcpRole.yaml
  apiVersion: engine.kubevault.com/v1alpha1
  kind: GCPRole
  metadata:
    name: gcp-role
    namespace: demo
  spec:
    ref:
      name: vault-app
      namespace: demo
    config:
      credentialSecret: gcp-cred
    secretType: access_token
    project: ackube
    bindings: 'resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
          roles = ["roles/viewer"]
        }'
    tokenScopes: ["https://www.googleapis.com/auth/cloud-platform"]
```

Before deploying a GCPRole crd, you need to make sure that `spec.ref` and `config.credentialSecret` fields are valid.

`spec.ref` contains [appbinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) reference. You can use any valid [auth method](/docs/concepts/vault-server-crds/auth-methods/overview.md) while creating appbinding. We will use [token auth method](/docs/concepts/vault-server-crds/auth-methods/token.md) in this tutorial.

```console
$ cat examples/guides/secret-engins/gcp/token.yaml
  apiVersion: v1
  data:
    token: cy40cFptR3hnWTZHb2FIVEVac3ZQZVZhaG4=
  kind: Secret
  metadata:
    name: vault-token
    namespace: demo
  type: kubevault.com/token

$ kubectl apply -f examples/guides/secret-engins/gcp/token.yaml
  secret/vault-token created
```

```console
$ cat examples/guides/secret-engins/gcp/tokenAppbinding.yaml
  apiVersion: appcatalog.appscode.com/v1alpha1
  kind: AppBinding
  metadata:
    name: vault-app
    namespace: demo
  spec:
    secret:
      name: vault-token
    clientConfig:
      service:
        name: vault
        scheme: HTTPS
        port: 8200
      caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9UQTBNak13TkRNNE16bGFGdzB5T1RBME1qQXdORE00TXpsYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBd2VBZ0xkNHpmcmEzZCs5a3RJNW1sNWZpCko4SFc5WGUwZXlyOXVLNW4walNNaXFEcWJ0QUJnNUJKUHUzYXBNSlZxRFc3TXE0WHVxcURMeFZtQzVuQXNRcjcKbEFudXdQVHEvZXJqdWREUXk3Rm5UVEFyaVQ4c0JXK28vRW55c25QUk4yNko3WkZWaUdXdk1QUTZKNnpJaGkwbwpSaUVPZysyQm1nWS9TbHNOejVETWNzb25vbHJ5eitCeE1XQjFUbmI1QjcvWVg5UXNHanN5clNITXIwOVlyMDNiCkVWclpNZnNLaUxUWDduV2ZXSTZjYWVyb0lGenpQNDRkWm9nYnc2UUMwQXFJN1pHd2MrYXJIQm9IZUMwYXZDRGkKNkZHTGpKMjlTd0kzYll3cWhDNnNSZlpXU2IxWDc5UVJQS1A0OWNXcStrbU5Vc2k5WEFtNUNnWTd3QlovVlFJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFNMU03dVJxTGg2cHBiSmRFZ1JZS3JTeWxkSE00UWdrQjdRMy94ZE5ZdDEyclp4NG8KdFVjOXBDQ2VlVkYwZ01GZjViZ0VUTmU4QTFuL1ZnRGx5aUZTdEtrc0c5VHpxRmIrbzBsQTJUSi9aeTQxdnluegptOTZBWCtROUNuRnJteW1YTWc5TU1PbEtHTUFQUW41R2xpbzA0Yi9ESFczcmlsNEVVVDNNSm56TndwZ3VWRmtzCnhSL3VFY2IxZ2ZtejQxcGhUcm9CcWhselMwWmVXbzR4eElSTG5qTFN3Y0V4S1NvK0xyU2RheTdyK1ZpelA3NDUKUXB4eEF2aUtPN1BBWnlDL2FrT2Z5NkVxYm5OK1ZJa3dnNTQ3NUYvYi9YNHRqYTQwaWY5TXRNUmFPKzlWNm8vNgpoNnJEeDQ1QUdLckNvS25IMzk5R05nL2Naclh2YWR1NDRjOHRoZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K

$ kubectl apply -f examples/guides/secret-engins/gcp/tokenAppbinding.yaml
  appbinding.appcatalog.appscode.com/vault-app created
```

`spec.config.credentialSecret` contains the name of the Kubernetes secret that holds Google application credentials

```console
$ cat examples/guides/secret-engins/gcp/gcp_cred.yaml
  apiVersion: v1
  kind: Secret
  metadata:
    name: gcp-cred
    namespace: demo
    annotations:
      kubevault.com/auth-path: gcp
  data:
    sa.json: ewogICJ0eXBzMm1oZHVyLzgra2FJWW1xZFVsWWxmdWVKS......
  type: kubernetes.io/gcp


$ kubectl apply -f examples/guides/secret-engins/gcp/gcp_cred.yaml
  secret/gcp-cred created
```

Now we can deploy our GCPRole CRD.

```console
$ kubectl apply -f examples/guides/secret-engins/gcp/gcpRole.yaml
gcprole.engine.kubevault.com/gcp-role created

$ kubectl get gcprole -n demo
NAME       STATUS
gcp-role   Success
```

To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{spec.clusterName or -}.{spec.namespace}.{spec.name}`.

```console
$ vault list gcp/roleset
  Keys
  ----
  k8s.-.demo.gcp-role

$ vault read gcp/roleset/k8s.-.demo.gcp-role
  Key                        Value
  ---                        -----
  bindings                   map[//cloudresourcemanager.googleapis.com/projects/ackube:[roles/viewer]]
  secret_type                access_token
  service_account_email      vaultk8s---demo-gcp-1555998460@ackube.iam.gserviceaccount.com
  service_account_project    ackube
  token_scopes               [https://www.googleapis.com/auth/cloud-platform]
```

If we delete GCPRole, then respective role will be deleted from Vault.

```console
$ kubectl delete -f examples/guides/secret-engins/gcp/gcpRole.yaml
  gcprole.engine.kubevault.com "gcp-role" deleted

# check in vault whether role exists
$ vault read gcp/roleset/k8s.-.demo.gcp-role
  No value found at gcp/roleset/k8s.-.demo.gcp-role

$ vault list gcp/roleset
  No value found at gcp/roleset/
```

## GCPAccessKeyRequest

Using [GCPAccessKeyRequest](/docs/concepts/secret-engine-crds/gcpaccesskeyrequest), you can issue GCP credential from Vault. In this tutorial, we are going to issue GCP credential by creating `gcp-credential` GCPAccessKeyRequest in `demo` namespace.

```console
$ cat examples/guides/secret-engins/gcp/gcpAKR.yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: GCPAccessKeyRequest
metadata:
  name: gcp-credential
  namespace: demo
spec:
  roleRef:
    name: gcp-role
    namespace: demo
  subjects:
  - kind: ServiceAccount
    name: sa
    namespace: demo
  secretType: access_token
```

Here, `spec.roleRef` is the reference of GCPRole against which credential will be issued. `spec.subjects` is the reference to the object or user identities a role binding applies to and it will have read access of the credential secret. Also, Vault operator will use AppBinding reference from GCPRole which is specified in `spec.roleRef`.

Now, we are going to create `gcp-credential` GCPAccessKeyRequest.

```console
$ kubectl apply -f examples/guides/secret-engins/gcp/gcpAKR.yaml
gcpaccesskeyrequest.engine.kubevault.com/gcp-credential created

$ kubectl get gcpaccesskeyrequests -n demo
NAME        AGE
gcp-credential   3s
```

GCP credential will not be issued until it is approved. To approve it, you have to add `Approved` in `status.conditions[].type` field. You can use [KubeVault CLI](https://github.com/kubevault/cli) as [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) to approve or deny GCPAccessKeyRequest.

```console
# using KubeVault cli as kubectl plugin to approve request
$ kubectl vault approve gcpaccesskeyrequest gcp-credential -n demo
  approved

$ kubectl get gcpaccesskeyrequests -n demo gcp-credential -o yaml
  apiVersion: engine.kubevault.com/v1alpha1
  kind: GCPAccessKeyRequest
  metadata:
    creationTimestamp: 2019-04-23T10:07:22Z
    finalizers:
    - gcpaccesskeyrequest.engine.kubevault.com
    generation: 1
    name: gcp-credential
    namespace: demo
    resourceVersion: "7088"
    selfLink: /apis/engine.kubevault.com/v1alpha1/namespaces/demo/gcpaccesskeyrequests/gcp-credential
    uid: 964b46be-65af-11e9-97c0-08002778f766
  spec:
    roleRef:
      name: gcp-role
      namespace: demo
    secretType: access_token
    subjects:
    - kind: ServiceAccount
      name: sa
      namespace: demo
  status:
    conditions:
    - lastUpdateTime: 2019-04-23T10:16:19Z
      message: This was approved by kubectl vault approve gcpaccesskeyrequest
      reason: KubectlApprove
      type: Approved
    lease:
      duration: 0s
    secret:
      name: gcp-credential-ewhtzp
```

Once GCPAccessKeyRequest is approved, Vault operator will issue credential from Vault and create a secret containing the credential. Also it will create rbac role and rolebinding so that `spec.subjects` can access secret. You can view the information in `status` field.

```console
$ kubectl get gcpaccesskeyrequest/gcp-credential -n demo -o json | jq '.status'
  {
    "conditions": [
      {
        "lastUpdateTime": "2019-04-23T11:21:54Z",
        "message": "This was approved by kubectl vault approve gcpaccesskeyrequest",
        "reason": "KubectlApprove",
        "type": "Approved"
      }
    ],
    "lease": {
      "duration": "0s"
    },
    "secret": {
      "name": "gcp-credential-kn5lbq"
    }


$ kubectl get secret -n demo gcp-credential-kn5lbq -o yaml
  apiVersion: v1
  data:
    expires_at_seconds: MTU1NjAyMjExOA==
    token: eWEyOS5jLkVsbjBCci1lWEN==
    token_ttl: MzU5OQ==
  kind: Secret
  metadata:
    creationTimestamp: 2019-04-23T11:21:58Z
    name: gcp-credential-kn5lbq
    namespace: demo
    ownerReferences:
    - apiVersion: engine.kubevault.com/v1alpha1
      controller: true
      kind: GCPAccessKeyRequest
      name: gcp-credential
      uid: e36945e5-65b9-11e9-97c0-08002778f766
    resourceVersion: "12409"
    selfLink: /api/v1/namespaces/demo/secrets/gcp-credential-kn5lbq
    uid: 020f7cc6-65ba-11e9-97c0-08002778f766
  type: Opaque
```

If GCPAccessKeyRequest is deleted, then credential lease (if have any) will be revoked.

```console
$ kubectl delete gcpaccesskeyrequest -n demo gcp-credential
  gcpaccesskeyrequest.engine.kubevault.com "gcp-credential" deleted
```

If GCPAccessKeyRequest is `Denied`, then Vault operator will not issue any credential.

```console
$ kubectl vault deny gcpaccesskeyrequest gcp-credential -n demo
  Denied
```

> Note: Once GCPAccessKeyRequest is `Approved` or `Denied`, you can not change `spec.roleRef` and `spec.subjects` field.
