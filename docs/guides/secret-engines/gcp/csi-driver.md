---
title: Mount GCP Secrets into Kubernetes pod using CSI Driver
menu:
  docs_{{ .version }}:
    identifier: csi-driver-gcp
    name: CSI Driver
    parent: gcp-secret-engines
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Mount GCP Secrets into a Kubernetes Pod using CSI Driver

## Before you Begin

At first, you need to have a Kubernetes 1.14 or later cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/). To check the version of your cluster, run:

```console
$ kubectl version --short
Client Version: v1.15.0
Server Version: v1.15.0
```

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace "demo" created
```

> Note: YAML files used in this tutorial stored in [examples](/docs/examples/csi-driver/gcp) folder in github repository [KubeVault/docs](https://github.com/kubevault/docs)

## Configure Vault

The following steps are required to retrieve secrets from GCP secrets engine using `Vault server` into a Kubernetes pod.

- **Vault server:** used to provision and manager GCP secrets
- **Appbinding:** required to connect `CSI driver` with Vault server
- **Role:** using this role `CSI driver` can access credentials from Vault server

There are two ways to configure Vault server. You can use either use `KubeVault operator` or use `vault` cli to manually configure a Vault server.

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

Let's assume that you have KubeVault operator installed in your cluster. If you don't have KubeVault operator yet, you can follow the [installation guide](/docs/setup/operator/install.md).

You should be familiar with the following CRDs:

- [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md)
- [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md)
- [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md)
- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [GCPRole](/docs/concepts/secret-engine-crds/gcprole.md)

We are going to start our tutorial by deploying Vault using KubeVault operator.

```console
$ cat examples/csi-driver/gcp/vault.yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  replicas: 1
  version: "1.0.1"
  backend:
    inmem: {}
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      kubernetesSecret:
        secretName: vault-keys

$ kubectl apply -f examples/csi-driver/gcp/vault.yaml
vaultserver.kubevault.com/vault created

$ kubectl get vaultserver -n demo
NAME    NODES   VERSION   STATUS    AGE
vault   1       1.0.1     Running   40h
```

Before creating [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md), we need to create a service account that has the permission to read GCP secret in vault.

```console
$ kubectl create serviceaccount -n demo demo-sa
serviceaccount/demo-sa created
```

Give permissions to `demo-sa` by [VaultPolicy](/docs/concepts/policy-crds/vaultpolicy.md) along with [VaultPolicyBinding](/docs/concepts/policy-crds/vaultpolicybinding.md)

```console
$ cat examples/csi-driver/gcp/policy.yaml
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicy
metadata:
  name: gcp-policy
  namespace: demo
spec:
  ref:
    name: vault
    namespace: demo
  policyDocument: |
    path "sys/mounts" {
      capabilities = ["read", "list"]
    }

    path "sys/mounts/*" {
      capabilities = ["create", "read", "update", "delete"]
    }

    path "gcp/*" {
        capabilities = ["create", "read", "update", "delete"]
    }

    path "sys/leases/revoke/*" {
        capabilities = ["update"]
    }
---
apiVersion: policy.kubevault.com/v1alpha1
kind: VaultPolicyBinding
metadata:
  name: gcp-role
  namespace: demo
spec:
  roleName: "gcp-role"
  policies : ["gcp-policy"]
  serviceAccountNames: ["demo-sa"]
  serviceAccountNamespaces: ["demo"]
  ttl: "1000"
  maxTTL: "2000"
  period: "1000"

$ kubectl apply -f examples/csi-driver/gcp/policy.yaml
vaultpolicy.policy.kubevault.com/gcp-policy created
vaultpolicybinding.policy.kubevault.com/gcp-role created
```
We've create `demo-sa` service account along with vault policies. Now we are ready to deploy `appbinding`: 

```console
$ kubectl apply -f examples/csi-driver/gcp/vault-app.yaml
appbinding.appcatalog.appscode.com/vault-app created
```

From your local machine check the Vault server is running with following command:

```console
$ kubectl get pods -l app=vault -n demo
NAME                     READY   STATUS    RESTARTS   AGE
vault-848797ffdf-xdnn8   3/3     Running   0          8m44s
```

To generate gcp secret under a roleset, we need to `enable` gcp secret engine and `configure` it using google service account credentials. We can do it using [GCPRole](/docs/concepts/secret-engine-crds/gcprole.md) CRD. 

Deploy secret containing google service account credentials.

```console
$ cat examples/csi-driver/gcp/gcp_cred.yaml 
apiVersion: v1
kind: Secret
metadata:
  name: gcp-cred
  namespace: demo
  annotations:
    kubevault.com/auth-path: gcp
