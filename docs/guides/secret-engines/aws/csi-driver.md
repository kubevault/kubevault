---
title: Mount AWS IAM Secrets using CSI Driver
menu:
  docs_{{ .version }}:
    identifier: csi-driver-aws
    name: CSI Driver
    parent: aws-secret-engines
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="KubeVault's built-in CSI driver has been removed in favor of [Secrets Store CSI driver for Kubernetes secrets](https://github.com/kubernetes-sigs/secrets-store-csi-driver)." >}}

# Mount AWS IAM Secrets using CSI Driver

At first, you need to have a Kubernetes 1.14 or later cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/). To check the version of your cluster, run:

```console
$ kubectl version --short
Client Version: v1.16.2
Server Version: v1.14.0
```

Before you begin:

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Install Secrets Store CSI driver for Kubernetes secrets in your cluster from [here](https://secrets-store-csi-driver.sigs.k8s.io/getting-started/installation.html).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial stored in [examples](/docs/examples/guides/secret-engines/aws) folder in GitHub repository [KubeVault/docs](https://github.com/kubevault/kubevault)

## Vault Server

If you don't have a Vault Server, you can deploy it by using the KubeVault operator.

- [Deploy Vault Server](/docs/guides/vault-server/vault-server.md)

The KubeVault operator can manage policies and secret engines of Vault servers which are not provisioned by the KubeVault operator. You need to configure both the Vault server and the cluster so that the KubeVault operator can communicate with your Vault server.

- [Configure cluster and Vault server](/docs/guides/vault-server/external-vault-sever.md#configuration)

Now, we have the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) that contains connection and authentication information about the Vault server. And we also have the service account that the Vault server can authenticate.

```console
$ kubectl get serviceaccounts -n demo
NAME                       SECRETS   AGE
vault                      1         20h

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
    kubernetes:
      serviceAccountName: vault
      tokenReviewerServiceAccountName: vault-k8s-token-reviewer
      usePodServiceAccountForCSIDriver: true
    path: kubernetes
    vaultRole: vault-policy-controller
```

## Enable and Configure AWS Secret Engine

The following steps are required to enable and configure the AWS secrets engine in the Vault server.

There are two ways to configure the Vault server. You can either use the `KubeVault operator` or the  `Vault CLI` to manually configure a Vault server.

<ul class="nav nav-tabs" id="conceptsTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="operator-tab" data-toggle="tab" href="#operator" role="tab" aria-controls="operator" aria-selected="true">Using KubeVault operator</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="csi-driver-tab" data-toggle="tab" href="#csi-driver" role="tab" aria-controls="csi-driver" aria-selected="false">Using Vault CLI</a>
  </li>
</ul>
<div class="tab-content" id="conceptsTabContent">
  <div open class="tab-pane fade show active" id="operator" role="tabpanel" aria-labelledby="operator-tab">

### Using KubeVault operator

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [AWSRole](/docs/concepts/secret-engine-crds/aws-secret-engine/awsrole.md)

Let's enable and configure AWS secret engine by deploying the following `SecretEngine` yaml:

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

Configure an AWS role using the following `AWSRole` yaml:

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

</div>
<div class="tab-pane fade" id="csi-driver" role="tabpanel" aria-labelledby="csi-driver-tab">

### Using Vault CLI

You can also use [Vault CLI](https://www.vaultproject.io/docs/commands/) to [enable and configure](https://www.vaultproject.io/docs/secrets/aws/index.html#setup) the AWS secret engine.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

To generate secret from the AWS secret engine, you have to perform the following steps.

- **Enable `AWS`  secret engine:** To enable `AWS` secret engine run the following command.

```console
$ vault secrets enable aws
Success! Enabled the aws secrets engine at: aws/
```

- **Crete AWS config:** To communicate with AWS for generating IAM credentials, Vault needs to configure credentials. Run:

```console
$ vault write aws/config/root \
  access_key=AKIAJWVN5ZFT7NLNA \
  secret_key=R4nm063hgMVo4BTT5xOs5nH3Nt0i \
  region=us-east-1
Success! Data written to: aws/config/root
```

- **Configure a role:** We need to configure a vault role that maps to a set of permissions in AWS and an AWS credential type. When users generate credentials, they are generated against this role,

```console
$ vault write aws/roles/k8s.-.demo.aws-role \
  credential_type=iam_user \
  policy_document=-<<EOF
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
EOF
Success! Data written to: aws/roles/k8s.-.demo.aws-role
```

Here, `k8s.-.demo.aws-role` will be treated as a secret name on storage class.

- **Read the role:**

```console
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

If you use Vault CLI to enable and configure the AWS secret engine then you need to **update the vault policy** for the service account 'vault' created during vault server configuration and add the permission to read at "aws/roles/*" with previous permissions. That is why it is recommended to use the KubeVault operator because the operator updates the policies automatically when needed.

Find how to update the policy for service account in [here](/docs/guides/secret-engines/kv/csi-driver.md#update-vault-policy).

  </div>
</div>

## Mount secrets into a Kubernetes pod

Since Kubernetes 1.14, `storage.k8s.io/v1beta1` `CSINode` and `CSIDriver` objects were introduced. Let's check [CSIDriver](https://kubernetes-csi.github.io/docs/csi-driver-object.html) and [CSINode](https://kubernetes-csi.github.io/docs/csi-node-object.html) are available or not.

```console
$ kubectl get csidrivers
NAME                        CREATED AT
secrets.csi.kubevault.com   2019-12-09T04:32:50Z

$ kubectl get csinodes
NAME             CREATED AT
2gb-pool-57jj7   2019-12-09T04:32:52Z
2gb-pool-jrvtj   2019-12-09T04:32:58Z
```

So, we can create `StorageClass` now.

### Create StorageClass

Create `StorageClass` object with the following content:

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: vault-aws-storage
  annotations:
    storageclass.kubernetes.io/is-default-class: "false"
provisioner: secrets.csi.kubevault.com
parameters:
  ref: demo/vault # namespace/AppBinding, we created this while configuring vault server
  engine: AWS # vault engine name
  role: k8s.-.demo.aws-role # role name on vault which you want get access
  path: aws # specify the secret engine path, default is aws
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/aws/storageClass.yaml
storageclass.storage.k8s.io/vault-aws-storage created
```

## Test & Verify

Let's create a separate namespace called `trial` for testing purpose.

```console
$ kubectl create ns trial
namespace/trail created
```

### Create PVC

Create a `PersistentVolumeClaim` with the following data. This makes sure a volume will be created and provisioned on your behalf.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: csi-pvc-aws
  namespace: trial
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Mi
  storageClassName: vault-aws-storage
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/aws/pvc.yaml
persistentvolumeclaim/csi-pvc-aws created
```

### Create VaultPolicy and VaultPolicyBinding for Pod's Service Account

Let's say pod's service account name is `pod-sa` located in `trial` namespace. We need to create a [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) and a [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md) so that the pod has access to read secrets from the Vault server.

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: aws-se-policy
  namespace: demo
spec:
  vaultRef:
    name: vault
  # Here, aws secret engine is enabled at "aws".
  # If the path was "demo-se", policy should be like
  # path "demo-se/*" {}.
  policyDocument: |
    path "aws/*" {
      capabilities = ["create", "read"]
    }
---
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: aws-se-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  policies:
  - ref: aws-se-policy
  subjectRef:
    kubernetes:
      serviceAccountNames:
      - "pod-sa"
      serviceAccountNamespaces:
      - "trial"
```

Let's create VaultPolicy and VaultPolicyBinding:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/aws/vaultPolicy.yaml
vaultpolicy.policy.kubevault.com/aws-se-policy created

$ kubectl apply -f docs/examples/guides/secret-engines/aws/vaultPolicyBinding.yaml
vaultpolicybinding.policy.kubevault.com/aws-se-role created
```

Check if the VaultPolicy and the VaultPolicyBinding are successfully registered to the Vault server:

```console
$ kubectl get vaultpolicy -n demo
NAME                           STATUS    AGE
aws-se-policy                  Success   8s

$ kubectl get vaultpolicybinding -n demo
NAME                           STATUS    AGE
aws-se-role                    Success   10s
```

### Create Service Account for Pod

Let's create the service account `pod-sa` which was used in VaultPolicyBinding. When a VaultPolicyBinding object is created, the KubeVault operator create an auth role in the Vault server. The role name is generated by the following naming format: `k8s.(clusterName or -).namespace.name`. Here, it is `k8s.-.demo.aws-se-role`. We need to provide the auth role name as service account `annotations` while creating the service account. If the annotation `secrets.csi.kubevault.com/vault-role` is not provided, the CSI driver will not be able to perform authentication to the Vault.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pod-sa
  namespace: trial
  annotations:
    secrets.csi.kubevault.com/vault-role: k8s.-.demo.aws-se-role
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/aws/podServiceAccount.yaml
serviceaccount/pod-sa created
```

### Create Pod

Now we can create a Pod which refers to this volume. When the Pod is created, the volume will be attached, formatted and mounted to the specific container.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod
  namespace: trial
spec:
  containers:
  - name: mypod
    image: busybox
    command:
    - sleep
    - "3600"
    volumeMounts:
    - name: my-vault-volume
      mountPath: "/etc/aws"
      readOnly: true
  serviceAccountName: pod-sa # service account that was created
  volumes:
  - name: my-vault-volume
    persistentVolumeClaim:
      claimName: csi-pvc-aws
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/aws/pod.yaml
pod/mypod created
```

Check if the Pod is running successfully, by running:

```console
$ kubectl get pods -n demo
NAME                    READY   STATUS    RESTARTS   AGE
mypod                   1/1     Running   0          5m21s
```

### Verify Secret

If the Pod is running successfully, then check inside the app container by running

```console
$ kubectl exec -it -n trial  mypod sh
/ # ls /etc/aws
access_key  secret_key

/ # cat /etc/aws/access_key
AKIAWS2...

/ # cat /etc/aws/secret_key
9Qa5WP.....
```

So, we can see that the aws IAM credentials `access_key` and  `secret_key` are mounted into the pod.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted

$ kubectl delete ns trial
namespace "trial" deleted
```
