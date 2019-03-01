---
title: Mount PKI(certificates) Secrets into Kubernetse pod using CSI Driver
menu:
  docs_0.1.0:
    identifier: csi-driver-pki
    name: CSI Driver
    parent: pki-secret-engines
    weight: 15
menu_name: docs_0.1.0
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Mount PKI(certificates) Secrets into Kubernetse pod using CSI Driver

## Before you Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

>Note: YAML files used in this tutorial stored in [docs/examples/csi-driver/pki](https://github.com/kubevault/docs/tree/master/docs/examples/csi-driver/pki) folder in github repository [KubeVault/docs](https://github.com/kubevault/docs).

## Configure Vault

The following steps are required to retrieve secrets from `PKI` secrets engine using `Vault server` into a Kubernetes pod.

- **Vault server:** used to provision and manage PKI(certificates) secrets
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
  <div class="tab-pane fade show active" id="operator" role="tabpanel" aria-labelledby="operator-tab">

### Using Vault Operator

Follow [this](/docs/guides/secret-engines/pki/overview.md) tutorial to manage PKI(certificates) secrets with `Vault operator`. After successful configuration you should have following resources present in your cluster.

- AppBinding: An appbinding with name `vault-app` in `demo` namespace

</div>
<div class="tab-pane fade" id="csi-driver" role="tabpanel" aria-labelledby="csi-driver-tab">

### Using Vault CLI

You can use Vault cli to manually configure an existing Vault server. The Vault server may be running inside a Kubernetes cluster or running outside a Kubernetes cluster. If you don't have a Vault server, you can deploy one by running the following command:

   ```console
    $ kubectl apply -f https://raw.githubusercontent.com/kubevault/docs/master/docs/examples/csi-driver/vault-install.yaml
    service/vault created
    statefulset.apps/vault created
   ```

  To use secret from `PKI` secret engine, you have to do following things.

1. **Enable `PKI` Engine:** To enable `PKI` secret engine run the following command.

    ```console
    $ vault secrets enable pki
    Success! Enabled the pki secrets engine at: pki/
   ```

2. **Create Engine Policy:**  To issue certificate from engine, we need to create a policy with `read`, `create`, `update`, `delete` capability. Create a `policy.hcl` file and write the following content:

    ```yaml
    # capability of get secret
    path "pki/*" {
        capabilities = ["read", "create", "update", "delete"]
    }
    ```

    Write this policy into vault naming `test-policy` with following command:

    ```console
    $ vault policy write test-policy policy.hcl
    Success! Uploaded policy: test-policy
    ```

3. **Configure CA certificate and Private key:** According to Vault documentation, Vault can accept an existing key pair, or it can generate its own self-signed root. You can learn more from [here](https://www.vaultproject.io/docs/secrets/pki/index.html#setup). In this documentation we generate self-signed root.

    ```console
    $ vault write pki/root/generate/internal \
        common_name=my-website.com \
        ttl=8760h

    Key              Value
    ---              -----
    certificate      -----BEGIN CERTIFICATE-----...
    expiration       1536807433
    issuing_ca       -----BEGIN CERTIFICATE-----...
    serial_number    7c:f1:fb:2c:6e:4d:99:0e:82:1b:08:0a:81:ed:61:3e:1d:fa:f5:29
    ```

4. **Write a PKI role:** We need to configure a role that maps a name in vault to a procedure for generating certificate. When users of machines generate credentials, they are generated agains this role:

    ```console
    $ vault write pki/roles/pki-role \
        allowed_domains=my-website.com \
        allow_subdomains=true \
        max_ttl=72h
    Success! Data written to: pki/roles/pki-role
    ```

    Here, `pki-role` will be treated as secret name on storage class.

## Configure Cluster

1. **Create Service Account:** Create `service.yaml` file with following content:

    ```yaml
    apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: ClusterRoleBinding
    metadata:
      name: role-pkicreds-binding
      namespace: demo
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: system:auth-delegator
    subjects:
    - kind: ServiceAccount
      name: pki-vault
      namespace: demo
    ---
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: pki-vault
      namespace: demo
    ```
   After that, run `kubectl apply -f service.yaml` to create a service account.

2. **Enable Kubernetes Auth:**  To enable Kubernetes auth backend, we need to extract the token reviewer JWT, Kubernetes CA certificate and Kubernetes host information.

    ```console
    export VAULT_SA_NAME=$(kubectl get sa pki-vault -n demo -o jsonpath="{.secrets[*]['name']}")

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

    $ vault write auth/kubernetes/role/pki-cred-role \
        bound_service_account_names=pki-vault \
        bound_service_account_namespaces=demo \
        policies=test-policy \
        ttl=24h
    Success! Data written to: auth/kubernetes/role/pki-cred-role
    ```

    Here, `pki-cred-role` is the name of the role.

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
      policyControllerRole: pki-cred-role # we created this in previous step
    ```

  </div>
</div>

## Mount secrets into a Kubernetes pod

After configuring `Vault server`, now we have ` vault-app` AppBinding in `demo` namespace.

So, we can create `StorageClass` now.

**Create StorageClass:** Create `storage-class.yaml` file with following content, then run `kubectl apply -f storage-class.yaml`

    ```yaml
    kind: StorageClass
    apiVersion: storage.k8s.io/v1
    metadata:
      name: vault-pki-storage
      namespace: demo
    annotations:
      storageclass.kubernetes.io/is-default-class: "false"
    provisioner: com.kubevault.csi.secrets
    parameters:
      ref: demo/vault-app # namespace/AppBinding, we created this in previous step
      engine: PKI # vault engine name
      role: pki-role # role name on vault which you want get access
      path: pki # specify the secret engine path, default is pki
    ```

   Here, you can pass the following parameters optionally to issue the certificate

    ```yaml
    common_name (string: <required>) – Specifies the requested CN for the certificate. If the CN is allowed by role policy, it will be issued.

    alt_names (string: "") – Specifies requested Subject Alternative Names, in a comma-delimited list. These can be host names or email addresses; they will be parsed into their respective fields. If any requested names do not match role policy, the entire request will be denied.

    ip_sans (string: "") – Specifies requested IP Subject Alternative Names, in a comma-delimited list. Only valid if the role allows IP SANs (which is the default).

    uri_sans (string: "") – Specifies the requested URI Subject Alternative Names, in a comma-delimited list.

    other_sans (string: "") – Specifies custom OID/UTF8-string SANs. These must match values specified on the role in allowed_other_sans (globbing allowed). The format is the same as OpenSSL: <oid>;<type>:<value> where the only current valid type is UTF8. This can be a comma-delimited list or a JSON string slice.

    ttl (string: "") – Specifies requested Time To Live. Cannot be greater than the role's max_ttl value. If not provided, the role's ttl value will be used. Note that the role values default to system values if not explicitly set.

    format (string: "") – Specifies the format for returned data. Can be pem, der, or pem_bundle; defaults to pem. If der, the output is base64 encoded. If pem_bundle, the certificate field will contain the private key and certificate, concatenated; if the issuing CA is not a Vault-derived self-signed root, this will be included as well.

    private_key_format (string: "") – Specifies the format for marshaling the private key. Defaults to der which will return either base64-encoded DER or PEM-encoded DER, depending on the value of format. The other option is pkcs8 which will return the key marshalled as PEM-encoded PKCS8.

    exclude_cn_from_sans (bool: false) – If true, the given common_name will not be included in DNS or Email Subject Alternate Names (as appropriate). Useful if the CN is not a hostname or email address, but is instead some human-readable identifier.
    ```

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
      storageClassName: vault-pki-storage
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
      serviceAccountName: pki-vault
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
    certificate       issuing_ca        private_key       private_key_type  serial_number
    ```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted
```