data:
  sa.json: ewogICJ0eXBlIjogInNlcnZp...........
type: kubernetes.io/gcp

$ kubectl apply -f examples/csi-driver/gcp/gcp_cred.yaml
secret/gcp-cred created 
```
Deploy `GCPRole`:

```console
$ cat examples/csi-driver/gcp/gcpRole.yaml 
apiVersion: engine.kubevault.com/v1alpha1
kind: GCPRole
metadata: 
  name: gcp-role
  namespace: demo
spec:
  ref:
    name: vault-app
    namespace: demo
  config:
    credentialSecret: gcp-cred
  secretType: access_token
  project: ackube
  bindings: 'resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
        roles = ["roles/viewer"]
      }'
  tokenScopes: ["https://www.googleapis.com/auth/cloud-platform"]

$ kubectl apply -f examples/csi-driver/gcp/gcpRole.yaml 
gcprole.engine.kubevault.com/gcp-role created
```
Check whether `GCPRole` succeeded or not:

```console
$ kubectl get gcprole -n demo 
NAME       STATUS
gcp-role   Success
```
</details>

<details class="tab-pane fade" id="csi-driver" role="tabpanel" aria-labelledby="csi-driver-tab">

<summary>Using Vault CLI</summary>

If you don't want to use KubeVault operator and want to use Vault cli to manually configure an existing Vault server. The Vault server may be running inside a Kubernetes cluster or running outside a Kubernetes cluster. If you don't have a Vault server, you can deploy one by running the following command:

```console
$ kubectl apply -f https://github.com/kubevault/docs/raw/{{< param "info.version" >}}/docs/examples/csi-driver/vault-install.yaml
  service/vault created
  statefulset.apps/vault created
```

To generate secret from GCP secret engine, you have to do following things.

1.  **Enable GCP Secret Engine:** To enable GCP secret engine run the following command.

    ```console
    $ vault secrets enable gcp
    Success! Enabled the gcp secrets engine at: gcp/
    ```

2.  **Configure the secrets engine:** Configure the secrets engine with google service account credentials
    ```console
    $ vault write gcp/config credentials=@/home/user/Downloads/ackube-38827e3def0a.json
    Success! Data written to: gcp/config
    ```
3.  **Configure a roleset:**

    ```console
    $ vault write gcp/roleset/my-token-roleset \
                                    project="ackube" \
                                    secret_type="access_token"  \
                                    token_scopes="https://www.googleapis.com/auth/cloud-platform" \
                                 bindings='resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
                                        roles = ["roles/viewer"]
                                      }'
    Success! Data written to: gcp/roleset/my-token-roleset
    ```

4.  **Create Engine Policy:** To read secret from engine, we need to create a policy with `read` capability. Create a `policy.hcl` file and write the following content:

        ```yaml
        # capability of get secret
        path "gcp/*" {
            capabilities = ["read"]
        }
        ```

        Write this policy into vault naming `test-policy` with following command:

        ```console
        $ vault policy write test-policy policy.hcl
        Success! Uploaded policy: test-policy
        ```

    For more detailed explanation visit [Vault official website](https://www.vaultproject.io/docs/secrets/gcp/index.html#setup)

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
       name: gcp-vault
       namespace: demo
     ---
     apiVersion: v1
     kind: ServiceAccount
     metadata:
       name: gcp-vault
       namespace: demo
   ```

   After that, run `kubectl apply -f service.yaml` to create a service account.

