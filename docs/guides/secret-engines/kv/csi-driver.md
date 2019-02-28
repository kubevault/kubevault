---
title: Manage Key/Value Secrets using the Vault CSI Driver
menu:
  docs_0.1.0:
    identifier: csi-driver-kv
    name: CSI Driver
    parent: kv-secret-engines
    weight: 15
menu_name: docs_0.1.0
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Key/Value Secrets using the Vault CSI Driver

## Before you Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

>Note: YAML files used in this tutorial stored in [docs/examples/csi-driver/kv](/docs/examples/csi-driver/kv) folder in github repository [KubeVault/docs](https://github.com/kubevault/docs).

## Configure Vault

To use secret from `KV` engine, you have to do following things.

1. **Enable `KV` Engine:** To enable `KV` secret engine run the following command.

   ```console
   $ vault secrets enable -version=1 kv
   Success! Enabled the kv secrets engine at: kv/
   ```

2. **Create Engine Policy:**  To read secret from engine, we need to create a policy with `read` capability. Create a `policy.hcl` file and write the following content:

   ```yaml
   # capability of get secret
    path "kv/*" {
        capabilities = ["read"]
    }
   ```

    Write this policy into vault naming `test-policy` with following command:

    ```console
    $ vault policy write test-policy policy.hcl
    Success! Uploaded policy: test-policy
    ```
3. **Write Secret on Vault:** Create a secret on vault by running:

    ```console
    $ vault kv put kv/my-secret my-value=s3cr3t
    Success! Data written to: kv/my-secret
    ```

## Configure Cluster

1. **Create Service Account:** Create `service.yaml` file with following content:

    ```yaml
    apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: ClusterRoleBinding
    metadata:
      name: role-tokenreview-binding
      namespace: demo
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: system:auth-delegator
    subjects:
    - kind: ServiceAccount
      name: kv-vault
      namespace: demo
    ---
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: kv-vault
      namespace: demo
    ```
   After that, run `kubectl apply -f service.yaml` to create a service account.

2. **Enable Kubernetes Auth:**  To enable Kubernetes auth backend, we need to extract the token reviewer JWT, Kubernetes CA certificate and Kubernetes host information.

    ```console
    export VAULT_SA_NAME=$(kubectl get sa kv-vault -n demo -o jsonpath="{.secrets[*]['name']}")

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

    $ vault write auth/kubernetes/role/kvrole \
        bound_service_account_names=kv-vault \
        bound_service_account_namespaces=demo \
        policies=test-policy \
        ttl=24h
    Success! Data written to: auth/kubernetes/role/kvrole
    ```

    Here, `kvrole` is the name of the role.

3. **Create AppBinding:** To connect CSI driver with Vault, we need to create an `AppBinding`. First we need to make sure, if `AppBinding` CRD is installed in your cluster by running:

    ```console
    $ kubectl get crd -l app=catalog
    NAME                                          CREATED AT
    appbindings.appcatalog.appscode.com           2018-12-12T06:09:34Z
    ```

   If you don't see that CRD, you can register it via the following command:

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
    parameters:
      apiVersion: "kubevault.com/v1alpha1"
      kind: "VaultServerConfiguration"
      usePodServiceAccountForCSIDriver: true
      authPath: "kubernetes"
      policyControllerRole: kvrole # we created this in previous step
    ```

4. **Create StorageClass:** Create `storage-class.yaml` file with following content, then run `kubectl apply -f storage-class.yaml`

    ```yaml
    kind: StorageClass
    apiVersion: storage.k8s.io/v1
    metadata:
      name: vault-kv-storage
      namespace: demo
    annotations:
      storageclass.kubernetes.io/is-default-class: "false"
    provisioner: secrets.csi.kubevault.com # For Kubernetes 1.12.x(csi-vault:0.1.0) use -> `com.kubevault.csi.secrets`
    parameters:
      ref: demo/vaultapp # namespace/AppBinding, we created this in previous step
      engine: KV # vault engine name
      secret: my-secret # secret name on vault which you want get access
      path: kv # specify the secret engine path, default is kv
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
      storageClassName: vault-kv-storage
      volumeMode: DirectoryOrCreate
    ```

2. **Create Pod:** Now we can create a Pod which refers to this volume. When the Pod is created, the volume will be attached, formatted and mounted to the specific container.

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
          mountPath: "/etc/foo"
          readOnly: true
      serviceAccountName: kv-vault
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
    $ kubectl exec -ti mypod /bin/sh
    / # ls /etc/foo
    my-value
    / # cat /etc/foo/my-value
    s3cr3t
    ```

   So, we can see that the secret `my-secret` is mounted into the pod, where the secret key is mounted as file and value is the content of that file.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted
```