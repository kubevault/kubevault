---
title: Manage Azure service principals using the KubeVault operator
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

# Manage Azure service principals using the KubeVault operator

The Azure secrets engine dynamically generates Azure service principals and role assignments. Vault roles can be mapped to one or more Azure roles, providing a simple, flexible way to manage the permissions granted to generated service principals. You can easily manage the [Azure secret engine](https://www.vaultproject.io/docs/secrets/azure/index.html) using the KubeVault operator.

![Azure secret engine](/docs/images/guides/secret-engines/azure/azure_secret_engine_guide.svg)

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [AzureRole](/docs/concepts/secret-engine-crds/azure-secret-engine/azurerole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we are going to create a [role](https://www.vaultproject.io/api/secret/azure/index.html#create-update-role) using AzureRole and issue credential using SecretAccessRequest.

## Vault Server

If you don't have a Vault Server, you can deploy it by using the KubeVault operator.

- [Deploy Vault Server](/docs/guides/vault-server/vault-server.md)

The KubeVault operator can manage policies and secret engines of Vault servers which are not provisioned by the KubeVault operator. You need to configure both the Vault server and the cluster so that the KubeVault operator can communicate with your Vault server.

- [Configure cluster and Vault server](/docs/guides/vault-server/external-vault-sever.md#configuration)

Now, we have the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains connection and authentication information about the Vault server.

```bash
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

## Enable and Configure Azure Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on specified path and configure the secret engine with given configurations.

A sample SecretEngine object for Azure secret engine:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: azure-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  azure:
    credentialSecret: azure-cred
```

To configure the Azure secret engine, you need to provide azure credentials through a Kubernetes secret.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: azure-cred
  namespace: demo
data:
  client-secret: eyJtc2ciOiJleGFtcGxlIn0=
  subscription-id: eyJtc2ciOiJleGFtcGxlIn0=
  client-id: eyJtc2ciOiJleGFtcGxlIn0=
  tenant-id: eyJtc2ciOiJleGFtcGxlIn0=
```

Let's deploy SecretEngine:

```bash
$ kubectl apply -f docs/examples/guides/secret-engines/azure/secret.yaml
secret/azure-cred created

$ kubectl apply -f docs/examples/guides/secret-engines/azure/secretengine.yaml
secretengine.engine.kubevault.com/azure-engine created
```

Wait till the status become `Success`:

```bash
$ kubectl get secretengines -n demo
NAME           STATUS
azure-engine   Success
```

Since the status is `Success`, the Azure secret engine is enabled and successfully configured. You can use `kubectl describe secretengine -n <namepsace> <name>` to check for error events, if any.

## Create Azure Role

By using [AzureRole](/docs/concepts/secret-engine-crds/azure-secret-engine/azurerole.md), you can create a [role](https://www.vaultproject.io/api/secret/azure/index.html#create-update-role) on the Vault server in Kubernetes native way.

A sample AzureRole object is given below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AzureRole
metadata:
  name: azure-role
  namespace: demo
spec:
  secretEngineRef:
    name: vault
  applicationObjectID: e211afbc-cc4a-462f-ad6f-59e26eb5406f
  ttl: 1h
```

Let's deploy AzureRole:

```bash
$ kubectl apply -f docs/examples/guides/secret-engines/azure/secretenginerole.yaml
azurerole.engine.kubevault.com/azure-role created

$ kubectl get azureroles -n demo
NAME         STATUS
azure-role   Success
```

You can also check from Vault that the role is created.
To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

```bash
$ vault list azure/roles
Keys
----
k8s.-.demo.azure-role

$ vault read azure/roles/k8s.-.demo.azure-role
Key                      Value
---                      -----
application_object_id    c1cb042d-96d7-423a-8dba-243c2e5010d3
azure_roles              <nil>
max_ttl                  0s
ttl                      1h
```

If we delete the AzureRole, then the respective role will be deleted from the Vault.

```bash
$ kubectl delete -f docs/examples/guides/secret-engines/azure/secretenginerole.yaml
  azurerole.engine.kubevault.com "azure-role" deleted
```

Check from Vault whether the role exists:

```bash
$ vault read azure/roles/k8s.-.demo.azure-role
  No value found at azure/roles/k8s.-.demo.azure-role

$ vault list azure/roles
  No value found at azure/roles/
```

## Generate Azure credentials


Here, we are going to make a request to Vault for Azure credentials by creating `azure-cred-rqst` SecretAccessRequest in `demo` namespace.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: azure-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: AzureRole
    name: azure-role 
  subjects:
  - kind: ServiceAccount
    name: demo-sa
    namespace: demo
```

Here, `spec.roleRef` is the reference of AzureRole against which credentials will be issued. `spec.subjects` is the reference to the object or user identities a role binding applies to and it will have read access of the credential secret.

Now, we are going to create SecretAccessRequest.

```bash
$ kubectl apply -f docs/examples/guides/secret-engines/azure/azureaccessrequest.yaml
secretaccessrequest.engine.kubevault.com/azure-cred-rqst created

$ kubectl get secretaccessrequests -n demo
NAME        AGE
azure-cred-rqst  3s
```

Azure credentials will not be issued until it is approved. The KubeVault operator will watch for the approval in the `status.conditions[].type` field of the request object. You can use [KubeVault CLI](https://github.com/kubevault/cli), a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), to approve or deny SecretAccessRequest.

```bash
# using KubeVault CLI as kubectl plugin to approve request
$ kubectl vault approve secretaccessrequest azure-cred-rqst -n demo
  approved

$ kubectl get secretaccessrequests -n demo azure-cred-rqst -o yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: azure-cred-rqst
  namespace: demo
spec:
  roleRef:
    name: azure-role
    namespace: demo
  subjects:
  - kind: ServiceAccount
    name: demo-sa
    namespace: demo
status:
  conditions:
  - lastUpdateTime: "2019-11-14T09:21:49Z"
    message: This was approved by kubectl vault approve secretaccessrequest
    reason: KubectlApprove
    type: Approved
  lease:
    duration: 1h0m0s
    id: azure/creds/k8s.-.demo.azure-role/FJVEWUW9NpGlFOdIIMd900Zr
    renewable: true
  secret:
    name: azure-cred-rqst-luc5p4


```

Once SecretAccessRequest is approved, the KubeVault operator will issue credentials from Vault and create a secret containing the credential. It will also create a role and rolebinding so that `spec.subjects` can access secret. You can view the information in the `status` field.

```bash
$ kubectl get secretaccessrequest azure-cred-rqst -n demo -o json | jq '.status'
{
  "conditions": [
    {
      "lastUpdateTime": "2019-11-14T09:21:49Z",
      "message": "This was approved by kubectl vault approve secretaccessrequest",
      "reason": "KubectlApprove",
      "type": "Approved"
    }
  ],
  "lease": {
    "duration": "1h0m0s",
    "id": "azure/creds/k8s.-.demo.azure-role/FJVEWUW9NpGlFOdIIMd900Zr",
    "renewable": true
  },
  "secret": {
    "name": "azure-cred-rqst-luc5p4"
  }
}

$  kubectl get secret -n demo azure-cred-rqst-luc5p4 -o yaml
apiVersion: v1
data:
  client_id: MmI4NzFkNGEtN...
  client_secret: ZjJlMjA3N...
kind: Secret
metadata:
  name: azure-cred-rqst-luc5p4
  namespace: demo
  ownerReferences:
  - apiVersion: engine.kubevault.com/v1alpha1
    controller: true
    kind: SecretAccessRequest
    name: azure-cred-rqst
    uid: d944491b-a22c-4777-bc8f-2e2c94b47b7b
type: Opaque

```

If SecretAccessRequest is deleted, then credential lease (if any) will be revoked.

```bash
$ kubectl delete secretaccessrequest -n demo azure-cred-rqst
secretaccessrequest.engine.kubevault.com "azure-cred-rqst" deleted
```

If SecretAccessRequest is `Denied`, then the KubeVault operator will not issue any credentials.

```bash
$ kubectl vault deny secretaccessrequest azure-cred-rqst -n demo
  Denied

$ kubectl get secretaccessrequest  azure-cred-rqst -n demo -o json | jq '.status'
  {
    "conditions": [
      {
        "lastUpdateTime": "2019-11-14T09:21:49Z",
        "message": "This was denied by kubectl vault deny secretaccessrequest",
        "reason": "KubectlDeny",
        "type": "Denied"
      }
    ]
  }
```

> Note: Once SecretAccessRequest is `Approved` or `Denied`, you cannot change `spec.roleRef` and `spec.subjects` field.
