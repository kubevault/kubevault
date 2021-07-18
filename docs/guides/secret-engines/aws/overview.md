---
title: Manage AWS IAM Secrets using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-aws
    name: Overview
    parent: aws-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage AWS IAM Secrets using the KubeVault operator

The AWS secrets engine generates AWS access credentials dynamically based on IAM policies. The AWS IAM credentials are time-based and are automatically revoked when the Vault lease expires. You can easily manage the [AWS secret engine](https://www.vaultproject.io/docs/secrets/aws/index.html) using KubeVault operator.

![AWS secret engine](/docs/images/guides/secret-engines/aws/aws_secret_engine_guide.svg)

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [AWSRole](/docs/concepts/secret-engine-crds/aws-secret-engine/awsrole.md)
- [AWSAccessKeyRequest](/docs/concepts/secret-engine-crds/aws-secret-engine/awsaccesskeyrequest.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/operator/install.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we are going to create a [role](https://www.vaultproject.io/api/secret/aws/index.html#create-update-role) using AWSRole and issue credential using AWSAccessKeyRequest.

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

## Enable and Configure AWS Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on specified path and configure the secret engine with given configurations.

A sample SecretEngine object for AWS secret engine:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: aws-secret-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  aws:
    credentialSecret: aws-cred
    region: us-east-1
    leaseConfig:
      lease: 1h
      leaseMax: 1h
```

To configure the AWS secret engine, you need to provide `aws_access_key_id` and `aws_secret_access_key` through a Kubernetes secret.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-cred
  namespace: demo
data:
  access_key: eyJtc2ciOiJleGFtcGxlIn0= # base64 encoded aws access key id
  secret_key: eyJtc2ciOiJleGFtcGxlIn0= # base64 encoded aws secret access key
```

Let's deploy SecretEngine:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/aws/awsCred.yaml
secret/aws-cred created

$ kubectl apply -f docs/examples/guides/secret-engines/aws/awsSecretEngine.yaml
secretengine.engine.kubevault.com/aws-secret-engine created
```

Wait till the status become `Success`:

```console
$ kubectl get secretengines -n demo
NAME                STATUS
aws-secret-engine   Success
```

Since the status is `Success`, the AWS secret engine is enabled and successfully configured. You can use `kubectl describe secretengine -n <namepsace> <name>` to check for error events if any.

## Create AWS Role

By using [AWSRole](/docs/concepts/secret-engine-crds/aws-secret-engine/awsrole.md), you can create a [role](https://www.vaultproject.io/api/secret/aws/index.html#create-update-role) on the Vault server in Kubernetes native way.

A sample AWSRole object is given below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AWSRole
metadata:
  name: aws-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  credentialType: iam_user
  policyDocument: |
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": "ec2:*",
          "Resource": "*"
        }
      ]
    }
```

Let's deploy AWSRole:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/aws/awsRole.yaml
awsrole.engine.kubevault.com/aws-role created

$ kubectl get awsrole -n demo
NAME       STATUS
aws-role   Success
```

You can also check from Vault that the role is created.
To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

```console
$ vault list aws/roles
Keys
----
k8s.-.demo.aws-role

$ vault read aws/roles/k8s.-.demo.aws-role
Key                Value
---                -----
credential_type    iam_user
default_sts_ttl    0s
max_sts_ttl        0s
policy_arns        <nil>
policy_document    {"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"ec2:*","Resource":"*"}]}
role_arns          <nil>
user_path          n/a
```

If we delete the AWSRole, then the respective role will be deleted from the Vault.

```console
$ kubectl delete awsrole -n demo aws-role
awsrole.engine.kubevault.com "aws-role" deleted
```

Check from Vault whether the role exists:

```console
$ vault read aws/roles/k8s.-.demo.aws-role
No value found at aws/roles/k8s.-.demo.aws-role

$ vault list aws/roles
No value found at aws/roles/
```

## Generate AWS credentials

By using [AWSAccessKeyRequest](/docs/concepts/secret-engine-crds/aws-secret-engine/awsaccesskeyrequest.md), you can generate AWS credentials from Vault.

Here, we are going to make a request to Vault for AWS credential by creating `aws-cred-rqst` AWSAccessKeyRequest in `demo` namespace.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AWSAccessKeyRequest
metadata:
  name: aws-cred-rqst
  namespace: demo
spec:
  roleRef:
    name: aws-role
    namespace: demo
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Here, `spec.roleRef` is the reference of AWSRole against which credentials will be issued. `spec.subjects` is the reference to the object or user identities a role binding applies to and it will have read access of the credential secret.

Now, we are going to create an AWSAccessKeyRequest.

```console
$ kubectl apply -f docs/examples/guides/secret-engines/aws/awsAccessKeyRequest.yaml
awsaccesskeyrequest.engine.kubevault.com/aws-cred-rqst created

$ kubectl get awsaccesskeyrequest -n demo
NAME            AGE
aws-cred-rqst   35s
```

AWS credentials will not be issued until it is approved. The KubeVault operator will watch for the approval in the `status.conditions[].type` field of the request object. You can use [KubeVault CLI](https://github.com/kubevault/cli), a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), to approve or deny AWSAccessKeyRequest.

```console
# using KubeVault CLI as kubectl plugin to approve request
$ kubectl vault approve awsaccesskeyrequest aws-cred-rqst -n demo
  approved

$ kubectl get awsaccesskeyrequest -n demo aws-cred-rqst -o yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: AWSAccessKeyRequest
metadata:
  name: aws-cred-rqst
  namespace: demo
spec:
  roleRef:
    name: aws-role
    namespace: demo
  subjects:
  - kind: ServiceAccount
    name: demo-sa
    namespace: demo
status:
  conditions:
  - lastUpdateTime: "2019-11-13T12:18:07Z"
    message: This was approved by kubectl vault approve awsaccesskeyrequest
    reason: KubectlApprove
    type: Approved
  lease:
    duration: 1h0m0s
    id: aws/creds/k8s.-.demo.aws-role/X9dCjtiQCykbuJ7UmzM64xfh
    renewable: true
  secret:
    name: aws-cred-rqst-ryym7w
```

Once AWSAccessKeyRequest is approved, the KubeVault operator will issue credentials from Vault and create a secret containing the credential. It will also create a role and rolebinding so that `spec.subjects` can access secret. You can view the information in the `status` field.

```console
$ kubectl get awsaccesskeyrequest aws-cred-rqst -n demo -o json | jq '.status'
{
  "conditions": [
    {
      "lastUpdateTime": "2019-11-13T12:18:07Z",
      "message": "This was approved by kubectl vault approve awsaccesskeyrequest",
      "reason": "KubectlApprove",
      "type": "Approved"
    }
  ],
  "lease": {
    "duration": "1h0m0s",
    "id": "aws/creds/k8s.-.demo.aws-role/X9dCjtiQCykbuJ7UmzM64xfh",
    "renewable": true
  },
  "secret": {
    "name": "aws-cred-rqst-ryym7w"
  }
}

$ kubectl get secret -n demo aws-cred-rqst-ryym7w -o yaml
apiVersion: v1
data:
  access_key: QUtJQVdTWV....=
  secret_key: RVA1dXdXWnVlTX....==
  security_token: ""
kind: Secret
metadata:
  name: aws-cred-rqst-ryym7w
  namespace: demo
  ownerReferences:
  - apiVersion: engine.kubevault.com/v1alpha1
    controller: true
    kind: AWSAccessKeyRequest
    name: aws-cred-rqst
type: Opaque
```

If AWSAccessKeyRequest is deleted, then credential lease (if any) will be revoked.

```console
$ kubectl delete awsaccesskeyrequest -n demo aws-cred-rqst
awsaccesskeyrequest.engine.kubevault.com "aws-cred-rqst" deleted
```

If AWSAccessKeyRequest is `Denied`, then the KubeVault operator will not issue any credential.

```console
$ kubectl vault deny awsaccesskeyrequest aws-cred-rqst -n demo
  Denied
```

> Note: Once AWSAccessKeyRequest is `Approved` or `Denied`, you can not change `spec.roleRef` and `spec.subjects` field.
