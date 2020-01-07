---
title: Manage PostgreSQL credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-postgres
    name: Overview
    parent: postgres-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage PostgreSQL Credentials Using the KubeVault Operator

PostgreSQL is one of the supported plugins for the database secrets engine. This plugin generates database credentials dynamically based on configured roles for the PostgreSQL database, and also supports Static Roles. You can easily manage [PostgreSQL Database secret engine](https://www.vaultproject.io/docs/secrets/databases/postgresql.html) using the KubeVault operator.

![PostgreSQL secret engine](/docs/images/guides/secret-engines/postgresql/postgres_secret_engine_guide.svg)

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [PostgresRole](/docs/concepts/secret-engine-crds/database-secret-engine/postgresrole.md)
- [DatabaseAccessRequest](/docs/concepts/secret-engine-crds/database-secret-engine/databaseaccessrequest.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/operator/install.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

In this tutorial, we are going to create a [role](https://www.vaultproject.io/api/secret/databases/postgresql.html#statements) using PostgresRole and issue credential using DatabaseAccessRequest.

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
    authMethodControllerRole: k8s.-.demo.vault-auth-method-controller
    path: kubernetes
    vaultRole: vault-policy-controller
    serviceAccountName: vault
    tokenReviewerServiceAccountName: vault-k8s-token-reviewer
    usePodServiceAccountForCsiDriver: true
```

## Enable and Configure PostgreSQL Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on specified path and configure the secret engine with given configurations.

A sample SecretEngine object for the PostgreSQL  secret engine:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: postgresql-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  path: my-postgres-se
  postgres:
    databaseRef:
      name: postgres-app
      namespace: demo
    pluginName: postgresql-database-plugin
    allowedRoles:
      - "*"
```

To configure the PostgreSQL secret engine, you need to provide the PostgreSQL database connection and authentication information through an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).

```console
$ kubectl get services -n demo
NAME       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
postgres   ClusterIP   10.97.24.153     <none>        5432/TCP    86m
```

Let's consider `postgres` is the Kubernetes service name that communicate with postgres servers. You can also connect to the database server using `URL`. Visit [AppBinding documentation](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) for more details. A sample AppBinding example with necessary k8s secret is given below:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: postgres-app
  namespace: demo
spec:
  secret:
    name: postgres-user-cred # secret
  clientConfig:
    service:
      name: postgres
      scheme: postgresql
      port: 5432
      path: "postgres"
      query: "sslmode=disable"
    insecureSkipTLSVerify: true
---
apiVersion: v1
data:
  username: cG9zdGdyZXM=
  password: cm9vdA==
kind: Secret
metadata:
  name: postgres-user-cred
  namespace: demo
```

Let's deploy SecretEngine:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/postgres/postgres-app.yaml
appbinding.appcatalog.appscode.com/postgres-app created
secret/postgres-user-cred created

$ kubectl apply -f docs/examples/guides/secret-engines/postgres/postgresSecretEngine.yaml
secretengine.engine.kubevault.com/postgresql-engine created

```

Wait till the status become `Success`:

```console
$ kubectl get secretengines -n demo
NAME                STATUS
postgresql-engine   Success
```

Since the status is `Success`, the PostgreSQL secret engine is enabled and successfully configured. You can use `kubectl describe secretengine -n <namepsace> <name>` to check for error events, if any.

## Create PostgreSQL Database Role

By using [PostgresRole](/docs/concepts/secret-engine-crds/database-secret-engine/postgresrole.md), you can create a [role](https://www.vaultproject.io/api/secret/databases/postgresql.html#statements) on the Vault server in Kubernetes native way.

A sample PostgresRole object is given below:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: PostgresRole
metadata:
  name: psql-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  path: my-postgres-se
  databaseRef:
    name: postgres-app
    namespace: demo
  creationStatements:
    - "CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';"
    - "GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";"
  defaultTTL: 1h
  maxTTL: 24h
```

Let's deploy PostgresRole:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/postgres/postgresRole.yaml
postgresrole.engine.kubevault.com/psql-role created

$ kubectl get postgresrole -n demo
NAME        AGE
psql-role   12s
```

You can also check from Vault that the role is created.
To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

```console
$ vault list my-postgres-se/roles
Keys
----
k8s.-.demo.psql-role

$ vault read my-postgres-se/roles/k8s.-.demo.psql-role
Key                      Value
---                      -----
creation_statements      [CREATE ROLE "{{name}}" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT SELECT ON ALL TABLES IN SCHEMA public TO "{{name}}";]
db_name                  k8s.-.demo.postgres-app
default_ttl              1h
max_ttl                  24h
renew_statements         []
revocation_statements    []
rollback_statements      []
```

If we delete the PostgresRole, then the respective role will be deleted from the Vault.

```console
$ kubectl delete postgresrole -n demo psql-role
postgresrole.engine.kubevault.com "psql-role" deleted
```

Check from Vault whether the role exists:

```console
$ vault read my-postgres-se/roles/k8s.-.demo.psql-role
No value found at my-postgres-se/roles/k8s.-.demo.psql-role

$ vault list my-postgres-se/roles
No value found at my-postgres-se/roles/
```

## Generate PostgreSQL Database Credentials

By using [DatabaseAccessRequest](/docs/concepts/secret-engine-crds/database-secret-engine/databaseaccessrequest.md), you can generate database access credentials from Vault.

Here, we are going to make a request to Vault for PostgreSQL database credentials by creating `postgres-cred-rqst` DatabaseAccessRequest in `demo` namespace.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: DatabaseAccessRequest
metadata:
  name: postgres-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: PostgresRole
    name: psql-role
    namespace: demo
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Here, `spec.roleRef` is the reference of PostgresRole against which credentials will be issued. `spec.subjects` is the reference to the object or user identities a role binding applies to and it will have read access of the credential secret.

Now, we are going to create DatabaseAccessRequest.

```console
$ kubectl apply -f docs/examples/guides/secret-engines/postgres/postgresAccessRequest.yaml
databaseaccessrequest.engine.kubevault.com/postgres-cred-rqst created

$ kubectl get databaseaccessrequest -n demo
NAME                 AGE
postgres-cred-rqst   34s
```

Database credentials will not be issued until it is approved. The KubeVault operator will watch for the approval in the `status.conditions[].type` field of the request object. You can use [KubeVault CLI](https://github.com/kubevault/cli), a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), to approve or deny DatabaseAccessRequest.

```console
# using KubeVault CLI as kubectl plugin to approve request
$ kubectl vault approve databaseaccessrequest postgres-cred-rqst -n demo
approved

$ kubectl get databaseaccessrequest -n demo postgres-cred-rqst -o yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: DatabaseAccessRequest
metadata:
  name: postgres-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: PostgresRole
    name: psql-role
    namespace: demo
  subjects:
  - kind: ServiceAccount
    name: demo-sa
    namespace: demo
status:
  conditions:
  - lastUpdateTime: "2019-11-20T11:40:26Z"
    message: This was approved by kubectl vault approve databaseaccessrequest
    reason: KubectlApprove
    type: Approved
  lease:
    duration: 1h0m0s
    id: my-postgres-se/creds/k8s.-.demo.psql-role/chQO2c89Wf2zieXYA9KoL9sb
    renewable: true
  secret:
    name: postgres-cred-rqst-pmq5gp
```

Once DatabaseAccessRequest is approved, the KubeVault operator will issue credentials from Vault and create a secret containing the credential. It will also create a role and rolebinding so that `spec.subjects` can access secret. You can view the information in the `status` field.

```console
$ kubectl get databaseaccessrequest postgres-cred-rqst -n demo -o json | jq '.status'
{
  "conditions": [
    {
      "lastUpdateTime": "2019-11-20T11:40:26Z",
      "message": "This was approved by kubectl vault approve databaseaccessrequest",
      "reason": "KubectlApprove",
      "type": "Approved"
    }
  ],
  "lease": {
    "duration": "1h0m0s",
    "id": "my-postgres-se/creds/k8s.-.demo.psql-role/chQO2c89Wf2zieXYA9KoL9sb",
    "renewable": true
  },
  "secret": {
    "name": "postgres-cred-rqst-pmq5gp"
  }
}

$ kubectl get secret -n demo postgres-cred-rqst-pmq5gp -o yaml
apiVersion: v1
data:
  password: QTFhLWdxYTBCeExneXdjS1hkRmI=
  username: di1rdWJlcm5ldC1rOHMuLS5kZS1lWmlkSFloyNUNUQy0xNTc0MjUwMDI2
kind: Secret
metadata:
  name: postgres-cred-rqst-pmq5gp
  namespace: demo
  ownerReferences:
  - apiVersion: engine.kubevault.com/v1alpha1
    controller: true
    kind: DatabaseAccessRequest
    name: postgres-cred-rqst
    uid: a13e98a9-22d9-4e81-975e-a8408d8cb380
type: Opaque
```

If DatabaseAccessRequest is deleted, then credential lease (if any) will be revoked.

```console
$ kubectl delete databaseaccessrequest -n demo postgres-cred-rqst
databaseaccessrequest.engine.kubevault.com "postgres-cred-rqst" deleted
```

If DatabaseAccessRequest is `Denied`, then the KubeVault operator will not issue any credential.

```console
$ kubectl vault deny databaseaccessrequest postgres-cred-rqst -n demo
  Denied
```

> Note: Once DatabaseAccessRequest is `Approved` or `Denied`, you cannot change `spec.roleRef` and `spec.subjects` field.
