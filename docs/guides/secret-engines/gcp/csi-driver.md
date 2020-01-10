---
title: Mount GCP Secrets using CSI Driver
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

# Mount GCP Secrets using CSI Driver

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

> Note: YAML files used in this tutorial stored in [examples](/docs/examples/guides/secret-engines/gcp) folder in github repository [KubeVault/docs](https://github.com/kubevault/docs)

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
      usePodServiceAccountForCsiDriver: true
```

## Enable and Configure GCP Secret Engine

The following steps are required to enable and configure the GCP secrets engine in the Vault server.

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
- [GCPRole](/docs/concepts/secret-engine-crds/gcp-secret-engine/gcprole.md)

Let's enable and configure GCP secret engine by deploying the following `SecretEngine` yaml:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: gcp-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  gcp:
    credentialSecret: gcp-cred
```

To configure the GCP secret engine, you need to provide google service account credentials through a Kubernetes secret.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gcp-cred
  namespace: demo
data:
  sa.json: ewogICJ0e...== ## base64 encoded google service account credential
```

Let's deploy SecretEngine:

```console
$ kubectl apply -f docs/examples/guides/secret-engines/gcp/gcpCred.yaml
secret/gcp-cred created

$ kubectl apply -f docs/examples/guides/secret-engines/gcp/gcpSecretEngine.yaml
secretengine.engine.kubevault.com/gcp-engine created
```

Wait till the status become `Success`:

```console
$ kubectl get secretengines -n demo
NAME         STATUS
gcp-engine   Success
```

Create GCP roleset using the following `GCPRole` yaml:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: GCPRole
metadata:
  name: gcp-role
  namespace: demo
spec:
  vaultRef:
    name: vault
  secretType: access_token
  project: ackube
  bindings: |
    resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
      roles = ["roles/viewer"]
    }
  tokenScopes:
    - https://www.googleapis.com/auth/cloud-platform
```

Let's deploy GCPRole:

```console
$ kubectl apply -f examples/guides/secret-engines/gcp/gcpRole.yaml
gcprole.engine.kubevault.com/gcp-role created

$ kubectl get gcprole -n demo
NAME       STATUS
gcp-role   Success
```

You can also check from Vault that the roleset is created.
To resolve the naming conflict, name of the roleset in Vault will follow this format: `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

```console
$ vault list gcp/roleset
Keys
----
k8s.-.demo.gcp-role

$ vault read gcp/roleset/k8s.-.demo.gcp-role
Key                      Value
---                      -----
bindings                 map[//cloudresourcemanager.googleapis.com/projects/ackube:[roles/viewer]]
project                  ackube
secret_type              access_token
service_account_email    vaultk8s---demo-gcp-424523423@ackube.iam.gserviceaccount.com
token_scopes             [https://www.googleapis.com/auth/cloud-platform]
```

</div>
<div class="tab-pane fade" id="csi-driver" role="tabpanel" aria-labelledby="csi-driver-tab">

### Using Vault CLI

You can also use [Vault CLI](https://www.vaultproject.io/docs/commands/) to [enable and configure](https://www.vaultproject.io/docs/secrets/gcp/index.html#setup) the GCP secret engine.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

To generate secret from the GCP secret engine, you have to perform the following steps.

- **Enable GCP Secret Engine:** To enable the GCP secret engine, run the following command.

```console
$ vault secrets enable gcp
Success! Enabled the gcp secrets engine at: gcp/
```

- **Configure the secrets engine:** Configure the secrets engine with google service account credentials

```console
$ vault write gcp/config credentials=@/home/user/Downloads/ackube-sa.json
Success! Data written to: gcp/config
```

- **Configure a roleset:** Configure a roleset that generates OAuth2 access tokens

```console
$ vault write gcp/roleset/k8s.-.demo.gcp-role \
                                project="ackube" \
                                secret_type="access_token"  \
                                token_scopes="https://www.googleapis.com/auth/cloud-platform" \
                              bindings='resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
                                    roles = ["roles/viewer"]
                                  }'
Success! Data written to: gcp/roleset/k8s.-.demo.gcp-role
```

- **Read the roleset:**

```console
$ vault read gcp/roleset/k8s.-.demo.gcp-role
Key                      Value
---                      -----
bindings                 map[//cloudresourcemanager.googleapis.com/projects/ackube:[roles/viewer]]
project                  ackube
secret_type              access_token
service_account_email    vaultk8s---demo-gcp-424523423@ackube.iam.gserviceaccount.com
token_scopes             [https://www.googleapis.com/auth/cloud-platform]
```

If you use Vault CLI to enable and configure the GCP secret engine then you need to **update the vault policy** for the service account 'vault' created during vault server configuration and add the permission to read at "gcp/roleset/*" with previous permissions. That is why it is recommended to use the KubeVault operator because the operator updates the policies automatically when needed.

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

So, we can create `StorageClass` now.

### Create StorageClass

Create `StorageClass` object with the following content:

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
  ref: demo/vault # namespace/AppBinding, we created this in previous step
  engine: GCP # vault engine name
  role: k8s.-.demo.gcp-role # roleset name created during vault configuration
  path: gcp # specifies the secret engine path, default is gcp
  secret_type: access_token # Specifies the type of secret generated for this role set, i.e. access_token or service_account_key
```

```console
$ kubectl apply -f examples/guides/secret-engines/gcp/storageClass.yaml
storageclass.storage.k8s.io/vault-gcp-storage created
```

> Note: you can also provide `key_algorithm` and `key_type` fields as parameters when `secret_type` is `service_account_key`.

## Test & Verify

### Create PVC

Create a `PersistentVolumeClaim` with the following data. This makes sure a volume will be created and provisioned on your behalf.

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
      storage: 100Mi
  storageClassName: vault-gcp-storage
```

```console
$ kubectl apply -f examples/guides/secret-engines/gcp/pvc.yaml
persistentvolumeclaim/csi-pvc-gcp created
```

### Create Pod

Now we can create a Pod which refers to this volume. When the Pod is created, the volume will be attached, formatted and mounted to the specific container.

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
  serviceAccountName: vault # service account that was created during vault server configuration
  volumes:
    - name: my-vault-volume
      persistentVolumeClaim:
        claimName: csi-pvc-gcp
```

```console
$ kubectl apply -f examples/guides/secret-engines/gcp/pod.yaml
pod/mypod created
```

Check if the Pod is running successfully, by running:

```console
$ kubectl get pods -n demo
NAME                    READY   STATUS    RESTARTS   AGE
mypod                   1/1     Running   0          11s
```

### Verify Secret

If the Pod is running successfully, then check inside the app container by running

```console
$ kubectl exec -it -n demo  mypod sh
/ # ls /etc/gcp
expires_at_seconds  token               token_ttl

/ # cat /etc/gcp/token
ya29.c.Kl20BwwWtb6DoTjY4-eSVgQQq.......
```

So, we can see that the secret `token` is mounted into the pod, where the secret key is mounted as file and the value is the content of that file.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted
```
