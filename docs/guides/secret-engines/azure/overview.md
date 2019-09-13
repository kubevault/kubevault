---
title: Manage Azure service principals using the Vault Operator
menu:
  docs_{{ .version }}:
    identifier: overview-azure
    name: Overview
    parent: azure-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Azure service principals using the Vault Operator

You can easily manage [Azure secret engine](https://www.vaultproject.io/docs/secrets/azure/index.html) using Vault operator.

You need to be familiar with the following CRDs:

- [AzureRole](/docs/concepts/secret-engine-crds/azurerole.md)
- [AzureAccessKeyRequest](/docs/concepts/secret-engine-crds/azureaccesskeyrequest)
- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)

Before you begin:

- Install Vault operator in your cluster following the steps [here](/docs/setup/operator/install).

- Deploy Vault. It could be in the Kubernetes cluster or external.

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we are going to create a [role](https://www.vaultproject.io/api/secret/azure/index.html#create-update-role) using AzureRole and issue credential using AzureAccessKeyRequest. For this tutorial, we are going to deploy Vault using Vault operator.

```console
$ cat examples/guides/secret-engins/azure/vaultseverInmem.yaml

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
$ kubectl apply -f examples/guides/secret-engins/azure/vaultseverInmem.yaml
  vaultserver.kubevault.com/vault created

$ kubectl get vaultserver/vault -n demo
  NAME    NODES   VERSION   STATUS    AGE
  vault   1       1.0.1     Running   15m
```

## AzureRole

Using [AzureRole](/docs/concepts/secret-engine-crds/azurerole.md), you can [configure azure secret backend](https://www.vaultproject.io/docs/secrets/azure/index.html#setup) and [create azure role](https://www.vaultproject.io/api/secret/azure/index.html#create-update-role).

```console
$ cat examples/guides/secret-engins/azure/azureRole.yaml

apiVersion: engine.kubevault.com/v1alpha1
kind: AzureRole
metadata:
  name: demo-role
  namespace: demo
spec:
  ref:
    name: vault-app
    namespace: demo
  applicationObjectID: c1cb042d-96d7-423a-8dba-243c2e5010d3
  config:
    credentialSecret: azure-cred
  ttl: 1h
  maxTTL: 1h
```

Before deploying a AzureRole crd, you need to make sure that `spec.ref` and `config.clientSecret` fields are valid.

`spec.ref` contains [appbinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) reference. You can use any valid [auth method](/docs/concepts/vault-server-crds/auth-methods/overview.md) while creating appbinding. We will use [token auth method](/docs/concepts/vault-server-crds/auth-methods/token.md) in this tutorial.

```console
$ cat examples/guides/secret-engins/azure/token.yaml

apiVersion: v1
data:
  token: cy4xUnpySndvakZ6WjlYRU9vNjVaWmR6Q2Y=
kind: Secret
metadata:
  name: vault-token
  namespace: demo
type: kubevault.com/token

$ kubectl apply -f examples/guides/secret-engins/azure/token.yaml
  secret/vault-token created
```

```console
$ cat examples/guides/secret-engins/azure/tokenAppbinding.yaml
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
      caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9UQTFNVFV4TURFMU5EUmFGdzB5T1RBMU1USXhNREUxTkRSYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBc0E0VzVBVTZaZ29nKzhISjFKcTYvV3MrCkI1VDdDMDdOeGNpbEZwdkJleEVoMnZ2ZFpmQkp3NEw3QmNkaExQYzc1OGVOMUxnaEpaSlRQRjdvRXJRbzFZM0EKaUp0YUtURHlVU1ZZSXdvOVZhSG9zWkMyWFdRTEFWZ0NOVmlVbjN3Y0pnT043cjJUVSs2dWNxY2RidUVsYWJkWgpVbzBRNHRSUDgvaWEwcElCWnV2a0ZqY2R3QUpEMzFKaGE5WGdibFBVSnROdW9pYVcvSytXWjM1TU5iK2JGQ2tRCk14Zm1aMTFNb1dsZzlUYjBRQ1ZoRzBPSWdHVy9ySU1LTHphWS85MXJSNmFFU25JRWdyUjZTczhPZ0NKVEVLUzMKaS8zd2laNVVPVlI5MVFPSkRMSkROeXJCc1lRcENGVjJ2VHcydVpoYWZUWUZ4UWZrNTlvU0w5aGpoMFZyQ1FJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFyelJ1TnpUc2tpd0JPZzBFdHcrc2Y4UU00ZG01eklHS2huQ3pkVUw2Z2w3azhhUVgKUlkxM1RrME9EZnBlbEM1KzVIZTYrK1U1aUo2amFYcEIyUEZhQWtwRVJ3a0h3Qm1lOHN0b01rVHRWd1hhUVVBcwpyT1A2Y2REOVJJMlBjbFZXaG5NdVdSVnJKZStYVG1lakVjRW53cWJIT0hMTlhSYkNpNW1XNVU4a0xVc2JISmZRClB0T3RRdlYySStFNmJ4MjdaNjduQXg3QnVCWlpKbm1SQUJoc1lQQzllbFdmdlFoOVVFcVk0QWUvaTh2MEdTamcKTjRVbVNUMXFDaGQ5dG9IZVlORURMM2hDMkpORExJNWhSMy80eTBDRUcvclhsQ1ZwSmtpTlJhSm9CVmErckhNdQorM0IwZW85MGo0UUhBbVpQWTkxVmdZc09ZaGdOc2s3ZUtFdmxYZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K

$ kubectl apply -f examples/guides/secret-engins/azure/tokenAppbinding.yaml
  appbinding.appcatalog.appscode.com/vault-app created
```

`spec.config.credentialSecret` contains the name of the Kubernetes secret that holds the credential for Azure Active Directory(AAD)

```console
$ cat examples/guides/secret-engins/azure/azure-secret.yaml
  apiVersion: v1
  kind: Secret
  metadata:
    name: azure-cred
    namespace: demo
  data:
    subscription-id: MWJmYzlmNjYtMzE2ZC0****
    tenant-id: NzcyMjY4ZTUtZDk0MC00Ym*******
    client-id: MmI4NzFkNGEtNzU3ZS00YjJ******
    client-secret: TU1hRjdRZWVzTGZxbGRpVD***

$ kubectl apply -f examples/guides/secret-engins/azure/azure-secret.yaml
  secret/azure-client-secret created
```

Now we can deploy our AzureRole

```console
$ kubectl apply -f examples/guides/secret-engins/azure/azureRole.yaml
  azurerole.engine.kubevault.com/demo-role created

$ kubectl get azureroles -n demo
  NAME        STATUS
  demo-role   Success
```

To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{spec.clusterName or -}.{spec.namespace}.{spec.name}`.

```console
$ vault list azure/roles
  Keys
  ----
  k8s.-.demo.demo-role

$ vault read azure/roles/k8s.-.demo.demo-role
  Key                      Value
  ---                      -----
  application_object_id    c1cb042d-96d7-423a-8dba-...
  azure_roles              <nil>
  max_ttl                  1h
  ttl                      1h
```

If we delete AzureRole, then respective role will be deleted from Vault.

```console
$ kubectl delete -f examples/guides/secret-engins/azure/azureRole.yaml
  azurerole.engine.kubevault.com "demo-role" deleted

# check in vault whether role exists
$ vault read azure/roles/k8s.-.demo.demo-role
  No value found at azure/roles/k8s.-.demo.demo-role

$ vault list azure/roles
  No value found at azure/roles/
```

## AzureAccessKeyRequest

Using [AzureAccessKeyRequest](/docs/concepts/secret-engine-crds/azureaccesskeyrequest), you can generate azure service principals from Vault. In this tutorial, we are going to azure service principals by creating `azure-credential` AzureAccessKeyRequest in `demo` namespace.

```console
$ cat examples/guides/secret-engins/azure/azureAKR.yaml
  apiVersion: engine.kubevault.com/v1alpha1
  kind: AzureAccessKeyRequest
  metadata:
    name: azure-credential
    namespace: demo
  spec:
    roleRef:
      name: demo-role
      namespace: demo
    subjects:
    - kind: ServiceAccount
      name: sa
      namespace: demo
```

Here, `spec.roleRef` is the reference of AzureRole against which credentials will be issued. `spec.subjects` is the reference to the object or user identities a role binding applies to and it will have read access of the credential secret. Also, Vault operator will use AppBinding reference from AzureRole which is specified in `spec.roleRef`.

Now, we are going to create `azure-credential` AzureAccessKeyRequest.

```console
$ kubectl apply -f examples/guides/secret-engins/azure/azureAKR.yaml
  azureaccesskeyrequest.engine.kubevault.com/azure-credential created

$ kubectl get azureaccesskeyrequest -n demo
  NAME               AGE
  azure-credential   31s
```

Azure credential will not be issued until it is approved. To approve it, you have to add `Approved` in `status.conditions[].type` field. You can use [KubeVault CLI](https://github.com/kubevault/cli) as [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) to approve or deny AzureAccessKeyRequest.

```console
# using KubeVault cli as kubectl plugin to approve request
$ kubectl vault approve azureaccesskeyrequest azure-credential -n demo
  approved

$ kubectl get azureaccesskeyrequest -n demo azure-credential -o yaml
  apiVersion: engine.kubevault.com/v1alpha1
  kind: AzureAccessKeyRequest
  metadata:
    name: azure-credential
    namespace: demo
  spec:
    roleRef:
      name: demo-role
      namespace: demo
    subjects:
    - kind: ServiceAccount
      name: sa
      namespace: demo
  status:
    conditions:
    - lastUpdateTime: 2019-05-15T10:40:55Z
      message: This was approved by kubectl vault approve azureaccesskeyrequest
      reason: KubectlApprove
      type: Approved
    lease:
      duration: 1h0m0s
      id: azure/creds/k8s.-.demo.demo-role/qelACTqQw7ELFcTItfWNz4Aq
      renewable: true
    secret:
      name: azure-credential-z3qwsi
```

Once AzureAccessKeyRequest is approved, Vault operator will issue credential from Vault and create a secret containing the credential. Also it will create rbac role and rolebinding so that `spec.subjects` can access secret. You can view the information in `status` field.

```console
$ kubectl get azureaccesskeyrequest/azure-credential -n demo -o json | jq '.status'
  {
    "conditions": [
      {
        "lastUpdateTime": "2019-05-15T10:40:55Z",
        "message": "This was approved by kubectl vault approve azureaccesskeyrequest",
        "reason": "KubectlApprove",
        "type": "Approved"
      }
    ],
    "lease": {
      "duration": "1h0m0s",
      "id": "azure/creds/k8s.-.demo.demo-role/qelACTqQw7ELFcTItfWNz4Aq",
      "renewable": true
    },
    "secret": {
      "name": "azure-credential-z3qwsi"
    }
  }

$ kubectl get secret -n demo azure-credential-z3qwsi -o yaml
  apiVersion: v1
  data:
    client_id: MmI4NzFkNGEtNzU3ZS00Y.....
    client_secret: N2VjODRhZmUtM2YzMS.....
  kind: Secret
  metadata:
    creationTimestamp: 2019-05-15T10:40:57Z
    name: azure-credential-z3qwsi
    namespace: demo
    ownerReferences:
    - apiVersion: engine.kubevault.com/v1alpha1
      controller: true
      kind: AzureAccessKeyRequest
      name: azure-credential
      uid: ab115819-76fd-11e9-8494-08002770ee4f
    resourceVersion: "6650"
    selfLink: /api/v1/namespaces/demo/secrets/azure-credential-z3qwsi
    uid: ec4fa1a2-76fd-11e9-8494-08002770ee4f
  type: Opaque
```

If AzureAccessKeyRequest is deleted, then credential lease (if have any) will be revoked.

```console
$ kubectl delete azureaccesskeyrequest -n demo azure-credential
  azureaccesskeyrequest.engine.kubevault.com "azure-credential" deleted
```

If AzureAccessKeyRequest is `Denied`, then Vault operator will not issue any credential.

```console
$ kubectl vault deny azureaccesskeyrequest azure-credential -n demo
  Denied

$ kubectl get azureaccesskeyrequest/azure-credential -n demo -o json | jq '.status'
  {
    "conditions": [
      {
        "lastUpdateTime": "2019-05-15T11:10:39Z",
        "message": "This was denied by kubectl vault deny azureaccesskeyrequest",
        "reason": "KubectlDeny",
        "type": "Denied"
      }
    ]
  }
```

> Note: Once AzureAccessKeyRequest is `Approved` or `Denied`, you cannot change `spec.roleRef` and `spec.subjects` field.
