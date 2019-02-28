---
title: Manage MongoDB credentials using the Vault Operator
menu:
  docs_0.1.0:
    identifier: overview-mongodb
    name: Overview
    parent: mongodb-secret-engines
    weight: 10
menu_name: docs_0.1.0
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage MongoDB credentials using the Vault Operator

You can easily manage [MongoDB Database secret engine](https://www.vaultproject.io/api/secret/databases/mongodb.html) using Vault operator.

You should be familiar with the following CRD:

- [MongoDBRole](/docs/concepts/database-crds/mongodb.md)
- [DatabaseAccessRequest](/docs/concepts/database-crds/databaseaccessrequest.md)
- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)

Before you begin:

- Install Vault operator in your cluster following the steps [here](/docs/setup/operator/install).

- Deploy Vault. It could be in the Kubernetes cluster or external.


To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we will create [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role) using MongoDBRole and issue credential using DatabaseAccessRequest. For this tutorial, we are going to deploy Vault using Vault operator.

```console
$ cat examples/guides/secret-engins/mongodb/vault.yaml 
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  nodes: 1
  version: "1.0.0"
  backend:
    inmem: {}
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      kubernetesSecret:
        secretName: vault-keys

$ kubectl get vaultserverversions/1.0.0 -o yaml
apiVersion: catalog.kubevault.com/v1alpha1
kind: VaultServerVersion
metadata:
  name: 1.0.0
spec:
  exporter:
    image: kubevault/vault-exporter:0.1.0
  unsealer:
    image: kubevault/vault-unsealer:0.1.0
  vault:
    image: vault:1.0.0
  version: 1.0.0

$ kubectl apply -f examples/guides/secret-engins/mongodb/vault.yaml 
vaultserver.kubevault.com/vault created

$ kubectl get vaultserver/vault -n demo
NAME      NODES     VERSION   STATUS    AGE
vault     1         1.0.0     Running   1h
```

## MongoDBRole

