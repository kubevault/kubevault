---
title: Manage Redis credentials using the KubeVault operator
menu:
    docs_{{ .version }}:
        identifier: overview-redis
        name: Overview
        parent: redis-secret-engines
        weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Redis credentials using the KubeVault operator

Redis is one of the supported plugins for the database secrets engine. This plugin generates database credentials dynamically based on configured roles for the Redis database. You can easily manage [Redis secret engine](https://www.vaultproject.io/docs/secrets/databases/redis.html) using the KubeVault operator.

![Redis secret engine](/docs/images/guides/secret-engines/redis/redis_secret_engine_guide.svg)

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [RedisRole](/docs/concepts/secret-engine-crds/database-secret-engine/redis.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we are going to create a [role](https://developer.hashicorp.com/vault/docs/secrets/databases/redis#setup) using Redis and issue credential using SecretAccessRequest.

## Vault Server

If you don't have a Vault Server, you can deploy it by using the KubeVault operator. To create a Redis Secret Engine, VaultServer version needs to be 1.12.1+

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

## Enable and Configure Redis Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on specified path and configure the secret engine with given configurations.

A sample SecretEngine object for the Redis secret engine:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: redis-secret-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
    namespace: demo
  redis:
    databaseRef:
      name: redis-standalone
      namespace: demo
    pluginName: "redis-database-plugin"

```

Let's deploy SecretEngine:

```bash
$ kubectl apply -f docs/examples/guides/secret-engines/redis/secretengine.yaml
secretengine.engine.kubevault.com/redis-secret-engine created
```

Wait till the status become `Success`:

```bash
$ kubectl get secretengines -n demo
NAME           STATUS    AGE
redis-secret-engine   Success   10s
```

Since the status is `Success`, the Redis secret engine is enabled and successfully configured. You can use `kubectl describe secretengine -n <namepsace> <name>` to check for error events, if any.

## Create Redis Role

By using [RedisRole](/docs/concepts/secret-engine-crds/database-secret-engine/redis.md), you can create a [role](https://developer.hashicorp.com/vault/docs/secrets/databases/redis#setup) on the Vault server in Kubernetes native way.

A sample RedisRole object is given below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: RedisRole
metadata:
  name: write-read-role
  namespace: demo
spec:
  secretEngineRef:
    name: redis-secret-engine
  creationStatements:
    - '["~*", "+@read","+@write"]'
  defaultTTL: 1h
  maxTTL: 24h
```

Let's deploy RedisRole:

```bash
$ kubectl apply -f docs/examples/guides/secret-engines/redis/secretenginerole.yaml
redisrole.engine.kubevault.com/write-read-role created

$ kubectl get redisrole -n demo
NAME                   STATUS    AGE
write-read-role        Success   34m
```

You can also check from Vault that the role is created.
To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

```bash
$ vault list k8s.kubevault.com.redis.demo.redis-secret-engine/roles
Keys
----
k8s.kubevault.com.demo.write-read-role

$ vault read k8s.kubevault.com.redis.demo.redis-secret-engine/roles/k8s.kubevault.com.demo.write-read-role
Key                      Value
---                      -----
creation_statements      [["~*", "+@read","+@write"]]
credential_type          password
db_name                  k8s.kubevault.com.demo.redis-standalone
default_ttl              1h
max_ttl                  24h
renew_statements         []
revocation_statements    []
rollback_statements      []
```

If we delete the Redis, then the respective role will be deleted from the Vault.

```bash
$ kubectl delete -n demo redisrole write-read-role
redisrole.engine.kubevault.com "write-read-role" deleted
```

Check from Vault whether the role exists:

```bash
$  vault read k8s.kubevault.com.redis.demo.redis-secret-engine/roles/k8s.kubevault.com.demo.write-read-role
No value found at k8s.kubevault.com.redis.demo.redis-secret-engine/roles/k8s.kubevault.com.demo.write-read-role

$ vault list k8s.kubevault.com.redis.demo.redis-secret-engine/roles
No value found at k8s.kubevault.com.redis.demo.redis-secret-engine/roles
```

## Generate Redis credentials

Here, we are going to make a request to Vault for Redis credentials by creating `write-read-access-req` SecretAccessRequest in `demo` namespace.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: write-read-access-req
  namespace: demo
spec:
  roleRef:
    kind: RedisRole
    name: write-read-role
  subjects:
    - kind: ServiceAccount
      name: write-read-user
      namespace: demo

```

Here, `spec.roleRef` is the reference of Redis against which credentials will be issued. `spec.subjects` is the reference to the object or user identities a role binding applies to it will have read access of the credential secret.

Now, we are going to create SecretAccessRequest.

```bash
$ kubectl apply -f docs/examples/guides/secret-engines/redis/redisaccessrequest.yaml
secretaccessrequest.engine.kubevault.com/write-read-access-req created

$ kubectl get secretaccessrequest -n demo
NAME                    STATUS               AGE
write-read-access-req   WaitingForApproval   14s
```

Database credentials will not be issued until it is approved. The KubeVault operator will watch for the approval in the `status.conditions[].type` field of the request object. You can use [KubeVault CLI](https://github.com/kubevault/cli), a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), to approve or deny SecretAccessRequest.

```bash
# using KubeVault CLI as kubectl plugin to approve request
$ kubectl vault approve secretaccessrequest write-read-access-req -n demo
secretaccessrequests write-read-access-req approved

$ kubectl get secretaccessrequest -n demo write-read-access-req -o yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"engine.kubevault.com/v1alpha1","kind":"SecretAccessRequest","metadata":{"annotations":{},"name":"write-read-access-req","namespace":"demo"},"spec":{"roleRef":{"kind":"RedisRole","name":"write-read-role"},"subjects":[{"kind":"ServiceAccount","name":"write-read-user","namespace":"demo"}]}}
    vaultservers.kubevault.com/name: vault
    vaultservers.kubevault.com/namespace: demo
  creationTimestamp: "2022-12-28T09:14:25Z"
  finalizers:
  - kubevault.com
  generation: 1
  name: write-read-access-req
  namespace: demo
  resourceVersion: "341401"
  uid: 0bf92c6a-fbbb-4600-8bc8-8bddbf2c34dd
spec:
  roleRef:
    kind: RedisRole
    name: write-read-role
  subjects:
  - kind: ServiceAccount
    name: write-read-user
    namespace: demo
status:
  conditions:
  - lastTransitionTime: "2022-12-28T09:15:22Z"
    message: 'This was approved by: kubectl vault approve secretaccessrequest'
    observedGeneration: 1
    reason: KubectlApprove
    status: "True"
    type: Approved
  - lastTransitionTime: "2022-12-28T09:15:22Z"
    message: The requested credentials successfully issued.
    observedGeneration: 1
    reason: SuccessfullyIssuedCredential
    status: "True"
    type: Available
  lease:
    duration: 1h0m0s
    id: k8s.kubevault.com.redis.demo.redis-secret-engine/creds/k8s.kubevault.com.demo.write-read-role/mNeREfw0SJQBekA8ZkzJn2Tf
    renewable: true
  observedGeneration: 1
  phase: Approved
  secret:
    name: write-read-access-req-c9ttdf
    namespace: demo
```

Once SecretAccessRequest is approved, the KubeVault operator will issue credentials from Vault and create a secret containing the credential. It will also create a role and rolebinding so that `spec.subjects` can access secret. You can view the information in the `status` field.

```bash
$ kubectl get secretaccessrequest write-read-access-req -n demo -o json | jq '.status'
{
  "conditions": [
    {
      "lastTransitionTime": "2022-12-28T09:15:22Z",
      "message": "This was approved by: kubectl vault approve secretaccessrequest",
      "observedGeneration": 1,
      "reason": "KubectlApprove",
      "status": "True",
      "type": "Approved"
    },
    {
      "lastTransitionTime": "2022-12-28T09:15:22Z",
      "message": "The requested credentials successfully issued.",
      "observedGeneration": 1,
      "reason": "SuccessfullyIssuedCredential",
      "status": "True",
      "type": "Available"
    }
  ],
  "lease": {
    "duration": "1h0m0s",
    "id": "k8s.kubevault.com.redis.demo.redis-secret-engine/creds/k8s.kubevault.com.demo.write-read-role/mNeREfw0SJQBekA8ZkzJn2Tf",
    "renewable": true
  },
  "observedGeneration": 1,
  "phase": "Approved",
  "secret": {
    "name": "write-read-access-req-c9ttdf",
    "namespace": "demo"
  }
}

$ kubectl get secret -n demo
NAME                                   TYPE                                  DATA   AGE
write-read-access-req-c9ttdf           Opaque                                2      2m3s

$ kubectl get secret -n demo write-read-access-req-c9ttdf -o yaml
apiVersion: v1
data:
  password: MUtwT2YtV0lyZG1qSTJQUktwSFg=
  username: Vl9LVUJFUk5FVEVTLURFTU8tVkFVTFRfSzhTLktVQkVWQVVMVC5DT00uREVNTy5XUklURS1SRUFELVJPTEVfQVlCVFBLTVhGUEdPM0tHR05NUjJfMTY3MjIxODkyMg==
kind: Secret
metadata:
  creationTimestamp: "2022-12-28T09:15:22Z"
  name: write-read-access-req-c9ttdf
  namespace: demo
  ownerReferences:
  - apiVersion: engine.kubevault.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: SecretAccessRequest
    name: write-read-access-req
    uid: 0bf92c6a-fbbb-4600-8bc8-8bddbf2c34dd
  resourceVersion: "341397"
  uid: b038419d-59ff-4946-8ff3-1a04984d6f0f
type: Opaque
```

If SecretAccessRequest is deleted, then credential lease (if any) will be revoked.

```bash
$ kubectl delete secretaccessrequest -n demo write-read-access-req
secretaccessrequest.engine.kubevault.com "write-read-access-req" deleted
```

If SecretAccessRequest is `Denied`, then the KubeVault operator will not issue any credential.

```bash
$ kubectl vault deny secretaccessrequest write-read-access-req -n demo
secretaccessrequest.engine.kubevault.com "write-read-access-req" deleted
```

> Note: Once SecretAccessRequest is `Approved`, you cannot change `spec.roleRef` and `spec.subjects` field.
