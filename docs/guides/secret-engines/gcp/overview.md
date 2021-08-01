---
title: Manage GCP IAM Secrets using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-gcp
    name: Overview
    parent: gcp-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage GCP IAM Secrets using the KubeVault operator

The Google Cloud Vault secrets engine dynamically generates Google Cloud service account keys and OAuth tokens based on IAM policies. This enables users to gain access to Google Cloud resources without needing to create or manage a dedicated service account. You can easily manage [GCP secret engine](https://www.vaultproject.io/docs/secrets/gcp/index.html) using the KubeVault operator.

![GCP secret engine](/docs/images/guides/secret-engines/gcp/gcp_guide.svg)

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [GCPRole](/docs/concepts/secret-engine-crds/gcp-secret-engine/gcprole.md)
- [GCPAccessKeyRequest](/docs/concepts/secret-engine-crds/gcp-secret-engine/gcpaccesskeyrequest.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we are going to create a [role](https://www.vaultproject.io/api/secret/gcp/index.html#create-update-roleset) using GCPRole and issue credential using GCPAccessKeyRequest.

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

## Enable and Configure GCP Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on specified path and configure the secret engine with given configurations.

A sample SecretEngine object for GCP secret engine:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: gcp-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  gcp:
    credentialSecret: gcp-cred
```

To configure the GCP secret engine, you need to provide google service account credentials through a Kubernetes secret.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gcp-cred
  namespace: demo
data:
  sa.json: eyJtc2ciOiJleGFtcGxlIn0= ## base64 encoded google service account credential
```

Let's deploy SecretEngine:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/gcp/gcpCred.yaml
secret/gcp-cred created

$ kubectl apply -f docs/examples/guides/secret-engines/gcp/gcpSecretEngine.yaml
secretengine.engine.kubevault.com/gcp-engine created
```

Wait till the status become `Success`:

```console
$ kubectl get secretengines -n demo
NAME         STATUS
gcp-engine   Success
```

Since the status is `Success`, the GCP secret engine is enabled and successfully configured. You can use `kubectl describe secretengine -n <namepsace> <name>` to check for error events, if any.

## Create GCP Roleset

By using [GCPRole](/docs/concepts/secret-engine-crds/gcp-secret-engine/gcprole.md), you can [create gcp roleset](https://www.vaultproject.io/api/secret/gcp/index.html#create-update-roleset) on the Vault server in Kubernetes native way.

A sample GCPRole object is given below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: GCPRole
metadata:
  name: gcp-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  secretType: access_token
  project: ackube
  bindings: 'resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
    roles = ["roles/viewer"]
    }'
  tokenScopes:
    - https://www.googleapis.com/auth/cloud-platform
```

Let's deploy GCPRole:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/gcp/gcpRole.yaml
gcprole.engine.kubevault.com/gcp-role created

$ kubectl get gcprole -n demo
NAME       STATUS
gcp-role   Success
```

You can also check from Vault that the roleset is created.
To resolve the naming conflict, name of the roleset in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

```console
$ vault list gcp/roleset
Keys
----
k8s.-.demo.gcp-role

$ vault read gcp/roleset/k8s.-.demo.gcp-role
Key                      Value
---                      -----
bindings                 map[//cloudresourcemanager.googleapis.com/projects/ackube:[roles/viewer]]
project                  ackube
secret_type              access_token
service_account_email    vaultk8s---demo-gcp-424523423@ackube.iam.gserviceaccount.com
token_scopes             [https://www.googleapis.com/auth/cloud-platform]
```

If we delete the GCPRole, then the respective role will be deleted from the Vault.

```console
$ kubectl delete -f docs/examples/guides/secret-engines/gcp/gcpRole.yaml
  gcprole.engine.kubevault.com "gcp-role" deleted
```

Check from Vault whether the roleset exists:

```console
$ vault read gcp/roleset/k8s.-.demo.gcp-role
  No value found at gcp/roleset/k8s.-.demo.gcp-role

$ vault list gcp/roleset
  No value found at gcp/roleset/
```

## Generate GCP credentials

By using [GCPAccessKeyRequest](/docs/concepts/secret-engine-crds/gcp-secret-engine/gcpaccesskeyrequest.md), you can generate GCP credential from Vault.

Here, we are going to make a request to Vault for GCP credential by creating `gcp-cred-req` GCPAccessKeyRequest in `demo` namespace.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: GCPAccessKeyRequest
metadata:
  name: gcp-cred-req
  namespace: demo
spec:
  roleRef:
    name: gcp-role
    namespace: demo
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Here, `spec.roleRef` is the reference of GCPRole against which credentials will be issued. `spec.subjects` is the reference to the object or user identities a role binding applies to and it will have read access of the credential secret.

Now, we are going to create GCPAccessKeyRequest.

```console
$ kubectl apply -f docs/examples/guides/secret-engines/gcp/gcpAccessKeyRequest.yaml
gcpaccesskeyrequest.engine.kubevault.com/gcp-cred-req created

$ kubectl get gcpaccesskeyrequests -n demo
NAME        AGE
gcp-cred-req  3s
```

GCP credentials will not be issued until it is approved. The KubeVault operator will watch for the approval in the `status.conditions[].type` field of the request object. You can use [KubeVault CLI](https://github.com/kubevault/cli), a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), to approve or deny GCPAccessKeyRequest.

```console
# using KubeVault CLI as kubectl plugin to approve request
$ kubectl vault approve gcpaccesskeyrequest gcp-cred-req -n demo
  approved

$ kubectl get gcpaccesskeyrequest -n demo gcp-cred-req -o yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: GCPAccessKeyRequest
metadata:
  name: gcp-cred-req
  namespace: demo
spec:
  roleRef:
    name: gcp-role
    namespace: demo
  subjects:
  - kind: ServiceAccount
    name: demo-sa
    namespace: demo
status:
  conditions:
  - lastUpdateTime: "2019-11-12T10:40:30Z"
    message: This was approved by kubectl vault approve gcpaccesskeyrequest
    reason: KubectlApprove
    type: Approved
  lease:
    duration: 0s
  secret:
    name: gcp-cred-req-unyzu6

```

Once GCPAccessKeyRequest is approved, the KubeVault operator will issue credentials from Vault and create a secret containing the credential. It will also create a role and rolebinding so that `spec.subjects` can access secret. You can view the information in the `status` field.

```console
$ kubectl get gcpaccesskeyrequest gcp-cred-req -n demo -o json | jq '.status'
{
  "conditions": [
    {
      "lastUpdateTime": "2019-11-12T10:40:30Z",
      "message": "This was approved by kubectl vault approve gcpaccesskeyrequest",
      "reason": "KubectlApprove",
      "type": "Approved"
    }
  ],
  "lease": {
    "duration": "0s"
  },
  "secret": {
    "name": "gcp-cred-req-unyzu6"
  }
}

$ kubectl get secret -n demo gcp-cred-req-unyzu6 -o yaml
apiVersion: v1
data:
  expires_at_seconds: MTU3MzU1ODgzMA==
  token: eWEyOS5jLktsMndCLTR5=
  token_ttl: MzU5OQ==
kind: Secret
metadata:
  name: gcp-cred-req-unyzu6
  namespace: demo
type: Opaque

```

If GCPAccessKeyRequest is deleted, then credential lease (if any) will be revoked.

```console
$ kubectl delete gcpaccesskeyrequest -n demo gcp-cred-req
  gcpaccesskeyrequest.engine.kubevault.com "gcp-cred-req" deleted
```

If GCPAccessKeyRequest is `Denied`, then the KubeVault operator will not issue any credential.

```console
$ kubectl vault deny gcpaccesskeyrequest gcp-cred-req -n demo
  Denied
```

> Note: Once GCPAccessKeyRequest is `Approved` or `Denied`, you cannot change `spec.roleRef` and `spec.subjects` field.
