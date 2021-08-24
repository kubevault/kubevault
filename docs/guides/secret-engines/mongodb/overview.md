---
title: Manage MongoDB credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-mongodb
    name: Overview
    parent: mongodb-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage MongoDB credentials using the KubeVault operator

MongoDB is one of the supported plugins for the database secrets engine. This plugin generates database credentials dynamically based on configured roles for the MongoDB database. You can easily manage [MongoDB secret engine](https://www.vaultproject.io/docs/secrets/databases/mongodb.html) using the KubeVault operator.

![Elasticsearch secret engine](/docs/images/guides/secret-engines/mongodb/mongodb_secret_engine_guide.svg)

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [MongoDBRole](/docs/concepts/secret-engine-crds/database-secret-engine/mongodb.md)
- [DatabaseAccessRequest](/docs/concepts/secret-engine-crds/database-secret-engine/databaseaccessrequest.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we are going to create a [role](https://www.vaultproject.io/docs/secrets/databases/mongodb#setup) using MongoDB and issue credential using DatabaseAccessRequest.

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

## Enable and Configure Elasticsearch Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on specified path and configure the secret engine with given configurations.

A sample SecretEngine object for the Elasticsearch secret engine:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: mongo-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  mongodb:
    databaseRef:
      name: mongodb
      namespace: db
    pluginName: "mongodb-database-plugin"
  path: "your-database-path"
```

Let's deploy SecretEngine:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mongodb/secretengine.yaml
secretengine.engine.kubevault.com/mongo-engine created
```

Wait till the status become `Success`:

```console
$ kubectl get secretengines -n demo
NAME        STATUS    AGE
mongo-engine   Success   10s
```

Since the status is `Success`, the Elasticsearch secret engine is enabled and successfully configured. You can use `kubectl describe secretengine -n <namepsace> <name>` to check for error events, if any.

## Create MongoDB Role

By using [MongoDBRole](/docs/concepts/secret-engine-crds/database-secret-engine/mongodb.md), you can create a [role](https://www.vaultproject.io/docs/secrets/databases/mongodb#setup) on the Vault server in Kubernetes native way.

A sample MongoDBRole object is given below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: MongoDBRole
metadata:
  name: mongo-superuser-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  databaseRef:
    name: mongodb
    namespace: db
  path: "your-database-path"
  creationStatements:
    - "{ \"db\": \"admin\", \"roles\": [{ \"role\": \"readWrite\" }, {\"role\": \"read\", \"db\": \"foo\"}] }"
  defaultTTL: 1h
  maxTTL: 24h
```

Let's deploy MongoDBRole:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mongodb/secretenginerole.yaml
mongodbrole.engine.kubevault.com/mongo-superuser-role created

$ kubectl get mongodbrole -n demo
NAME                STATUS    AGE
mongo-superuser-role   Success   34m
```

You can also check from Vault that the role is created.
To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

```console
$ vault list your-database-path/roles
Keys
----
k8s.-.demo.mongo-superuser-role

$ vault read your-database-path/roles/k8s.-.demo.mongo-superuser-role
Key                      Value
---                      -----
creation_statements      [{ "db": "admin", "roles": [{ "role": "readWrite" }, {"role": "read", "db": "foo"}] }]
db_name                  k8s.-.db.mongodb
default_ttl              1h
max_ttl                  24h
renew_statements         []
revocation_statements    []
rollback_statements      []
```

If we delete the MongoDBRole, then the respective role will be deleted from the Vault.

```console
$ kubectl delete mongodbrole -n demo mongo-superuser-role
mongodbrole.engine.kubevault.com "mongo-superuser-role" deleted
```

Check from Vault whether the role exists:

```console
$ vault read your-database-path/roles/k8s.-.demo.mongo-superuser-role
No value found at your-database-path/roles/k8s.-.demo.mongo-superuser-role

$ vault list your-database-path/roles
No value found at your-database-path/roles/
```

## Generate MongoDB credentials

By using [DatabaseAccessRequest](/docs/concepts/secret-engine-crds/database-secret-engine/databaseaccessrequest.md), you can generate database access credentials from Vault.

Here, we are going to make a request to Vault for Elasticsearch credentials by creating `mongo-cred-rqst` DatabaseAccessRequest in `demo` namespace.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: DatabaseAccessRequest
metadata:
  name: mongo-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: MongoDBRole
    name: mongo-superuser-role
    namespace: demo
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Here, `spec.roleRef` is the reference of MongoDB against which credentials will be issued. `spec.subjects` is the reference to the object or user identities a role binding applies to it will have read access of the credential secret.

Now, we are going to create DatabaseAccessRequest.

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mongodb/mongodbaccessrequest.yaml
databaseaccessrequest.engine.kubevault.com/mongo-cred-rqst created

$ kubectl get databaseaccessrequest -n demo
NAME              AGE
mongo-cred-rqst   72m
```

Database credentials will not be issued until it is approved. The KubeVault operator will watch for the approval in the `status.conditions[].type` field of the request object. You can use [KubeVault CLI](https://github.com/kubevault/cli), a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), to approve or deny DatabaseAccessRequest.

```console
# using KubeVault CLI as kubectl plugin to approve request
$ kubectl vault approve databaseaccessrequest mongo-cred-rqst -n demo
approved

$ kubectl get databaseaccessrequest -n demo mongo-cred-rqst -o yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: DatabaseAccessRequest
metadata:
  name: mongo-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: MongoDBRole
    name: mongo-superuser-role
    namespace: demo
  subjects:
  - kind: ServiceAccount
    name: demo-sa
    namespace: demo
status:
  conditions:
  - lastUpdateTime: "2020-11-18T06:41:57Z"
    message: This was approved by kubectl vault approve databaseaccessrequest
    reason: KubectlApprove
    type: Approved
  lease:
    duration: 1h0m0s
    id: your-database-path/creds/k8s.-.demo.mongo-superuser-role/ni3TCo2HkSwCUb8kmQuvIDdx
    renewable: true
  secret:
    name: mongo-cred-rqst-gy66wq
```

Once DatabaseAccessRequest is approved, the KubeVault operator will issue credentials from Vault and create a secret containing the credential. It will also create a role and rolebinding so that `spec.subjects` can access secret. You can view the information in the `status` field.

```console
$ kubectl get databaseaccessrequest mongo-cred-rqst -n demo -o json | jq '.status'
{
  "conditions": [
    {
      "lastUpdateTime": "2019-11-18T06:41:57Z",
      "message": "This was approved by kubectl vault approve databaseaccessrequest",
      "reason": "KubectlApprove",
      "type": "Approved"
    }
  ],
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.mongo-superuser-role/ni3TCo2HkSwCUb8kmQuvIDdx",
    "renewable": true
  },
  "secret": {
    "name": "mongo-cred-rqst-gy66wq"
  }
}

$ kubectl get secret -n demo mongo-cred-rqst-gy66wq -o yaml
apiVersion: v1
data:
  password: QTFhLVBkZGlsZFFxa0o1cnlvR20=
  username: di1rdWJlcm5ldGVzLWRlbW8TE1NzQwNTkzMTc=
kind: Secret
metadata:
  name: mongo-cred-rqst-gy66wq
  namespace: demo
  ownerReferences:
  - apiVersion: engine.kubevault.com/v1alpha1
    controller: true
    kind: DatabaseAccessRequest
    name: mongo-cred-rqst
    uid: 54ce63ca-d0e7-4b97-9085-b52eb3cb334f
type: Opaque
```

If DatabaseAccessRequest is deleted, then credential lease (if any) will be revoked.

```console
$ kubectl delete databaseaccessrequest -n demo mongo-cred-rqst
databaseaccessrequest.engine.kubevault.com "mongo-cred-rqst" deleted
```

If DatabaseAccessRequest is `Denied`, then the KubeVault operator will not issue any credential.

```console
$ kubectl vault deny databaseaccessrequest mongo-cred-rqst -n demo
  Denied
```

> Note: Once DatabaseAccessRequest is `Approved` or `Denied`, you cannot change `spec.roleRef` and `spec.subjects` field.
