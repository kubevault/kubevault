# Deploying Vault with Amazon EKS using AWS S3 backend and unsealing it using awsKmsSsm

Here, we are going to deploy Vault in Amazon EKS using Vault operator. We are going to use [AWS S3 bucket](https://aws.amazon.com/s3/) as Vault backend and `awsKmsSsm` unsealer mode for automatic unsealing the Vault. 

## Before You Begin 

At first, you need to have a EKS cluster. If you don't already have a cluster, create one from [here](https://aws.amazon.com/eks/). You can use [eksctl](https://github.com/weaveworks/eksctl) command line tool to create EKS cluster easily.

- Install Vault operator in your cluster following the steps [here](/docs/setup/operator/install.md).

- You should be familiar with the following CRD:
  - [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md)
  - [Unsealer](/docs/concepts/vault-server-crds/unsealer/unsealer.md)
  - [awsKmsSsm](/docs/concepts/vault-server-crds/unsealer/aws_kms_ssm.md)

- You will need a [AWS S3 Bucket](https://aws.amazon.com/s3/) to use it as Vault backend storage. In this tutorial, we are going to use `demo-vault-3` S3 bucket.

- You will need a [AWS KMS key](https://aws.amazon.com/kms/) to use it for Vault unsealer. In this tutorial, we are going to use `218daa5f-7173-429e-a030-288b30761f79` as KMS key id. 

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

### Provision Cluster

We are going to use [eksctl](https://github.com/weaveworks/eksctl) to provision a cluster.

```console
eksctl create cluster --name demo-cluster --nodes 1 --region us-east-1 --version 1.11
```

![aws ec2 instance](/docs/images/guides/provider/eks/aws-instance.png)

### Install Vault operator

See [here](/docs/setup/operator/install.md).

```console
$ kubectl get pods -n kube-system
NAME                             READY     STATUS    RESTARTS   AGE
vault-operator-798b75d78-qw74f   1/1       Running   1          2h
```

### Deploy Vault

We will deploy `my-vault` on `demo` namespace. We will configure it for S3 backend. We will use `awsKmsSsm` for auto initializing and unsealing. We already created a S3 bucket `demo-vault-3` in `us-east-1` region.

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: my-vault
  namespace: demo
spec:
  nodes: 1
  version: "0.11.1"
  backend:
    s3:
      bucket: "demo-vault-3"
      region: "us-east-1"
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      awsKmsSsm:
        region: "us-east-1"
        kmsKeyID: "218daa5f-7173-429e-a030-288b30761f79"
```

Here, `spec.version` specifies the name of the [VaultServerVersion](docs/concepts/vault-server-crds/vaultserverversion.md) CRD. If that does not exist, then create one.

```console
$ kubectl get vaultserverversions
NAME      VERSION   VAULT_IMAGE    DEPRECATED   AGE
0.11.1    0.11.1    vault:0.11.1   false        12m

$ kubectl get vaultserverversions/0.11.1 -o yaml
apiVersion: catalog.kubevault.com/v1alpha1
kind: VaultServerVersion
metadata:
  labels:
    app: vault-operator
  name: 0.11.1
spec:
  deprecated: false
  exporter:
    image: kubevault/vault-exporter:canary
  unsealer:
    image: kubevault/vault-unsealer:0.1.0
  vault:
    image: vault:0.11.1
  version: 0.11.1
```

Now, we are going to create `my-vault`

```console
$ cat examples/guides/provider/eks/my-vault.yaml 
  apiVersion: kubevault.com/v1alpha1
  kind: VaultServer
  metadata:
    name: my-vault
    namespace: demo
  spec:
    nodes: 1
    version: "0.11.1"
    backend:
      s3:
        bucket: "demo-vault-3"
        region: "us-east-1"
    unsealer:
      secretShares: 4
      secretThreshold: 2
      mode:
        awsKmsSsm:
          region: "us-east-1"
          kmsKeyID: "218daa5f-7173-429e-a030-288b30761f79"

$ kubectl apply -f examples/guides/provider/eks/my-vault.yaml 
vaultserver.kubevault.com/my-vault created
```

> **Note**: Here, vault will attempt to retrieve credentials from the AWS metadata service. Please, make sure that it's has permission for s3 bucket, encryption key and amazon ssm. Also, you can specify dedicated credential for this using `s3.credentialSecret` and `awsKmsSsm.credentialSecret`. AWS policy are given at bottom of this tutorial.

Check the `my-vault` status. It may take some time to reach `Running` stage.

```console
$ kubectl get vaultserver/my-vault -n demo
NAME       NODES     VERSION   STATUS    AGE
my-vault   1         0.11.1    Running   3m
```

`status` field in `my-vault` will show more detail information.

```console
$ kubectl get vaultserver/my-vault -n demo -o json | jq '.status'
{
  "initialized": true,
  "observedGeneration": "1$6208915667192219204",
  "phase": "Running",
  "updatedNodes": [
    "my-vault-6f48b4d96f-mzvgm"
  ],
  "vaultStatus": {
    "active": "my-vault-6f48b4d96f-mzvgm",
    "unsealed": [
      "my-vault-6f48b4d96f-mzvgm"
    ]
  }
}

```

Vault operator will create a service `{metadata.name}` for `my-vault` in the same namespace. For this case, service name is `my-vault`. You can specify service configuration in [spec.serviceTemplate](/docs/concepts/vault-server-crds/vaultserver.md#specservicetemplate). Vault operator will use that configuration to create service.

```console
$ kubectl get services -n demo
NAME       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
my-vault   ClusterIP   10.100.237.152   <none>        8200/TCP,8201/TCP,9102/TCP   46m
```

The configuration used to run Vault can be found in `{metadata.name}-vault-config` configMap. For this case, it is `my-vault-vault-config`. Confidential data are omitted in this configMap.

```console
$ kubectl get configmaps -n demo
NAME                    DATA      AGE
my-vault-vault-config   1         49m

$ kubectl get configmaps/my-vault-vault-config -n demo -o yaml
apiVersion: v1
data:
  vault.hcl: |2-

    listener "tcp" {
      address = "0.0.0.0:8200"
      cluster_address = "0.0.0.0:8201"
      tls_cert_file = "/etc/vault/tls/server.crt"
      tls_key_file  = "/etc/vault/tls/server.key"
    }

    storage "s3" {
    bucket = "demo-vault-3"
    region = "us-east-1"
    }

    telemetry {
      statsd_address = "0.0.0.0:9125"
    }
kind: ConfigMap
metadata:
  creationTimestamp: 2018-12-22T04:30:07Z
  labels:
    app: vault
    vault_cluster: my-vault
  name: my-vault-vault-config
  namespace: demo

```

In this `my-vault`, Vault operator will use self-signed certificates for Vault and also will create `{metadata.name}-vault-tls` secret containing certificates. You can optionally specify certificates in [spec.tls](/docs/concepts/vault-server-crds/vaultserver.md#spectls). 

```console
$ kubectl get secrets -n demo
NAME                                      TYPE                                  DATA      AGE
my-vault-vault-tls                        Opaque                                3         1h
```

We can see unseal keys and root token in AWS System Manager Parameter Store in the `unsealer.region` region. For this case, in `us-east-1` region.

![unseal keys](/docs/images/guides/provider/eks/unseal-keys.png)

### Using Vault

Download and decrypt the root token:
```console
$ aws ssm get-parameter --name vault-root-token --region us-east-1 --output json | jq -r '.Parameter.Value' | base64 -d - > root.enc

$ tree .
.
└── root.enc

$ aws kms decrypt --ciphertext-blob fileb://root.enc --output text --query Plaintext --encryption-context "Tool=vault-unsealer" --region us-east-1 | base64 -d -
9116f849-2085-9c28-015f-aec3e184e90f
```

> Note: Make sure you have the permission to do above operation. Also we highly recommend not to use root token for using vault.

For testing purpose, we are going to port forward the active vault pod, since the service we exposed for Vault is ClusterIP type. Make sure Vault cli is installed.

```console
$ kubectl port-forward my-vault-6f48b4d96f-mzvgm -n demo 8200:8200
Forwarding from 127.0.0.1:8200 -> 8200

# run following commands on another terminal
$ export VAULT_SKIP_VERIFY="true"

$ export VAULT_ADDR='https://127.0.0.1:8200'

$ vault status
Key             Value
---             -----
Seal Type       shamir
Sealed          false
Total Shares    4
Threshold       2
Version         0.11.1
Cluster Name    vault-cluster-e4eda2ce
Cluster ID      d05fec0c-7e09-20f6-0d88-0283ed9c7b72
HA Enabled      false

```

Set Vault token for further use. In this case, we are going to use root token(not recommended).  

```console
$ export VAULT_TOKEN='9116f849-2085-9c28-015f-aec3e184e90f'

$ vault secrets list
Path          Type         Accessor              Description
----          ----         --------              -----------
cubbyhole/    cubbyhole    cubbyhole_9ce16bb9    per-token private secret storage
identity/     identity     identity_45904875     identity store
secret/       kv           kv_22970276           key/value secret storage
sys/          system       system_51cd4d05       system endpoints used for control, policy and debugging

```

We are going to write,read and delete a secret in Vault

```console
$ vault kv put secret/foo A=B
Success! Data written to: secret/foo

# see written secret data
$ vault kv get secret/foo
== Data ==
Key    Value
---    -----
A      B

# delete the secret
$ vault kv delete secret/foo
Success! Data deleted (if it existed) at: secret/foo

# check the secret whether it is exist or not
$ vault kv get secret/foo
No value found at secret/foo

```

## AWS IAM Policy

Policy for S3 bucket access:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VaultListBuckets",
            "Effect": "Allow",
            "Action": [
                "s3:ListAllMyBuckets",
                "s3:HeadBucket"
            ],
            "Resource": "*"
        },
        {
            "Sid": "VaultAccessBuckets",
            "Effect": "Allow",
            "Action": "s3:*",
            "Resource": [
                "arn:aws:s3:::<s3-bucket-name>",
                "arn:aws:s3:::<s3-bucket-name>/*"
            ]
        }
    ]
}
```

Policy for KMS:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VaultUnsealerEncryptDecryptKms",
            "Effect": "Allow",
            "Action": [
                "kms:Decrypt",
                "kms:Encrypt",
                "kms:DescribeKey"
            ],
            "Resource": "arn:aws:kms:<region>:<aws-account-id>:key/<key-uuid>"
        },
        {
            "Sid": "VaultUnsealerGetKMS",
            "Effect": "Allow",
            "Action": "kms:ListKeys",
            "Resource": "*"
        }
    ]
}
```

Policy for SSM:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VaultUnsealerParametersAccess",
            "Effect": "Allow",
            "Action": [
                "ssm:PutParameter",
                "ssm:DeleteParameter",
                "ssm:GetParameters"
            ],
            "Resource": "arn:aws:ssm:*:*:parameter/*"
        }
    ]
}
```
