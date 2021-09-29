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

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).

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
$ kubectl apply -f docs/examples/guides/secret-engines/aws/secret.yaml
secret/aws-cred created

$ kubectl apply -f docs/examples/guides/secret-engines/aws/secretengine.yaml
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
  secretEngineRef:
    name: aws-secret-engine
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
$ kubectl apply -f docs/examples/guides/secret-engines/aws/secretenginerole.yaml
awsrole.engine.kubevault.com/aws-role created

$ kubectl get awsrole -n demo
NAME       STATUS
aws-role   Success
```

You can also check from Vault that the role is created.
To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

```console
$ vault list your-aws-path/roles
Keys
----
k8s.-.demo.aws-role

$ vault read your-aws-path/roles/k8s.-.demo.aws-role
Key                         Value
---                         -----
credential_type             iam_user
default_sts_ttl             0s
iam_groups                  <nil>
iam_tags                    <nil>
max_sts_ttl                 0s
permissions_boundary_arn    n/a
policy_arns                 <nil>
policy_document             {"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"ec2:*","Resource":"*"}]}
role_arns                   <nil>
user_path                   n/a
```

If we delete the AWSRole, then the respective role will be deleted from the Vault.

```console
$ kubectl delete awsrole -n demo aws-role
awsrole.engine.kubevault.com "aws-role" deleted
```

Check from Vault whether the role exists:

```console
$ vault read your-aws-path/roles/k8s.-.demo.aws-role
No value found at your-aws-path/roles/k8s.-.demo.aws-role

$ vault list aws/roles
No value found at aws/roles/
```

## Generate AWS credentials

Here, we are going to make a request to Vault for AWS credential by creating `aws-cred-rqst` SecretAccessRequest in `demo` namespace.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: aws-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: AWSRole
    name: aws-role
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Here, `spec.roleRef` is the reference of AWSRole against which credentials will be issued. `spec.subjects` is the reference to the object or user identities a role binding applies to and it will have read access of the credential secret.

Now, we are going to create an SecretAccessRequest.

```console
$ kubectl apply -f docs/examples/guides/secret-engines/aws/secretaccessrequest.yaml
secretaccessrequest.engine.kubevault.com/aws-cred-rqst created

$ kubectl get secretaccessrequest -n demo
NAME            AGE
aws-cred-rqst   35s
```

AWS credentials will not be issued until it is approved. The KubeVault operator will watch for the approval in the `status.conditions[].type` field of the request object. You can use [KubeVault CLI](https://github.com/kubevault/cli), a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), to approve or deny SecretAccessRequest.

```console
# using KubeVault CLI as kubectl plugin to approve request
$ kubectl vault approve secretaccessrequest aws-cred-rqst -n demo
  approved

$ kubectl get secretaccessrequest -n demo aws-cred-rqst -o yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
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
    message: This was approved by kubectl vault approve secretaccessrequest
    reason: KubectlApprove
    type: Approved
  lease:
    duration: 1h0m0s
    id: your-aws-path/creds/k8s.-.demo.aws-role/X9dCjtiQCykbuJ7UmzM64xfh
    renewable: true
  secret:
    name: aws-cred-rqst-ryym7w
```

Once SecretAccessRequest is approved, the KubeVault operator will issue credentials from Vault and create a secret containing the credential. It will also create a role and rolebinding so that `spec.subjects` can access secret. You can view the information in the `status` field.

```console
$ kubectl get secretaccessrequest aws-cred-rqst -n demo -o json | jq '.status'
{
  "conditions": [
    {
      "lastUpdateTime": "2019-11-13T12:18:07Z",
      "message": "This was approved by kubectl vault approve secretaccessrequest",
      "reason": "KubectlApprove",
      "type": "Approved"
    }
  ],
  "lease": {
    "duration": "1h0m0s",
    "id": "your-aws-path/creds/k8s.-.demo.aws-role/X9dCjtiQCykbuJ7UmzM64xfh",
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
    kind: SecretAccessRequest
    name: aws-cred-rqst
type: Opaque
```

If SecretAccessRequest is deleted, then credential lease (if any) will be revoked.

```console
$ kubectl delete secretaccessrequest -n demo aws-cred-rqst
secretaccessrequest.engine.kubevault.com "aws-cred-rqst" deleted
```

If SecretAccessRequest is `Denied`, then the KubeVault operator will not issue any credential.

```console
$ kubectl vault deny secretaccessrequest aws-cred-rqst -n demo
  Denied
```

> Note: Once SecretAccessRequest is `Approved`, you can not change `spec.roleRef` and `spec.subjects` field.
