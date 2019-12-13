---
title: Mount PKI(certificates) Secrets into Kubernetse pod using CSI Driver
menu:
  docs_{{ .version }}:
    identifier: csi-driver-pki
    name: CSI Driver
    parent: pki-secret-engines
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Mount PKI(certificates) Secrets into Kubernetse pod using CSI Driver

At first, you need to have a Kubernetes 1.14 or later cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/). To check the version of your cluster, run:

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

> Note: YAML files used in this tutorial stored in [examples](/docs/examples/guides/secret-engins/pki) folder in github repository [KubeVault/docs](https://github.com/kubevault/docs)

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

## Enable and Configure PKI Secret Engine

We will use the [Vault CLI](https://www.vaultproject.io/docs/commands/#vault-commands-cli-) throughout the tutorial to [enable and configure](https://www.vaultproject.io/docs/secrets/pki/index.html#setup) the PKI secret engine.

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

To use secret from `PKI` secret engine, you have to perform the following steps.

### Enable PKI Secret Engine

To enable `PKI` secret engine run the following command.

```console
$ vault secrets enable pki
Success! Enabled the pki secrets engine at: pki/
```

Increase the TTL by tuning the secrets engine. The default value of 30 days may be too short, so increase it to 1 year:

```console
$ vault secrets tune -max-lease-ttl=8760h pki
Success! Tuned the secrets engine at: pki/
```

### Configure CA Certificate and Private Key

Configure a CA certificate and private key. Vault can accept an existing key pair, or it can generate its own self-signed root.

```console
$ vault write pki/root/generate/internal \
                          common_name=my-website.com \
                          ttl=8760h
Key              Value
---              -----
certificate      -----BEGIN CERTIFICATE-----
MIIDPjCCAiagAwIBAgIUEDmnAmC0siISlrezD3/CeUXTSfswDQYJKoZIhvcNAQEL
... ... ... ... ... ... ... ... ... ... ... ... ... ... ... ...
CsFVu+vfMM9XEMYeKHRWAq9onJFyGuwKGhF0/7RbZ3EunTj6Zph+UMucGoL4xfXj
ITltdU1N4JPvihQq+8Omryay
-----END CERTIFICATE-----
expiration       1606200496
issuing_ca       -----BEGIN CERTIFICATE-----
MIIDPjCCAiagAwIBAgIUEDmnAmC0siISlrezD3/CeUXTSfswDQYJKoZIhvcNAQEL
... ... ... ... ... ... ... ... ... ... ... ... ... ... ... ...
CsFVu+vfMM9XEMYeKHRWAq9onJFyGuwKGhF0/7RbZ3EunTj6Zph+UMucGoL4xfXj
ITltdU1N4JPvihQq+8Omryay
-----END CERTIFICATE-----
serial_number    10:39:a7:02:60:b4:b2:22:12:96:b7:b3:0f:7f:c2:79:45:d3:49:fb
```

### Configure a PKI Role

We need to configure a role that maps a name in vault to a procedure for generating certificate. When users of machines generate credentials, they are generated agains this role:

```console
$ vault write pki/roles/example-dot-com \
                          allowed_domains=my-website.com \
                          allow_subdomains=true \
                          max_ttl=72h
Success! Data written to: pki/roles/example-dot-com
```

## Update Vault Policy

Since Pod's service account will be used by the CSI driver to perform [Kubernetes authentication](https://www.vaultproject.io/docs/auth/kubernetes.html) to the Vault server, the auth method role must have the permission to read secret at `pki/*` path.

During the Vault server configuration, we have created a Kubernetes service account and registered an auth method role at the Vault server. If you have used the KubeVault operator to deploy the Vault server, then the operator has performed these tasks for you.

So, we have the service account that will be referenced from the Pod.

```console
kubectl get serviceaccounts -n demo
NAME                       SECRETS   AGE
vault                      1         7h23m
```

> Don't have Vault CLI? Download and configure it as described [here](/docs/guides/vault-server/vault-server.md#enable-vault-cli)

You can find the name of the auth method role in the AppBinding's `spec.parameters.policyControllerRole`. Let's list the token policies assigned for `vault` service account:

```console
$ vault read auth/kubernetes/role/vault-policy-controller
Key                                 Value
---                                 -----
bound_service_account_names         [vault]
bound_service_account_namespaces    [demo]
token_bound_cidrs                   []
token_explicit_max_ttl              0s
token_max_ttl                       24h
token_no_default_policy             false
token_num_uses                      0
token_period                        24h
token_policies                      [default vault-policy-controller]
token_ttl                           24h
token_type                          default
```

Now, we will update the Vault policy `vault-policy-controller` and add the permissions for `pki/*` path with existing permissions.

`pki-policy.hcl:`

```yaml

path "pki/*" {
    capabilities = ["read", "create", "update", "delete"]
}
```

Update the `vault-policy-controller` policy:

```console
# write existing polices to a file
$ vault policy read vault-policy-controller > examples/guides/secret-engins/pki/policy.hcl

# append the pki-policy at the end of the existing policies
$ cat examples/guides/secret-engins/pki/pki-policy.hcl >> examples/guides/secret-engins/pki/policy.hcl

# write the update policy to Vault
$ vault policy write vault-policy-controller examples/guides/secret-engins/pki/policy.hcl
Success! Uploaded policy: vault-policy-controller

# read updated policy
$ vault policy read vault-policy-controller
... ...
... ...
path "pki/*" {
    capabilities = ["read", "create", "update", "delete"]
}
```

So, we have updated the policy successfully and ready to mount the secrets into Kubernetes pods.

## Mount Certificates into a Kubernetes Pod

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

So, we can create `StorageClass` now.

### Create StorageClass

Create `StorageClass` object with the following content:

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: vault-pki-storage
  namespace: demo
  annotations:
    storageclass.kubernetes.io/is-default-class: "false"
provisioner: secrets.csi.kubevault.com
parameters:
  ref: demo/vault # namespace/AppBinding, we created during vault configuration
  engine: PKI # vault engine name
  role: example-dot-com # role name created
  path: pki # specifies the secret engine path, default is gcp
  common_name: www.my-website.com # specifies the requested CN for the certificate
```

```console
$ kubectl apply -f examples/guides/secret-engins/pki/storageClass.yaml
storageclass.storage.k8s.io/vault-pki-storage created
```

Here, you can also pass the following parameters optionally to issue the certificate

- `common_name` (string: <required>) – Specifies the requested CN for the certificate. If the CN is allowed by role policy, it will be issued.

- `alt_names` (string: "") – Specifies requested Subject Alternative Names, in a comma-delimited list. These can be host names or email addresses; they will be parsed into their respective fields. If any requested names do not match role policy, the entire request will be denied.

- `ip_sans` (string: "") – Specifies requested IP Subject Alternative Names, in a comma-delimited list. Only valid if the role allows IP SANs (which is the default).

- `uri_sans` (string: "") – Specifies the requested URI Subject Alternative Names, in a comma-delimited list.

- `other_sans` (string: "") – Specifies custom OID/UTF8-string SANs. These must match values specified on the role in allowed_other_sans (globbing allowed). The format is the same as OpenSSL: <oid>;<type>:<value> where the only current valid type is UTF8. This can be a comma-delimited list or a JSON string slice.

- `ttl` (string: "") – Specifies requested Time To Live. Cannot be greater than the role's max_ttl value. If not provided, the role's ttl value will be used. Note that the role values default to system values if not explicitly set.

- `format` (string: "") – Specifies the format for returned data. Can be pem, der, or pem_bundle; defaults to pem. If der, the output is base64 encoded. If pem_bundle, the certificate field will contain the private key and certificate, concatenated; if the issuing CA is not a Vault-derived self-signed root, this will be included as well.

- `private_key_format` (string: "") – Specifies the format for marshaling the private key. Defaults to der which will return either base64-encoded DER or PEM-encoded DER, depending on the value of format. The other option is pkcs8 which will return the key marshalled as PEM-encoded PKCS8.

- `exclude_cn_from_sans` (bool: false) – If true, the given common_name will not be included in DNS or Email Subject Alternate Names (as appropriate). Useful if the CN is not a hostname or email address, but is instead some human-readable identifier.

## Test & Verify

### Create PVC

Create a `PersistentVolumeClaim` with following data. This makes sure a volume will be created and provisioned on your behalf.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: csi-pvc-pki
  namespace: demo
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Mi
  storageClassName: vault-pki-storage
```

```console
$ kubectl apply -f examples/guides/secret-engins/pki/pvc.yaml 
persistentvolumeclaim/csi-pvc-pki created
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
          mountPath: "/etc/pki"
          readOnly: true
  serviceAccountName: vault # service account that was created during vault server configuration
  volumes:
    - name: my-vault-volume
      persistentVolumeClaim:
        claimName: csi-pvc-pki
```

```console
$ kubectl apply -f examples/guides/secret-engins/pki/pod.yaml
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
/ # ls /etc/pki/
certificate       issuing_ca        private_key_type
expiration        private_key       serial_number

/ # cat /etc/pki/certificate
-----BEGIN CERTIFICATE-----
MIIDVjCCAj6gAwIBAgIUNjTBC3qR7Zaj0XrzUc3QEbE+EhgwDQYJKoZIhvcNAQEL
BQAwGTEXMBUGA1UEAxMObXktd2Vic2l0ZS5jb20wHhcNMTkxMjEzMTExNDIwWhcN
..... .... .... .... .... .... .... .... .... .... .... .... ...
bo901cITjNyCTbAF2401pYFZ4rSlxhcuAvc7c27uqvKEh2/ctMGRkvPVygbPdvB8
LfCskfX0sk8PQiEznlmYlChK3KNsEp+xSCyjU+pDEw8AcDXwE6vVFft/fRX0oiHH
KIzTZ7R/QKUkLisloMUHStINISAehglLZTJjo79jB7GN66wyqP+E8iRLEYFAAsb0
aZ5wuSTYEpqOuP6G1tOdhiE7iptFu9Wg9dKtmXkZnc0iTBL60xMUUapH
-----END CERTIFICATE-----
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns demo
namespace "demo" deleted
```