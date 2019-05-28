---
title: Mount PostgreSQL credentials into Kubernetes pod using CSI Driver
menu:
  docs_0.2.0:
    identifier: csi-driver-postgres
    name: CSI Driver
    parent: postgres-secret-engines
    weight: 15
menu_name: docs_0.2.0
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Mount PostgreSQL credentials into Kubernetes pod using CSI Driver

## Before you Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

>Note: YAML files used in this tutorial stored in [docs/examples/csi-driver/database/postgres](https://github.com/kubevault/docs/tree/master/docs/examples/csi-driver/database/postgres) folder in github repository [KubeVault/docs](https://github.com/kubevault/docs)

## Configure Vault

We need to configure following things in this step to retrieve `PostgreSQL` credentials from `Vault server` into Kubernetes pod.

- **Vault server:** used to provision and manager database credentials
- **Appbinding:** required to connect `CSI driver` with Vault server
- **Role:** using this role `CSI driver` can access credentials from Vault server

There are two ways to configure Vault server. You can use either use `Vault Operator` or use `vault` cli to manually configure a Vault server.

<ul class="nav nav-tabs" id="conceptsTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="operator-tab" data-toggle="tab" href="#operator" role="tab" aria-controls="operator" aria-selected="true">Using Vault Operator</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="csi-driver-tab" data-toggle="tab" href="#csi-driver" role="tab" aria-controls="csi-driver" aria-selected="false">Using Vault CLI</a>
  </li>
</ul>
<div class="tab-content" id="conceptsTabContent">
  <details open class="tab-pane fade show active" id="operator" role="tabpanel" aria-labelledby="operator-tab">

<summary>Using Vault Operator</summary>

Follow [this](/docs/guides/secret-engines/postgres/overview.md) tutorial to manage PostgreSQL credentials with `Vault operator`. After successful configuration you should have following resources present in your cluster.

- AppBinding: An appbinding with name `vault-app` in `demo` namespace
- Role: A role named `k8s.-.demo.demo-role` which have access to read database credential

</details>
<details class="tab-pane fade" id="csi-driver" role="tabpanel" aria-labelledby="csi-driver-tab">

<summary>Using Vault CLI</summary>

You can use Vault cli to manually configure an existing Vault server. The Vault server may be running inside a Kubernetes cluster or running outside a Kubernetes cluster. If you don't have a Vault server, you can deploy one by running the following command:

    ```console
    $ kubectl apply -f https://github.com/kubevault/docs/raw/0.2.0/docs/examples/csi-driver/vault-install.yaml
    service/vault created
    statefulset.apps/vault created
    ```

To use secret from `database` engine, you have to do following things.

1. **Enable `database` Engine:** To enable `database` secret engine run the following command.

    ```console
    $ vault secrets enable database
    Success! Enabled the database secrets engine at: database/
    ```

2. **Create Engine Policy:**  To read database credentials from engine, we need to create a policy with `read` capability. Create a `policy.hcl` file and write the following content:

    ```yaml
      # capability of get secret
      path "database/creds/*" {
          capabilities = ["read"]
      }
    ```

    Write this policy into vault naming `test-policy` with following command:

    ```console
      $ vault policy write test-policy policy.hcl
      Success! Uploaded policy: test-policy
    ```

3. **Write Secret on Vault:** Configure Vault with the proper plugin and connection information by running:

    ```console
    $ vault write database/config/my-postgresql-database \
        plugin_name=postgresql-database-plugin \
        allowed_roles="k8s.-.demo.demo-role" \
        connection_url="postgresql://{{username}}:{{password}}@159.203.114.170:30595/postgresdb?sslmode=disable" \
        username="postgresadmin" \
        password="admin123"
    ```

4. **Write a DATABASE role:** We need to configure a role that maps a name in Vault to an SQL statement to execute to create the database credential:

   ```console
   $ vault write database/roles/k8s.-.demo.demo-role \
        db_name=my-postgresql-database \
        creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
        GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
        default_ttl="1h" \
        max_ttl="24h"
    Success! Data written to: database/roles/k8s.-.demo.demo-role
   ```

  Here, `k8s.-.demo.demo-role` will be treated as secret name on storage class.

## Configure Cluster

1. **Create Service Account:** Create `service.yaml` file with following content:

     ```yaml
        apiVersion: rbac.authorization.k8s.io/v1beta1
        kind: ClusterRoleBinding
        metadata:
          name: role-dbcreds-binding
          namespace: demo
        roleRef:
          apiGroup: rbac.authorization.k8s.io
          kind: ClusterRole
          name: system:auth-delegator
        subjects:
        - kind: ServiceAccount
          name: db-vault
          namespace: demo
        ---
        apiVersion: v1
        kind: ServiceAccount
        metadata:
          name: db-vault
          namespace: demo
     ```

    After that, run `kubectl apply -f service.yaml` to create a service account.

2. **Enable Kubernetes Auth:**  To enable Kubernetes auth back-end, we need to extract the token reviewer JWT, Kubernetes CA certificate and Kubernetes host information.

    ```console
    export VAULT_SA_NAME=$(kubectl get sa db-vault -n demo -o jsonpath="{.secrets[*]['name']}")

    export SA_JWT_TOKEN=$(kubectl get secret $VAULT_SA_NAME -n demo -o jsonpath="{.data.token}" | base64 --decode; echo)

    export SA_CA_CRT=$(kubectl get secret $VAULT_SA_NAME -n demo -o jsonpath="{.data['ca\.crt']}" | base64 --decode; echo)

    export K8S_HOST=<host-ip>
    export K8s_PORT=6443
    ```

    Now, we can enable the Kubernetes authentication back-end and create a Vault named role that is attached to this service account. Run:

    ```console
    $ vault auth enable kubernetes
    Success! Enabled Kubernetes auth method at: kubernetes/

    $ vault write auth/kubernetes/config \
        token_reviewer_jwt="$SA_JWT_TOKEN" \
        kubernetes_host="https://$K8S_HOST:$K8s_PORT" \
        kubernetes_ca_cert="$SA_CA_CRT"
    Success! Data written to: auth/kubernetes/config

    $ vault write auth/kubernetes/role/db-cred-role \
        bound_service_account_names=db-vault \
        bound_service_account_namespaces=demo \
        policies=test-policy \
        ttl=24h
    Success! Data written to: auth/kubernetes/role/db-cred-role
    ```

    Here, `db-cred-role` is the name of the role.

3. **Create AppBinding:** To connect CSI driver with Vault, we need to create an `AppBinding`. First we need to make sure, if `AppBinding` CRD is installed in your cluster by running:

    ```console
    $ kubectl get crd -l app=catalog
    NAME                                          CREATED AT
    appbindings.appcatalog.appscode.com           2018-12-12T06:09:34Z
    ```

   If you don't see that CRD, you can register it via the following command:

    ```console
    kubectl apply -f https://github.com/kmodules/custom-resources/raw/master/api/crds/appbinding.yaml

    ```

    If AppBinding CRD is installed, Create AppBinding with the following data:

    ```yaml
    apiVersion: appcatalog.appscode.com/v1alpha1
    kind: AppBinding
    metadata:
      name: vault-app
      namespace: demo
    spec:
    clientConfig:
      url: http://165.227.190.238:30001 # Replace this with Vault URL
    parameters:
      apiVersion: "kubevault.com/v1alpha1"
      kind: "VaultServerConfiguration"
      usePodServiceAccountForCSIDriver: true
      authPath: "kubernetes"
      policyControllerRole: db-cred-role # we created this in previous step
    ```

  </details>
</div>

## Mount secrets into a Kubernetes pod

After configuring `Vault server`, now we have ` vault-app` AppBinding in `demo` namespace, `k8s.-.demo.demo-role` access role which have access into `database` path.

So, we can create `StorageClass` now.

**Create StorageClass:** Create `storage-class.yaml` file with following content, then run `kubectl apply -f storage-class.yaml`

  ```yaml
    kind: StorageClass
    apiVersion: storage.k8s.io/v1
    metadata:
      name: vault-pg-storage
      namespace: demo
    annotations:
      storageclass.kubernetes.io/is-default-class: "false"
    provisioner: secrets.csi.kubevault.com
    parameters:
      ref: demo/vault-app # namespace/AppBinding, we created this in previous step
      engine: DATABASE # vault engine name
      role: k8s.-.demo.demo-role # role name on vault which you want get access
      path: database # specify the secret engine path, default is database
  ```

## Test & Verify

- **Create PVC:** Create a `PersistantVolumeClaim` with following data. This makes sure a volume will be created and provisioned on your behalf.

    ```yaml
      apiVersion: v1
      kind: PersistentVolumeClaim
      metadata:
        name: csi-pvc
        namespace: demo
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: vault-pg-storage
        volumeMode: DirectoryOrCreate
    ```

- **Create Pod:** Now we can create a Pod which refers to this volume. When the Pod is created, the volume will be attached, formatted and mounted to the specific container.

    ```yaml
      apiVersion: v1
      kind: Pod
      metadata:
        name: mypgpod
        namespace: demo
      spec:
        containers:
        - name: mypgpod
          image: busybox
          command:
            - sleep
            - "3600"
          volumeMounts:
          - name: my-vault-volume
            mountPath: "/etc/foo"
            readOnly: true
        serviceAccountName: db-vault
        volumes:
          - name: my-vault-volume
            persistentVolumeClaim:
              claimName: csi-pvc
    ```

    Check if the Pod is running successfully, by running:

    ```console
    $kubectl describe pods/my-pod
    ```

- **Verify Secret:** If the Pod is running successfully, then check inside the app container by running

    ```console
    $ kubectl exec -it mypgpod sh
    # ls /etc/foo
    password  username
    # cat /etc/foo/username
    v-kubernet-my-pg-ro-kikBd7yS6VQI070gAqSh-1544693186
    ```

 So, we can see that database credentials (username, password) are mounted to the specified path.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted
```