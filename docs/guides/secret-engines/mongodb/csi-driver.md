---
title: Mount MongoDB credentials using CSI Driver
menu:
  docs_{{ .version }}:
    identifier: csi-driver-mongodb
    name: CSI Driver
    parent: mongodb-secret-engines
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Mount MongoDB credentials using CSI Driver

At first, you need to have a Kubernetes 1.14 or later cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/). To check the version of your cluster, run:

```console
$ kubectl version --short
Client Version: v1.16.2
Server Version: v1.14.0
```

Before you begin:

- Install KubeVault operator in your cluster from [here](/docs/setup/operator/install.md).
- Install KubeVault CSI driver in your cluster from [here](/docs/setup/csi-driver/install.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial stored in [examples](/docs/examples/guides/secret-engines/mongodb) folder in GitHub repository [KubeVault/docs](https://github.com/kubevault/docs)

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
    path: kubernetes
    vaultRole: vault-policy-controller
    kubernetes:
      serviceAccountName: vault
      tokenReviewerServiceAccountName: vault-k8s-token-reviewer
      usePodServiceAccountForCSIDriver: true
```

## Enable and Configure MongoDB Database Secrets Engine

The following steps are required to enable and configure the MongoDB database secrets engine in the Vault server.

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
- [MongoDBRole](/docs/concepts/secret-engine-crds/database-secret-engine/mongodb.md)

Let's enable and configure the MongoDB database secret engine by deploying the following `SecretEngine` yaml:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: mongodb-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  mongodb:
    databaseRef:
      name: mongo-app
      namespace: demo
```

To configure the MongoDB secret engine, you need to provide the MongoDB connection information through an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).

```console
$ kubectl get services -n demo
NAME    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)           AGE
mongo   ClusterIP   10.98.184.214   <none>        27017/TCP         7h7m
```

Let's consider `mongo` is the Kubernetes service name that communicate with MongoDB servers. You can also connect to the database server using `URL`. Visit [AppBinding documentation](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) for more details. A sample AppBinding example with necessary k8s secret is given below:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: mongo-app
  namespace: demo
spec:
  secret:
    name: mongo-user-cred # k8s secret name
  clientConfig:
    service:
      name: mongo
      scheme: mongodb
      port: 27017
    insecureSkipTLSVerify: true
---
apiVersion: v1
data:
  username: cm9vdA== # base64 encoded database username
  password: cm9vdA== # base64 encoded database password
kind: Secret
metadata:
  name: mongo-user-cred
  namespace: demo
```

Let's deploy SecretEngine:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mongodb/mongodbApp.yaml
appbinding.appcatalog.appscode.com/mongo-app created
secret/mongo-user-cred created

$ kubectl apply -f docs/examples/guides/secret-engines/mongodb/mongoSecretEngine.yaml
secretengine.engine.kubevault.com/mongodb-engine created
```

Wait till the status become `Success`:

```console
$ kubectl get secretengines -n demo
NAME             STATUS
mongodb-engine   Success
```

Create database role using the following `MongoDBRole` yaml:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: MongoDBRole
metadata:
  name: mdb-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  databaseRef:
    name: mongo-app
    namespace: demo
  creationStatements:
    - "{ \"db\": \"admin\", \"roles\": [{ \"role\": \"readWrite\" }, {\"role\": \"read\", \"db\": \"foo\"}] }"
  defaultTTL: 1h
  maxTTL: 24h
```

Let's deploy MongoDBRole:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mongodb/mongoRole.yaml
mongodbrole.engine.kubevault.com/mdb-role created

$ kubectl get mongodbrole -n demo
NAME       AGE
mdb-role   16s
```

You can also check from Vault that the role is created.
To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

```console
$ vault list database/roles
Keys
----
k8s.-.demo.mdb-role

$ vault read database/roles/k8s.-.demo.mdb-role
Key                      Value
---                      -----
creation_statements      [{ "db": "admin", "roles": [{ "role": "readWrite" }, {"role": "read", "db": "foo"}] }]
db_name                  k8s.-.demo.mongo-app
default_ttl              1h
max_ttl                  24h
renew_statements         []
revocation_statements    []
rollback_statements      []
```

</div>
<div class="tab-pane fade" id="csi-driver" role="tabpanel" aria-labelledby="csi-driver-tab">

### Using Vault CLI

You can also use [Vault CLI](https://www.vaultproject.io/docs/commands/) to [enable and configure](https://www.vaultproject.io/docs/secrets/databases/mongodb.html#setup) the MongoDB secret engine.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

To generate secret from the database secret engine, you have to perform the following steps.

To use secret from `database` engine, you have to do following things.

- **Enable Secret Engine:** To enable the `database` secret engine, run the following command.

```console
$ vault secrets enable database
Success! Enabled the database secrets engine at: database/
```

- **Configure Secret Engine:** Configure Vault with the proper plugin and connection information by running:

```console
$ vault write database/config/k8s.-.demo.mongo-app \
    plugin_name=mongodb-database-plugin \
    allowed_roles="*" \
    connection_url="mongodb://{{username}}:{{password}}@mongodb.acme.com:27017/admin?ssl=true" \
    username="admin" \
    password="Password!"
Success! Data written to: database/config/k8s.-.demo.mongo-app
```

- **Configure a Role:** We need to configure a role that maps a name in Vault to an SQL statement to execute to create the database credential:

```console
$ vault write database/roles/k8s.-.demo.mdb-role \
    db_name=k8s.-.demo.mongo-app \
    creation_statements='{ "db": "admin", "roles": [{ "role": "readWrite" }, {"role": "read", "db": "foo"}] }' \
    default_ttl="1h" \
    max_ttl="24h"
Success! Data written to: database/roles/k8s.-.demo.mdb-role
```

- **Read the Role:**

```console
$ vault list database/roles
Keys
----
k8s.-.demo.mdb-role

$ vault read database/roles/k8s.-.demo.mdb-role
Key                      Value
---                      -----
creation_statements      [{ "db": "admin", "roles": [{ "role": "readWrite" }, {"role": "read", "db": "foo"}] }]
db_name                  k8s.-.demo.mongo-app
default_ttl              1h
max_ttl                  24h
renew_statements         []
revocation_statements    []
rollback_statements      []
```

If you use Vault CLI to enable and configure the MongoDB secret engine then you need to **update the vault policy** for the service account 'vault' created during vault server configuration and add the permission to read at "database/roles/*" with previous permissions. That is why it is recommended to use the KubeVault operator because the operator updates the policies automatically when needed.

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

After configuring the `Vault server`, now we have AppBinding `vault` in `demo` namespace.

### Create StorageClass

Create `StorageClass` object with the following content:

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: vault-mdb-storage
  annotations:
    storageclass.kubernetes.io/is-default-class: "false"
provisioner: secrets.csi.kubevault.com
parameters:
  ref: demo/vault # namespace/AppBinding, we created during vault server configuration
  engine: DATABASE # vault engine name
  role: k8s.-.demo.mdb-role # role name on vault which you want get access
  path: database # specify the secret engine path, default is database
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mongodb/storageClass.yaml
storageclass.storage.k8s.io/vault-mdb-storage created
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
  name: csi-pvc-mdb
  namespace: trial
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Mi
  storageClassName: vault-mdb-storage
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mongodb/pvc.yaml
persistentvolumeclaim/csi-pvc-mdb created
```

### Create VaultPolicy and VaultPolicyBinding for Pod's Service Account

Let's say pod's service account name is `pod-sa` located in `trial` namespace. We need to create a [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) and a [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md) so that the pod has access to read secrets from the Vault server.

```yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: mdb-se-policy
  namespace: demo
spec:
  vaultRef:
    name: vault
  # Here, mongodb secret engine is enabled at "database".
  # If the path was "demo-se", policy should be like
  # path "demo-se/*" {}.
  policyDocument: |
    path "database/*" {
      capabilities = ["create", "read"]
    }
---
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: mdb-se-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  policies:
  - ref: mdb-se-policy
  subjectRef:
    kubernetes:
      serviceAccountNames:
        - "pod-sa"
      serviceAccountNamespaces:
        - "trial"
```

Let's create VaultPolicy and VaultPolicyBinding:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mongodb/vaultPolicy.yaml
vaultpolicy.policy.kubevault.com/mdb-se-policy created

$ kubectl apply -f docs/examples/guides/secret-engines/mongodb/vaultPolicyBinding.yaml
vaultpolicybinding.policy.kubevault.com/mdb-se-role created
```

Check if the VaultPolicy and the VaultPolicyBinding are successfully registered to the Vault server:

```console
$ kubectl get vaultpolicy -n demo
NAME                           STATUS    AGE
mdb-se-policy                  Success   8s

$ kubectl get vaultpolicybinding -n demo
NAME                           STATUS    AGE
mdb-se-role                    Success   10s
```

### Create Service Account for Pod

Let's create the service account `pod-sa` which was used in VaultPolicyBinding. When a VaultPolicyBinding object is created, the KubeVault operator create an auth role in the Vault server. The role name is generated by the following naming format: `k8s.(clusterName or -).namespace.name`. Here, it is `k8s.-.demo.mdb-se-role`. We need to provide the auth role name as service account `annotations` while creating the service account. If the annotation `secrets.csi.kubevault.com/vault-role` is not provided, the CSI driver will not be able to perform authentication to the Vault.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pod-sa
  namespace: trial
  annotations:
    secrets.csi.kubevault.com/vault-role: k8s.-.demo.mdb-se-role
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mongodb/podServiceAccount.yaml
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
      mountPath: "/etc/mongodb"
      readOnly: true
  serviceAccountName: pod-sa
  volumes:
  - name: my-vault-volume
    persistentVolumeClaim:
      claimName: csi-pvc-mdb
```

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mongodb/pod.yaml
pod/mypod created
```

Check if the Pod is running successfully, by running:

```console
$ kubectl get pods -n trial
NAME                    READY   STATUS    RESTARTS   AGE
mypod                   1/1     Running   0          11s
```

### Verify Secret

If the Pod is running successfully, then check inside the app container by running:

```console
$ kubectl exec -it -n trial  mypod sh
/ # ls /etc/mongodb/
password  username

/ # cat /etc/mongodb/username
v-kubernetes-demo-k8s.-.demo.mdb--pAlXCTq9UoTcZM7LP0uH

/ # cat /etc/mongodb/password
A1a-fjNMk4aRr4kdyTc2
```

So, we can see that database credentials (username, password) are mounted to the specified path.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted

$ kubectl delete ns trial
namespace "trial" deleted
```
