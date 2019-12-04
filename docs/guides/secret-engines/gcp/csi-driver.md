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

At first, you need to have a Kubernetes 1.14 or later cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Kind](https://github.com/kubernetes-sigs/kind). To check the version of your cluster, run:

```console
$ kubectl version --short
Client Version: v1.16.2
Server Version: v1.14.0

```

Before you begin:

- Install KubeVault operator in your cluster from [here](/docs/setup/operator/install).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial stored in [examples](/docs/examples/csi-driver/gcp) folder in github repository [KubeVault/docs](https://github.com/kubevault/docs)

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

## Enable and Configure GCP Secret Engine

The following steps are required to enable and configure GCP secrets engine in the Vault server.

There are two ways to configure Vault server. You can use either use the `KubeVault operator` or the  `Vault CLI` to manually configure a Vault server.

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
- [GCPRole](/docs/concepts/secret-engine-crds/gcp-secret-engine/gcprole.md)

Let's enable and configure GCP secret engine by deploying the following yaml:

```yaml

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
