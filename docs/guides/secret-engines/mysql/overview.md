---
title: Manage MySQL credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-mysql
    name: Overview
    parent: mysql-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage MySQL credentials using the KubeVault operator

MySQL is one of the supported plugins for the database secrets engine. This plugin generates database credentials dynamically based on configured roles for the MySQL database. You can easily manage [MySQL secret engine](https://www.vaultproject.io/docs/secrets/databases/mysql-maria.html) using the KubeVault operator.

![MySQL secret engine](/docs/images/guides/secret-engines/mysql/mysql_secret_engine_guide.svg)

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [MySQLRole](/docs/concepts/secret-engine-crds/database-secret-engine/mysql.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we are going to create a [role](https://www.vaultproject.io/docs/secrets/databases/mysql-maria#setup) using MySQL and issue credential using SecretAccessRequest.

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

## Enable and Configure MySQL Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on specified path and configure the secret engine with given configurations.

A sample SecretEngine object for the MySQL secret engine:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: mysql-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  mysql:
    databaseRef:
      name: mysql
      namespace: demo
    pluginName: "mysql-database-plugin"
```

Let's deploy SecretEngine:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mysql/secretengine.yaml
secretengine.engine.kubevault.com/mysql-engine created
```

Wait till the status become `Success`:

```console
$ kubectl get secretengines -n demo
NAME           STATUS    AGE
mysql-engine   Success   10s
```

Since the status is `Success`, the MySQL secret engine is enabled and successfully configured. You can use `kubectl describe secretengine -n <namepsace> <name>` to check for error events, if any.

## Create MySQL Role

By using [MySQLRole](/docs/concepts/secret-engine-crds/database-secret-engine/mysql.md), you can create a [role](https://www.vaultproject.io/docs/secrets/databases/mysql-maria#setup) on the Vault server in Kubernetes native way.

A sample MySQLRole object is given below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: MySQLRole
metadata:
  name: mysql-superuser-role
  namespace: demo
spec:
  secretEngineRef:
    name: sql-secrt-engine
  creationStatements:
    - "CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';"
    - "GRANT SELECT ON *.* TO '{{name}}'@'%';"
  defaultTTL: 1h
  maxTTL: 24h
```

Let's deploy MySQLRole:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mysql/secretenginerole.yaml
mongodbrole.engine.kubevault.com/mysql-superuser-role created

$ kubectl get mysqlrole -n demo
NAME                   STATUS    AGE
mysql-superuser-role   Success   34m
```

You can also check from Vault that the role is created.
To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

```console
$ vault list your-database-path/roles
Keys
----
k8s.-.demo.mysql-superuser-role

$ vault read your-database-path/roles/k8s.-.demo.mysql-superuser-role
Key                      Value
---                      -----
creation_statements      [CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}'; GRANT SELECT ON *.* TO '{{name}}'@'%';]
db_name                  k8s.-.demo.mysql
default_ttl              1h
max_ttl                  24h
renew_statements         []
revocation_statements    []
rollback_statements      []
```

If we delete the MySQL, then the respective role will be deleted from the Vault.

```console
$ kubectl delete mysqlrole -n demo mysql-superuser-role
mysqlrole.engine.kubevault.com "mysql-superuser-role" deleted
```

Check from Vault whether the role exists:

```console
$ vault read your-database-path/roles/k8s.-.demo.mysql-superuser-role
No value found at your-database-path/roles/k8s.-.demo.mysql-superuser-role

$ vault list your-database-path/roles
No value found at your-database-path/roles/
```

## Generate MySQL credentials

Here, we are going to make a request to Vault for MySQL credentials by creating `mysql-cred-rqst` SecretAccessRequest in `demo` namespace.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: mysql-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: MySQLRole
    name: mysql-superuser-role
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Here, `spec.roleRef` is the reference of MySQL against which credentials will be issued. `spec.subjects` is the reference to the object or user identities a role binding applies to it will have read access of the credential secret.

Now, we are going to create SecretAccessRequest.

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mysql/mysqlaccessrequest.yaml
secretaccessrequest.engine.kubevault.com/mysql-cred-rqst created

$ kubectl get secretaccessrequest -n demo
NAME              AGE
mysql-cred-rqst   72m
```

Database credentials will not be issued until it is approved. The KubeVault operator will watch for the approval in the `status.conditions[].type` field of the request object. You can use [KubeVault CLI](https://github.com/kubevault/cli), a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), to approve or deny SecretAccessRequest.

```console
# using KubeVault CLI as kubectl plugin to approve request
$ kubectl vault approve secretaccessrequest mysql-cred-rqst -n demo
approved

$ kubectl get secretaccessrequest -n demo mysql-cred-rqst -o yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: mysql-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: MySQLRole
    name: mysql-superuser-role
    namespace: demo
  subjects:
  - kind: ServiceAccount
    name: demo-sa
    namespace: demo
status:
  conditions:
  - lastUpdateTime: "2020-11-18T06:41:57Z"
    message: This was approved by kubectl vault approve secretaccessrequest
    reason: KubectlApprove
    type: Approved
  lease:
    duration: 1h0m0s
    id: your-database-path/creds/k8s.-.demo.mysql-superuser-role/ni3TCo2HkSwCUb8kmQuvIDdx
    renewable: true
  secret:
    name: mysql-cred-rqst-gy66wq
```

Once SecretAccessRequest is approved, the KubeVault operator will issue credentials from Vault and create a secret containing the credential. It will also create a role and rolebinding so that `spec.subjects` can access secret. You can view the information in the `status` field.

```console
$ kubectl get secretaccessrequest mysql-cred-rqst -n demo -o json | jq '.status'
{
  "conditions": [
    {
      "lastUpdateTime": "2019-11-18T06:41:57Z",
      "message": "This was approved by kubectl vault approve secretaccessrequest",
      "reason": "KubectlApprove",
      "type": "Approved"
    }
  ],
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.mysql-superuser-role/ni3TCo2HkSwCUb8kmQuvIDdx",
    "renewable": true
  },
  "secret": {
    "name": "mysql-cred-rqst-gy66wq"
  }
}

$ kubectl get secret -n demo mysql-cred-rqst-gy66wq -o yaml
apiVersion: v1
data:
  password: QTFhLVBkZGlsZFFxa0o1cnlvR20=
  username: di1rdWJlcm5ldGVzLWRlbW8TE1NzQwNTkzMTc=
kind: Secret
metadata:
  name: mysql-cred-rqst-gy66wq
  namespace: demo
  ownerReferences:
  - apiVersion: engine.kubevault.com/v1alpha1
    controller: true
    kind: SecretAccessRequest
    name: mysql-cred-rqst
    uid: 54ce63ca-d0e7-4b97-9085-b52eb3cb334f
type: Opaque
```

If SecretAccessRequest is deleted, then credential lease (if any) will be revoked.

```console
$ kubectl delete secretaccessrequest -n demo mysql-cred-rqst
secretaccessrequest.engine.kubevault.com "mysql-cred-rqst" deleted
```

If SecretAccessRequest is `Denied`, then the KubeVault operator will not issue any credential.

```console
$ kubectl vault deny secretaccessrequest mysql-cred-rqst -n demo
  Denied
```

> Note: Once SecretAccessRequest is `Approved`, you cannot change `spec.roleRef` and `spec.subjects` field.
