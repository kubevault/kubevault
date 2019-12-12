---
title: Mount MySQL/MariaDB credentials into Kubernetes pod using CSI Driver
menu:
  docs_{{ .version }}:
    identifier: csi-driver-mysql
    name: CSI Driver
    parent: mysql-secret-engines
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Mount MySQL/MariaDB credentials into Kubernetes pod using CSI Driver

At first, you need to have a Kubernetes 1.14 or later cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Kind](https://github.com/kubernetes-sigs/kind). To check the version of your cluster, run:

```console
$ kubectl version --short
Client Version: v1.16.2
Server Version: v1.14.0

```

Before you begin:

- Install KubeVault operator in your cluster from [here](/docs/setup/operator/install.md).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial stored in [examples](/docs/examples/guides/secret-engins/mysql) folder in github repository [KubeVault/docs](https://github.com/kubevault/docs)

## Vault Server

If you don't have a Vault Server, you can deploy it by using the KubeVault operator.

- [Deploy Vault Server](/docs/guides/vault-server/vault-server.md)

The KubeVault operator is also compatible with external Vault servers that are not provisioned by itself. You need to configure both the Vault server and the cluster so that the KubeVault operator can communicate with your Vault server.

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
    authMethodControllerRole: k8s.-.demo.vault-auth-method-controller
    kind: VaultServerConfiguration
    path: kubernetes
    policyControllerRole: vault-policy-controller
    serviceAccountName: vault
    tokenReviewerServiceAccountName: vault-k8s-token-reviewer
    usePodServiceAccountForCsiDriver: true
```

## Enable and Configure MySQL Database Secret Engine

The following steps are required to enable and configure the MySQL database secrets engine in the Vault server.

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
  <details open class="tab-pane fade show active" id="operator" role="tabpanel" aria-labelledby="operator-tab">

<summary>Using KubeVault operator</summary>

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [MySQLRole](/docs/concepts/secret-engine-crds/database-secret-engine/mysql.md)

Let's enable and configure the MySQL database secret engine by deploying the following `SecretEngine` yaml:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: mysql-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  path: mysql-se
  mysql:
    databaseRef:
      name: mysql-app
      namespace: demo
    pluginName: "mysql-rds-database-plugin"
    allowedRoles:
      - "*"
```

To configure the MySQL secret engine, you need to provide the MySQL database connection and authentication information through an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).

```console
$ kubectl get services -n demo
NAME    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)       AGE
mysql   ClusterIP   10.96.33.240    <none>        3306/TCP      3h41
```

Let's consider `mysql` is the Kubernetes service name that communicate with MySQL servers. The connection `URL` generated using the service will be `mysql.demo.svc:3306`. Visit [AppBinding documentation](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) for more details. A sample AppBinding example with necessary k8s secret is given below:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: mysql-app
  namespace: demo
spec:
  secret:
    name: mysql-cred # secret name
  clientConfig:
    url: tcp(mysql.demo.svc:3306)/
    insecureSkipTLSVerify: true
---
apiVersion: v1
data:
  username: cm9vdA== # mysql username
  password: cm9vdA== # mysql password
kind: Secret
metadata:
  name: mysql-cred
  namespace: demo
```

Let's deploy SecretEngine:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mysql/mysql-app.yaml 
appbinding.appcatalog.appscode.com/mysql-app created
secret/mysql-cred created

$ kubectl apply -f docs/examples/guides/secret-engines/mysql/mysqlSecretEngine.yaml
secretengine.engine.kubevault.com/mysql-engine created
```

Wait till the status become `Success`:

```console
$ kubectl get secretengines -n demo
NAME           STATUS
mysql-engine   Success
```

Create database role using the following `MySQLRole` yaml:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: MySQLRole
metadata:
  name: mysql-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  path: mysql-se
  databaseRef:
    name: mysql-app
    namespace: demo
  creationStatements:
    - "CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';"
    - "GRANT SELECT ON *.* TO '{{name}}'@'%';"
  defaultTTL: 1h
  maxTTL: 24h
```

Let's deploy MySQLRole:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/mysql/mysqlRole.yaml
mysqlrole.engine.kubevault.com/mysql-role created

$ kubectl get mysqlrole -n demo mysql-role 
NAME         AGE
mysql-role   83m
```

You can also check from Vault that the role is created.
To resolve the naming conflict, name of the role in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

```console
$ vault list mysql-se/roles
Keys
----
k8s.-.demo.mysql-role

$ vault read mysql-se/roles/k8s.-.demo.mysql-role
Key                      Value
---                      -----
creation_statements      [CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}'; GRANT SELECT ON *.* TO '{{name}}'@'%';]
db_name                  k8s.-.demo.mysql-app
default_ttl              1h
max_ttl                  24h
renew_statements         []
revocation_statements    []
rollback_statements      []
```

</details>
<details class="tab-pane fade" id="csi-driver" role="tabpanel" aria-labelledby="csi-driver-tab">

<summary>Using Vault CLI</summary>

You can also use [Vault CLI](https://www.vaultproject.io/docs/commands/) to [enable and configure](https://www.vaultproject.io/docs/secrets/databases/mysql-maria.html#setup) the MySQL secret engine.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

To generate secret from the database secret engine, you have to perform the following steps.

- **Enable `database` Engine:** To enable `database` secret engine run the following command.

```console
$ vault secrets enable -path=mysql-se database
Success! Enabled the database secrets engine at: mysql-se/
```

- **Configure Secret Engine:** Configure Vault with the proper plugin and connection information by running:

```console
$ vault write mysql-se/config/k8s.-.demo.mysql-app \
    plugin_name=mysql-rds-database-plugin \
    allowed_roles="*" \
    connection_url="{{username}}:{{password}}@tcp(127.0.0.1:3306)/" \
    username="root" \
    password="password"
Success! Data written to: mysql-se/config/k8s.-.demo.mysql-app
```

- **Configure a Role:** We need to configure a role that maps a name in Vault to an SQL statement to execute to create the database credential:

```console
$ vault write mysql-se/roles/k8s.-.demo.mysql-role \
    db_name=my-mysql-database \
    creation_statements="CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';GRANT SELECT ON *.* TO '{{name}}'@'%';" \
    default_ttl="1h" \
    max_ttl="24h"
Success! Data written to: mysql-se/roles/k8s.-.demo.mysql-role
```

- **Read the Role:**

```console
$ vault list mysql-se/roles
Keys
----
k8s.-.demo.mysql-role

$ vault read mysql-se/roles/k8s.-.demo.mysql-role
Key                      Value
---                      -----
creation_statements      [CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}'; GRANT SELECT ON *.* TO '{{name}}'@'%';]
db_name                  k8s.-.demo.mysql-app
default_ttl              1h
max_ttl                  24h
renew_statements         []
revocation_statements    []
rollback_statements      []
```

If you use Vault CLI to enable and configure the MySQL secret engine then you need to **update the vault policy** for the service account 'vault' [created during vault server configuration] and add the permission to read at "mysql-se/roles/*" with previous permissions. That is why it is recommended to use the KubeVault operator because the operator updates the policies automatically when needed.

Find how to update the policy for service account in [here](/docs/guides/secret-engines/kv/csi-driver.md#update-vault-policy).

  </details>
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

- **Create StorageClass:** Create `StorageClass` object with the following content:

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: vault-mysql-storage
  namespace: demo
  annotations:
    storageclass.kubernetes.io/is-default-class: "false"
provisioner: secrets.csi.kubevault.com
parameters:
  ref: demo/vault # namespace/AppBinding, we created during vault server configuration
  engine: DATABASE # vault engine name
  role: k8s.-.demo.mysql-role # role name on vault which you want get access
  path: mysql-se # specify the secret engine path, default is database
```

```console
$ kubectl apply -f examples/guides/secret-engins/mysql/storageClass.yaml
storageclass.storage.k8s.io/vault-mysql-storage created
```

## Test & Verify

- **Create PVC:** Create a `PersistentVolumeClaim` with following data. This makes sure a volume will be created and provisioned on your behalf.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: csi-pvc-mysql
  namespace: demo
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Mi
  storageClassName: vault-mysql-storage
```

```console
$ kubectl apply -f examples/guides/secret-engins/mysql/pvc.yaml
persistentvolumeclaim/csi-pvc-mysql created
```

- **Create Pod:** Now we can create a Pod which refers to this volume. When the Pod is created, the volume will be attached, formatted and mounted to the specific container.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod
  namespace: demo
spec:
  containers:
  - name: mypod
    image: busybox
    command:
      - sleep
      - "3600"
    volumeMounts:
    - name: my-vault-volume
      mountPath: "/etc/mysql"
      readOnly: true
  serviceAccountName: vault # service account that was created during vault server configuration
  volumes:
    - name: my-vault-volume
      persistentVolumeClaim:
        claimName: csi-pvc-mysql
```

```console
$ kubectl apply -f examples/guides/secret-engins/mysql/pod.yaml
pod/mypod created
```

Check if the Pod is running successfully, by running:

```console
$ kubectl get pods -n demo
NAME                    READY   STATUS    RESTARTS   AGE
mypod                   1/1     Running   0          11s
```

- **Verify Secret:** If the Pod is running successfully, then check inside the app container by running

```console
$ kubectl exec -it -n demo  mypod sh
/ # ls /etc/mysql/
password  username

/ # cat /etc/mysql/username
v-k8s.-k0gFzzJyf

/ # cat /etc/mysql/password
A1a-7Q0RtEmH0Gj8lO4N
```

 So, we can see that database credentials (username, password) are mounted to the specified path.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted
```
