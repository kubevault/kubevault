---
title: CSI Driver with MongoDB
description: Vault CSI Driver with Database(MongoDB) secret engine
menu:
  product_vault:
    identifier: mongodb-csi-driver
    name: MongoDB CSI Driver
    parent: mongodb
    weight: 10
product_name: csi-driver
menu_name: product_vault
section_menu_id: guides
---
# Setup Database(MongoDB) secret engine for Vault CSI Driver

## Before you Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

Now, you need to have vault installed either on your cluster or outside the cluster. If you want to install Vault on your cluster, you can do it by running `kubectl apply -f ` to [this](/docs/examples/csi-driver/vault-install.yaml) file.

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace "demo" created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

>Note: Yaml files used in this tutorial stored in [docs/examples/csi-driver/database/mongodb](/docs/examples/csi-driver/database/mongodb) folder in github repository [kubevault/docs](https://github.com/kubevault/docs)


## Configure Vault

To use secret from `database` engine, you have to do following things.

1. **Enable `database` Engine:** To enable `database` secret engine run the following command.

   ```console
   $ vault secrets enable database
   Success! Enabled the database secrets engine at: database/
   ```

2. **Create Engine Policy:**  To read database credentials from engine, we need to create a policy with `read` capability. Create a `policy.hcl` file and write the following content:

   ```yaml
   # capability of get secret
    path "database/*" {
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
    $ vault write database/config/my-mongodb-database \
        plugin_name=mongodb-database-plugin \
        allowed_roles="my-mongodb-role" \
        connection_url="mongodb://{{username}}:{{password}}@mongodb.acme.com:27017/admin?ssl=true" \
        username="admin" \
        password="Password!"
    ```

4. **Write a DATABASE role:** We need to configure a role that maps a name in Vault to an SQL statement to exectute to create the database credential:

   ```console
   $ vault write database/roles/my-mongodb-role \
        db_name=my-mongodb-database \
        creation_statements='{ "db": "admin", "roles": [{ "role": "readWrite" }, {"role": "read", "db": "foo"}] }' \
        default_ttl="1h" \
        max_ttl="24h"
    Success! Data written to: database/roles/my-mongodb-role
   ```

    Here, `my-mongodb-role` will be treated as secret name on storage class.

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

2. **Enable Kubernetes Auth:**  To enable Kubernetes auth backend, we need to extract the token reviewer JWT, Kubernetes CA certificate and Kubernetes host information.

    ```console
    export VAULT_SA_NAME=$(kubectl get sa db-vault -n demo -o jsonpath="{.secrets[*]['name']}")

    export SA_JWT_TOKEN=$(kubectl get secret $VAULT_SA_NAME -n demo -o jsonpath="{.data.token}" | base64 --decode; echo)

    export SA_CA_CRT=$(kubectl get secret $VAULT_SA_NAME -n demo -o jsonpath="{.data['ca\.crt']}" | base64 --decode; echo)

    export K8S_HOST=<host-ip>
    export K8s_PORT=6443
    ```

    Now, we can enable the Kubernetes authentication backend and create a Vault named role that is attached to this service account. Run:

    ```console
    $ vault auth enable kubernetes
    Success! Enabled kubernetes auth method at: kubernetes/

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

    If you don't see that CRD, then you can partially install this with following command, otherwise skip this command

    ```console
    kubectl apply -f https://raw.githubusercontent.com/kmodules/custom-resources/master/api/crds/appbinding.yaml

    ```

    If AppBinding CRD is installed, Create AppBinding with the following data:

    ```yaml
    apiVersion: appcatalog.appscode.com/v1alpha1
    kind: AppBinding
    metadata:
      name: vaultapp
      namespace: demo
    spec:
    clientConfig:
      url: http://165.227.190.238:30001 # Replace this with Vault URL
      insecureSkipTLSVerify: true
    parameters:
      apiVersion: "kubevault.com/v1alpha1"
      kind: "VaultServerConfiguration"
      usePodServiceAccountForCSIDriver: true
      authPath: "kubernetes"
      policyControllerRole: db-cred-role # we created this in previous step
    ```

4. **Create StorageClass:** Create `storage-class.yaml` file with following content, then run `kubectl apply -f storage-class.yaml`

    ```yaml
    kind: StorageClass
    apiVersion: storage.k8s.io/v1
    metadata:
      name: vault-mongodb-storage
      namespace: demo
    annotations:
      storageclass.kubernetes.io/is-default-class: "false"
    provisioner: secrets.csi.kubevault.com # For Kubernetes 1.12.x(csi-vault:0.1.0) use -> com.kubevault.csi.secrets
    parameters:
      ref: demo/vaultapp # namespace/AppBinding, we created this in previous step
      engine: DATABASE # vault engine name
      role: my-mongodb-role # role name on vault which you want get access
      path: database # specify the secret engine path, default is database
    ```
   > N.B: If you use csi-vault:0.1.0, use `com.kubevault.csi.secrets` as provisioner name.

## Test & Verify

1. **Create PVC:** Create a `PersistantVolumeClaim` with following data. This makes sure a volume will be created and provisioned on your behalf.

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
      storageClassName: vault-mongodb-storage
      volumeMode: DirectoryOrCreate
    ```

2. **Create Pod:** Now we can create a Pod which refers to this volume. When the Pod is created, the volume will be attached, formatted and mounted to the specific container.

    ```yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: mymongodbpod
      namespace: demo
    spec:
      containers:
      - name: mymongodbpod
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
    kubectl describe pods/my-pod
    ```

3. **Verify Secret:** If the Pod is running successfully, then check inside the app container by running

    ```console
    $ kubectl exec -it mymongodbpod sh
    # ls /etc/foo
    password  username
    # cat /etc/foo/username
    v-kubernet-my-mongodb-ro-kikBd7yS6VQI070gAqSh-1544693186
    ```

 So, we can see that database credentials (username, password) are mounted to the specified path.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted
```