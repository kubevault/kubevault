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
- [AzureAccessKeyRequest](/docs/concepts/secret-engine-crds/azure-secret-engine/azureaccesskeyrequest.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/operator/install.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we are going to create a [role](https://www.vaultproject.io/api/secret/azure/index.html#create-update-role) using AzureRole and issue credential using AzureAccessKeyRequest.

## Vault Server

If you don't have a Vault Server, you can deploy it by using the KubeVault operator.

- [Deploy Vault Server](/docs/guides/vault-server/vault-server.md)

The KubeVault operator can manage policies and secret engines of Vault servers which are not provisioned by the KubeVault operator. You need to configure both the Vault server and the cluster so that the KubeVault operator can communicate with your Vault server.

- [Configure cluster and Vault server](/docs/guides/vault-server/external-vault-sever.md#configuration)

Now, we have the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains connection and authentication information about the Vault server.

```console
$ kubectl get appbinding -n demo
NAME    AGE
vault   50m

$ kubectl get appbinding -n demo vault -o yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9URXhNVEl3T1RFMU5EQmFGdzB5T1RFeE1Ea3dPVEUxTkRCYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdFZFZmtic2c2T085dnM2d1Z6bTlPQ1FYClBtYzBYTjlCWjNMbXZRTG0zdzZGaWF2aUlSS3VDVk1hN1NRSGo2L2YvOHZPeWhqNEpMcHhCM0hCYVFPZ3RrM2QKeEFDbHppU1lEd3dDbGEwSThxdklGVENLWndreXQzdHVQb0xybkppRFdTS2xJait6aFZDTHZ0enB4MDE3SEZadApmZEdhUUtlSXREUVdyNUV1QWlCMjhhSVF4WXREaVN6Y0h3OUdEMnkrblRMUEd4UXlxUlhua0d1UlIvR1B3R3lLClJ5cTQ5NmpFTmFjOE8wVERYRkIydWJQSFNza2xOU1VwSUN3S1IvR3BobnhGak1rWm4yRGJFZW9GWDE5UnhzUmcKSW94TFBhWDkrRVZxZU5jMlczN2MwQlhBSGwyMHVJUWQrVytIWDhnOVBVVXRVZW9uYnlHMDMvampvNERJRHdJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFabHRFN0M3a3ZCeTNzeldHY0J0SkpBTHZXY3ZFeUdxYUdCYmFUbGlVbWJHTW9QWXoKbnVqMUVrY1I1Qlg2YnkxZk15M0ZtZkJXL2E0NU9HcDU3U0RMWTVuc2w0S1RlUDdGZkFYZFBNZGxrV0lQZGpnNAptOVlyOUxnTThkOGVrWUJmN0paUkNzcEorYkpDU1A2a2p1V3l6MUtlYzBOdCtIU0psaTF3dXIrMWVyMUprRUdWClBQMzFoeTQ2RTJKeFlvbnRQc0d5akxlQ1NhTlk0UWdWK3ZneWJmSlFEMVYxbDZ4UlVlMzk2YkJ3aS94VGkzN0oKNWxTVklmb1kxcUlBaGJPbjBUWHp2YzBRRXBKUExaRDM2VDBZcEtJSVhjZUVGYXNxZzVWb1pINGx1Uk50SStBUAp0blg4S1JZU0xGOWlCNEJXd0N0aGFhZzZFZVFqYWpQNWlxZnZoUT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    service:
      name: vault
      port: 8200
      scheme: HTTPS
  parameters:
    apiVersion: config.kubevault.com/v1alpha1
    kind: VaultServerConfiguration
    path: kubernetes
    vaultRole: vault-policy-controller
    kubernetes:
      serviceAccountName: vault
      tokenReviewerServiceAccountName: vault-k8s-token-reviewer
      usePodServiceAccountForCSIDriver: true
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

```console
$ kubectl apply -f docs/examples/guides/secret-engines/azure/azureCred.yaml
secret/azure-cred created

$ kubectl apply -f docs/examples/guides/secret-engines/azure/azureSecretEngine.yaml
secretengine.engine.kubevault.com/azure-engine created
```

Wait till the status become `Success`:

```console
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
  vaultRef:
    name: vault
  applicationObjectID: c1cb042d-96d7-423a-8dba-243c2e5010d3
  ttl: 1h
```

Let's deploy AzureRole:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/azure/azureRole.yaml
azurerole.engine.kubevault.com/azure-role created

$ kubectl get azureroles -n demo
NAME         STATUS
azure-role   Success
```

You can also check from Vault that the role is created.
To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

```console
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

```console
$ kubectl delete -f docs/examples/guides/secret-engines/azure/azureRole.yaml
  azurerole.engine.kubevault.com "azure-role" deleted
```

Check from Vault whether the role exists:

```console
$ vault read azure/roles/k8s.-.demo.azure-role
  No value found at azure/roles/k8s.-.demo.azure-role

$ vault list azure/roles
  No value found at azure/roles/
```

## Generate Azure credentials

By using [AzureAccessKeyRequest](/docs/concepts/secret-engine-crds/azure-secret-engine/azureaccesskeyrequest.md), you can generate Azure credential from Vault.

Here, we are going to make a request to Vault for Azure credentials by creating `azure-cred-rqst` AzureAccessKeyRequest in `demo` namespace.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AzureAccessKeyRequest
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
```

Here, `spec.roleRef` is the reference of AzureRole against which credentials will be issued. `spec.subjects` is the reference to the object or user identities a role binding applies to and it will have read access of the credential secret.

Now, we are going to create AzureAccessKeyRequest.

```console
$ kubectl apply -f docs/examples/guides/secret-engines/azure/azureAccessKeyRequest.yaml
azureaccesskeyrequest.engine.kubevault.com/azure-cred-rqst created

$ kubectl get azureaccesskeyrequests -n demo
NAME        AGE
azure-cred-rqst  3s
```

Azure credentials will not be issued until it is approved. The KubeVault operator will watch for the approval in the `status.conditions[].type` field of the request object. You can use [KubeVault CLI](https://github.com/kubevault/cli), a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), to approve or deny AzureAccessKeyRequest.

```console
# using KubeVault CLI as kubectl plugin to approve request
$ kubectl vault approve azureaccesskeyrequest azure-cred-rqst -n demo
  approved

$ kubectl get azureaccesskeyrequests -n demo azure-cred-rqst -o yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AzureAccessKeyRequest
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
    message: This was approved by kubectl vault approve azureaccesskeyrequest
    reason: KubectlApprove
    type: Approved
  lease:
    duration: 1h0m0s
    id: azure/creds/k8s.-.demo.azure-role/FJVEWUW9NpGlFOdIIMd900Zr
    renewable: true
  secret:
    name: azure-cred-rqst-luc5p4


```

Once AzureAccessKeyRequest is approved, the KubeVault operator will issue credentials from Vault and create a secret containing the credential. It will also create a role and rolebinding so that `spec.subjects` can access secret. You can view the information in the `status` field.

```console
$ kubectl get azureaccesskeyrequest azure-cred-rqst -n demo -o json | jq '.status'
{
  "conditions": [
    {
      "lastUpdateTime": "2019-11-14T09:21:49Z",
      "message": "This was approved by kubectl vault approve azureaccesskeyrequest",
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
    kind: AzureAccessKeyRequest
    name: azure-cred-rqst
    uid: d944491b-a22c-4777-bc8f-2e2c94b47b7b
type: Opaque

```

If AzureAccessKeyRequest is deleted, then credential lease (if any) will be revoked.

```console
$ kubectl delete azureaccesskeyrequest -n demo azure-cred-rqst
azureaccesskeyrequest.engine.kubevault.com "azure-cred-rqst" deleted
```

If AzureAccessKeyRequest is `Denied`, then the KubeVault operator will not issue any credentials.

```console
$ kubectl vault deny azureaccesskeyrequest azure-cred-rqst -n demo
  Denied

$ kubectl get azureaccesskeyrequest  azure-cred-rqst -n demo -o json | jq '.status'
  {
    "conditions": [
      {
        "lastUpdateTime": "2019-11-14T09:21:49Z",
        "message": "This was denied by kubectl vault deny azureaccesskeyrequest",
        "reason": "KubectlDeny",
        "type": "Denied"
      }
    ]
  }
```

> Note: Once AzureAccessKeyRequest is `Approved` or `Denied`, you cannot change `spec.roleRef` and `spec.subjects` field.
