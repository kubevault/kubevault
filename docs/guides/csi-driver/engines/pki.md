
## Setup Vault

Follow [this](https://www.vaultproject.io/docs/secrets/pki/index.html) to do following

1. Configure vault with pki engine enabled
2. Configure a CA certificate and private key
3. Configure a role that maps a name in Vault to a procedure for generating a certificate
4. Create a policy with following capabilities

```yaml
# capability to get pki cred
path "pki/*" {
  capabilities = ["read", "create", "update", "delete"]
}
```

### Setup cluster

1. Create a service account with following content
```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: role-awscreds-binding
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- kind: ServiceAccount
  name: pki-vault
  namespace: default
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pki-vault
```

2. Create a role for this serviceaccount by running

```bash
$ vault write auth/kubernetes/role/pki-cred-role bound_service_account_names=pki-vault bound_service_account_namespaces=default policies=test-policy ttl=24h
```

3. Then create a `storageclass`

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: vault-pki-storage
  annotations:
    storageclass.kubernetes.io/is-default-class: "false"
provisioner: com.vault.csi.vaultdbs
parameters:
  authRole: pki-cred-role
  secretEngine: PKI
  secretName: my-pki-role
  common_name: www.my-website.com
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


4. Create a pvc using this storageclass

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: csi-aws-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: vault-aws-storage
  volumeMode: DirectoryOrCreate

```
5. Finally run the app

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod
spec:
  containers:
  - name: mypod
    image: redis
    volumeMounts:
    - name: my-vault-volume
      mountPath: "/etc/foo"
      readOnly: true
  serviceAccountName: pki-vault
  volumes:
    - name: my-vault-volume
      persistentVolumeClaim:
        claimName: csi-pki-pvc

```