2. **Enable Kubernetes Auth:** To enable Kubernetes auth backend, we need to extract the token reviewer JWT, Kubernetes CA certificate and Kubernetes host information.

   ```console
   export VAULT_SA_NAME=$(kubectl get sa gcp-vault -n demo -o jsonpath="{.secrets[*]['name']}")

   export SA_JWT_TOKEN=$(kubectl get secret $VAULT_SA_NAME -n demo -o jsonpath="{.data.token}" | base64 --decode; echo)

   export SA_CA_CRT=$(kubectl get secret $VAULT_SA_NAME -n demo -o jsonpath="{.data['ca\.crt']}" | base64 --decode; echo)

   export K8S_HOST=<host-ip>
   export K8s_PORT=6443
   ```

   Now, we can enable the Kubernetes authentication backend and create a Vault named role that is attached to this service account. Run:

   ```console
   $ vault auth enable kubernetes
   Success! Enabled Kubernetes auth method at: kubernetes/

   $ vault write auth/kubernetes/config \
       token_reviewer_jwt="$SA_JWT_TOKEN" \
       kubernetes_host="https://$K8S_HOST:$K8s_PORT" \
       kubernetes_ca_cert="$SA_CA_CRT"
   Success! Data written to: auth/kubernetes/config

   $ vault write auth/kubernetes/role/gcprole \
       bound_service_account_names=gcp-vault \
       bound_service_account_namespaces=demo \
       policies=test-policy \
       ttl=24h
   Success! Data written to: auth/kubernetes/role/gcprole
   ```

   Here, `gcprole` is the name of the role.

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
     policyControllerRole: gcprole # we created this in previous step
   ```

  </details>
</div>

## Mount secrets into a Kubernetes pod

Since Kubernetes 1.14, `storage.k8s.io/v1beta1` `CSINode` and `CSIDriver` objects were introduced. Let's check [CSIDriver](https://kubernetes-csi.github.io/docs/csi-driver-object.html) and [CSINode](https://kubernetes-csi.github.io/docs/csi-node-object.html) are available or not.

```console
$ kubectl get csidrivers
NAME                        CREATED AT
secrets.csi.kubevault.com   2019-07-22T11:57:02Z

$ kubectl get csinode
NAME             CREATED AT
2gb-pool-6tvtw   2019-07-22T10:54:52Z
```

After configuring `Vault server`, now we have `vault-app` AppBinding in `demo` namespace.

So, we can create `StorageClass` now.

**Create StorageClass:** Create `storage-class.yaml` file with following content, then run `kubectl apply -f storage-class.yaml`

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: vault-gcp-storage
  namespace: demo
  annotations:
    storageclass.kubernetes.io/is-default-class: "false"
provisioner: secrets.csi.kubevault.com
parameters:
  ref: demo/vault-app # namespace/AppBinding, we created this in previous step
  engine: GCP # vault engine name
  role: k8s.-.demo.gcp-role # roleset name created during vault configuration
  path: gcp # specifies the secret engine path, default is gcp
  secret_type: access_token # Specifies the type of secret generated for this role set, i.e. access_token or service_account_key
```
If roleset is created using KubeVault operator, the roleset name will follow the naming format `k8s.<cluster-name>.<namespace-name>.<GCPRole-name>`. If cluster name is not available then it will be replaced by `-`. For GCPRole name `gcp-role`, the roleset name will be `k8s.-.demo.gcp-role`

> Note: you can also provide `key_algorithm` and `key_type` fields as parameters when secret_type is service_account_key

## Test & Verify

1. **Create PVC:** Create a `PersistentVolumeClaim` with following data. This makes sure a volume will be created and provisioned on your behalf.

   ```yaml
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: csi-pvc-gcp
     namespace: demo
   spec:
     accessModes:
       - ReadWriteOnce
     resources:
       requests:
         storage: 1Gi
     storageClassName: vault-gcp-storage
   ```

2. **Create Service Account**: Create service account for the pod

   ```yaml
   apiVersion: v1
   kind: ServiceAccount
   metadata:
     name: gcp-vault
     namespace: demo
   ```

3. **Create Pod:** Now we can create a Pod which refers to this volume. When the Pod is created, the volume will be attached, formatted and mounted to the specific container.

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
             mountPath: "/etc/gcp"
             readOnly: true
     serviceAccountName: gcp-vault
     volumes:
       - name: my-vault-volume
         persistentVolumeClaim:
           claimName: csi-pvc-gcp
   ```

   Check if the Pod is running successfully, by running:

   ```console
   kubectl describe pods -n demo mypod
   ```

4. **Verify Secret:** If the Pod is running successfully, then check inside the app container by running

   ```console
   $ kubectl exec -it -n demo mypod sh
   / # ls /etc/gcp/
   expires_at_seconds  token               token_ttl
   / # cat /etc/gcp/token
   ya29.c.ElkqB1eesBVOX7Xg_Ip3RKEJtfgOLaP0......
   ```

   So, we can see that the secret `token` is mounted into the pod, where the secret key is mounted as file and value is the content of that file.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted
```