Using [MongoDBRole](/docs/concepts/database-crds/mongodb.md), you can configure [connection](https://www.vaultproject.io/api/secret/databases/mongodb.html#configure-connection) and create [role](https://www.vaultproject.io/api/secret/databases/index.html#create-role). In this tutorial, we are going to create `demo-role` in `demo` namespace.

```yaml
apiVersion: authorization.kubedb.com/v1alpha1
kind: MongoDBRole
metadata:
  name: demo-role
  namespace: demo
spec:
  creationStatements:
    - "{ \"db\": \"admin\", \"roles\": [{ \"role\": \"readWrite\" }, {\"role\": \"read\", \"db\": \"foo\"}] }"
  defaultTTL: 1h
  maxTTL: 24h
  authManagerRef:
    name: vault-app
    namespace: demo
  databaseRef:
    name: mongo-app
```

Here, `spec.databaseRef` is the reference of AppBinding containing MongoDB database connection and credential information.  

```yaml
$ cat examples/guides/secret-engins/mongodb/mongo-app.yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: mongo-app
  namespace: demo
spec:
  secret:
    name: mongo-user-cred # secret
  clientConfig:
    service:
      name: mongo
      scheme: mongodb
      port: 27017
    insecureSkipTLSVerify: true
  parameters:
    allowedRoles: "*" # names of the allowed roles to use this connection config in Vault, ref: https://www.vaultproject.io/api/secret/databases/index.html#allowed_roles

$ kubectl apply -f examples/guides/secret-engins/mongodb/mongo-app.yaml
+appbinding.appcatalog.appscode.com/mongo-app created
```

`spec.authManagerRef` is the reference of AppBinding containing Vault connection and credential information. See [here](/docs/concepts/vault-server-crds/auth-methods/overview) for Vault authentication using AppBinding in Vault operator.

```yaml
$ cat examples/guides/secret-engins/mongodb/vault-app.yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9ERXlNamN3TkRVNU1qVmFGdzB5T0RFeU1qUXdORFU1TWpWYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMVhid2wyQ1NNc2VQTU5RRzhMd3dUVWVOCkI1T05oSTlDNzFtdUoyZEZjTTlUc1VDQnlRRk1weUc5dWFvV3J1ZDhtSWpwMVl3MmVIUW5udmoybXRmWGcrWFcKSThCYkJUaUFKMWxMMFE5MlV0a1BLczlXWEt6dTN0SjJUR1hRRDhhbHZhZ0JrR1ViOFJYaUNqK2pnc1p6TDRvQQpNRWszSU9jS0xnMm9ldFZNQ0hwNktpWTBnQkZiUWdJZ1A1TnFwbksrbU02ZTc1ZW5hWEdBK2V1d09FT0YwV0Z2CmxGQmgzSEY5QlBGdTJKbkZQUlpHVDJKajBRR1FNeUxodEY5Tk1pZTdkQnhiTWhRVitvUXp2d1EvaXk1Q2pndXQKeDc3d29HQ2JtM0o4cXRybUg2Tjl6Tlc3WlR0YTdLd05PTmFoSUFEMSsrQm5rc3JvYi9BYWRKT0tMN2dLYndJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFXeWFsdUt3Wk1COWtZOEU5WkdJcHJkZFQyZnFTd0lEOUQzVjN5anBlaDVCOUZHN1UKSS8wNmpuRVcyaWpESXNHNkFDZzJKOXdyaSttZ2VIa2Y2WFFNWjFwZHRWeDZLVWplWTVnZStzcGdCRTEyR2NPdwpxMUhJb0NrekVBMk5HOGRNRGM4dkQ5WHBQWGwxdW5veWN4Y0VMeFVRSC9PRlc4eHJxNU9vcXVYUkxMMnlKcXNGCmlvM2lJV3EvU09Yajc4MVp6MW5BV1JSNCtSYW1KWjlOcUNjb1Z3b3R6VzI1UWJKWWJ3QzJOSkNENEFwOUtXUjUKU2w2blk3NVMybEdSRENsQkNnN2VRdzcwU25seW5mb3RaTUpKdmFzbStrOWR3U0xtSDh2RDNMMGNGOW5SOENTSgpiTjBiZzczeVlWRHgyY3JRYk0zcko4dUJnY3BsWlRpUy91SXJ2QT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    service:
      name: vault
      port: 8200
      scheme: HTTPS
  parameters:
    serviceAccountName: demo-sa
    policyControllerRole: mongodb-role
    authPath: kubernetes

$ kubectl apply -f examples/guides/secret-engins/mongodb/vault-app.yaml
appbinding.appcatalog.appscode.com/vault-app created
```

`demo-sa` serviceaccount in the above AppBinding has the following permission in Vault.

To create `demo-sa` serviceaccount run the following command:

```console
$ kubectl create serviceaccount -n demo demo-sa
serviceaccount/demo-sa created
```

Now you need to create policy with following capabilities, which will be assigned to a role.

```hcl
path "sys/mounts" {
  capabilities = ["read", "list"]
}

path "sys/mounts/*" {
  capabilities = ["create", "read", "update", "delete"]
}

path "database/config/*" {
	capabilities = ["create", "read", "update", "delete"]
}

path "database/roles/*" {
	capabilities = ["create", "update", "read", "delete"]
}

path "database/creds/*" {
    capabilities = ["read"]
}

path "sys/leases/revoke/*" {
    capabilities = ["update"]
}
```

You can manage policy in Vault using Vault operator, see [here](/docs/guides/policy-management/policy-management).

To create above policy run following command

```console
$ kubectl apply -f examples/guides/secret-engins/mysql/policy.yaml
vaultpolicy.policy.kubevault.com/mysql-role-policy created
vaultpolicybinding.policy.kubevault.com/mysql-role created
```

Now, we are going to create `demo-role`.

```console
$ cat examples/guides/secret-engins/mongodb/demo-role.yaml
apiVersion: authorization.kubedb.com/v1alpha1
kind: MongoDBRole
metadata:
  name: demo-role
  namespace: demo
spec:
  creationStatements:
    - "{ \"db\": \"admin\", \"roles\": [{ \"role\": \"readWrite\" }, {\"role\": \"read\", \"db\": \"foo\"}] }"
  defaultTTL: 1h
  maxTTL: 24h
  authManagerRef:
    name: vault-app
    namespace: demo
  databaseRef:
    name: mongo-app

$ kubectl apply -f examples/guides/secret-engins/mongodb/demo-role.yaml 
mongodbrole.authorization.kubedb.com/demo-role created
```

Check whether MongoDBRole is successful.

```console
$ kubectl get mongodbroles demo-role -n demo -o json | jq '.status'
{
  "observedGeneration": "1$6208915667192219204",
  "phase": "Success"
}
```

To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{spec.clusterName or -}.{spec.namespace}.{spec.name}`.

```console
$ vault list database/roles
Keys
----
k8s.-.demo.demo-role

$ vault read database/roles/k8s.-.demo.demo-role
Key                      Value
---                      -----
creation_statements      [{ "db": "admin", "roles": [{ "role": "readWrite" }, {"role": "read", "db": "foo"}] }]
db_name                  mongo-app
default_ttl              1h
max_ttl                  24h
renew_statements         <nil>
revocation_statements    <nil>
rollback_statements      <nil>
```
 
If we delete MongoDBRole, then respective role will be deleted from Vault.

```console
$ kubectl delete mongodbroles demo-role -n demo
mongodbrole.authorization.kubedb.com "demo-role" deleted

# check in vault whether role exists
$ vault read database/roles/k8s.-.demo.demo-role
Error reading database/roles/k8s.-.demo.demo-role: Error making API request.

URL: GET https://127.0.0.1:8200/v1/database/roles/k8s.-.demo.demo-role
Code: 400. Errors:

* Role 'k8s.-.demo.demo-role' not found

$ vault list database/roles
No value found at database/roles/
```

## DatabaseAccessRequest

Using [DatabaseAccessRequest](/docs/concepts/database-crds/databaseaccessrequest.md), you can issue MongoDB credential from Vault. In this tutorial, we are going to issue MongoDB credential by creating `demo-cred` DatabaseAccessRequest in `demo` namespace.

```yaml
apiVersion: authorization.kubedb.com/v1alpha1
kind: DatabaseAccessRequest
metadata:
  name: demo-cred
  namespace: demo
spec:
  roleRef:
    kind: MongoDBRole
    name: demo-role
    namespace: demo
  subjects:
    - kind: User
      name: nahid
      apiGroup: rbac.authorization.k8s.io
```

Here, `spec.roleRef` is the reference of MongoDBRole against which credential will be issued. `spec.subjects` is the reference to the object or user identities a role binding applies to and it will have read access of the credential secret. Also, Vault operator will use AppBinding reference from MongoDBRole which is specified in `spec.roleRef`. 

Now, we are going to create `demo-cred` DatabaseAccessRequest.

```console
$ kubectl apply -f examples/guides/secret-engins/mongodb/demo-cred.yaml 
databaseaccessrequest.authorization.kubedb.com/demo-cred created

$ kubectl get databaseaccessrequests -n demo
NAME        AGE
demo-cred   1m
```

MongoDB credential will not be issued until it is approved. To approve it, you have to add `Approved` in `status.conditions[].type` field. You can use [KubeVault CLI](https://github.com/kubevault/cli) as [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) to approve or deny DatabaseAccessRequest.

```console
# using kubevault cli as kubectl plugin to approve request
$ kubectl vault approve databaseaccessrequest demo-cred -n demo
approved

$ kubectl get databaseaccessrequest demo-cred -n demo -o yaml
apiVersion: authorization.kubedb.com/v1alpha1
kind: DatabaseAccessRequest
metadata:
  name: demo-cred
  namespace: demo
spec:
  roleRef:
    kind: MongoDBRole
    name: demo-role
    namespace: demo
  subjects:
  - apiGroup: rbac.authorization.k8s.io
    kind: User
    name: nahid
status:
  conditions:
  - lastUpdateTime: "2018-12-31T08:07:19Z"
    message: This was approved by kubectl vault approve databaseaccessrequest
    reason: KubectlApprove
    type: Approved
```

Once DatabaseAccessRequest is approved, Vault operator will issue credential from Vault and create a secret containing the credential. Also it will create rbac role and rolebinding so that `spec.subjects` can access secret. You can view the information in `status` field.

```console
$ kubectl get databaseaccessrequest demo-cred -n demo -o json | jq '.status'
{
  "conditions": [
    {
      "lastUpdateTime": "2018-12-31T10:58:43Z",
      "message": "This was approved by kubectl vault approve databaseaccessrequest",
      "reason": "KubectlApprove",
      "type": "Approved"
    }
  ],
  "lease": {
    "duration": "1h0m0s",
    "id": "database/creds/k8s.-.demo.demo-role/CXYWcdNzwVeZ3FqiOCyvl0Kc",
    "renewable": true
  },
  "secret": {
    "name": "demo-cred-f6kr6w"
  }
}

$ kubectl get secrets/demo-cred-f6kr6w -n demo -o yaml
apiVersion: v1
data:
  password: QTFhLThDVWl1SHpmT3h5YUJwVFQ=
  username: di1rdWJlcm5ldGVzLWRlbW8tazhzLi0uZGVtby5kZW1vLTFYbndHM1VXbUZrT01xQ0kwcmpILTE1NDYyNTM5MjQ=
kind: Secret
metadata:
  name: demo-cred-f6kr6w
  namespace: demo
type: Opaque
``` 

If DatabaseAccessRequest is deleted, then credential lease (if have any) will be revoked.

```console
$ kubectl delete databaseaccessrequest demo-cred -n demo
databaseaccessrequest.authorization.kubedb.com "demo-cred" deleted
```

If DatabaseAccessRequest is `Denied`, then Vault operator will not issue any credential. 

> Note: Once DatabaseAccessRequest is `Approved` or `Denied`, you can not change `spec.roleRef` and `spec.subjects` field.
