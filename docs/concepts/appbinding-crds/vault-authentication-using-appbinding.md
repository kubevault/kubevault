# Vault Authentication using AppBinding in Vault operator

In Vault operator, usually Vault connection information are handled by [AppBinding](/docs/concepts/appbinding-crds/appbinding.md). Following authentication method are currently supported by Vault operator using AppBinding:

- [Token Auth Method](https://www.vaultproject.io/docs/auth/token.html#token-auth-method)
- [Kubernetes Auth Method](https://www.vaultproject.io/docs/auth/kubernetes.html)
- [AWS IAM Auth Method](https://www.vaultproject.io/docs/auth/aws.html#iam-auth-method)
- [Userpass Auth Method](https://www.vaultproject.io/docs/auth/userpass.html)
- [TLS Certificates Auth Method](https://www.vaultproject.io/docs/auth/cert.html)

## Token Auth Method

To perform Token Authentication, 

- You have to specify `spec.secret` in the [AppBinding](/docs/concepts/appbinding-crds/appbinding.md).

- The type of the specified secret must be `kubevault.com/token`.

- The specified secret data must have the following key:
    - `Secret.Data["token"]` : `Required`. Specifies the Vault authentication token.
    
- The specified secret must be in AppBinding's namespace.

Sample AppBinding and Secret is given below:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  secret:
    name: vault-token
  clientConfig:
    service:
      name: vault
      scheme: http
      port: 8200
    insecureSkipTLSVerify: true
```

```yaml
apiVersion: v1
data:
  token: cm9vdA==
kind: Secret
metadata:
  name: vault-token
  namespace: demo
type: kubevault.com/token
```

## Kubernetes Auth Method

For Kubernetes Authentication, it has to be enabled and configured in Vault. In Vault operator, it can be performed in two ways:

- Using ServiceAccount Name

- Using ServiceAccount Token Secret

### Kubernetes Authentication using ServiceAccount Name

To perform Kubernetes Authentication using ServiceAccount Name,

- You have to specify serviceaccount name and [role](https://www.vaultproject.io/api/auth/kubernetes/index.html#create-role) name in `spec.parameters` of the [AppBinding](/docs/concepts/appbinding-crds/appbinding.md). If Kubernetes auth method is enabled in different path (not `kubernetes`), then you also have to specify it in `spec.parameters` of AppBinding. 

    ```yaml
    spec:
      parameters:
        serviceAccountName: vault-sa
        policyControllerRole: demo # role name against which login will be done
        authPath: k8s # kubernetes auth is enabled in this path
    ```
- The specified ServiceAccount must be in AppBinding's namespace.

Sample AppBinding is given below:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  parameters:
    serviceAccountName: vault-sa
    policyControllerRole: demo # role name against which login will be done
    authPath: k8s # kubernetes auth is enabled in this path
  clientConfig:
    service:
      name: vault
      scheme: http
      port: 8200
    insecureSkipTLSVerify: true
```

### Kubernetes Authentication using ServiceAccount Token Secret

To perform Kubernetes Authentication using ServiceAccount Token Secret,

- You have to specify serviceaccount token secret in `spec.secret` of the [AppBinding](/docs/concepts/appbinding-crds/appbinding.md). Kubernetes create token secret for every serviceaccount. You can use that in `spec.secret`.
    
    ```console
    $ kubectl create serviceaccount sa
    serviceaccount/sa created
    
    $ kubectl get serviceaccount/sa -o yaml
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: sa
      namespace: default
    secrets:
    - name: sa-token-6n9pv
    
    $ kubectl get secrets/sa-token-6n9pv -o yaml
    apiVersion: v1
    data:
      ca.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t
      namespace: ZGVmYXVsdA==
      token: ZXlKaGJHY2lPaUpTVXpJMU5pSXNJbXRwWkNJNklpSjkuZXlK
    kind: Secret
    metadata:
      annotations:
        kubernetes.io/service-account.name: sa
        kubernetes.io/service-account.uid: db22a517-0771-11e9-8744-080027907e77
      name: sa-token-6n9pv
      namespace: default
    type: kubernetes.io/service-account-token
    ```
- The specified token secret must have the following key:
    - `Secret.Data["token"]` : `Required`. Specifies the serviceaccount token.

- The specified secret annotation can have the following key:
    - `Secret.Annotations["kubevault.com/auth-path"]` : `Optional`. Specifies the path where kubernetes auth is enabled in Vault. If kubernetes auth is enabled in different path (not `kubernetes`), then you have to specify it.
  
- The type of the specified token secret must be `kubernetes.io/service-account-token`.

- The specified secret must be in AppBinding's namespace.

- You have to specify [role](https://www.vaultproject.io/api/auth/kubernetes/index.html#create-role) name in `spec.parameters` of the [AppBinding](/docs/concepts/appbinding-crds/appbinding.md).
    ```yaml
    spec:
      parameters:
        policyControllerRole: demo # role name against which login will be done
    ``` 

Sample AppBinding is given below:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  secret:
    name: sa-token
  parameters:
    policyControllerRole: demo # role name against which login will be done
  clientConfig:
    service:
      name: vault
      scheme: http
      port: 8200
    insecureSkipTLSVerify: true
```

### AWS IAM Auth Method

For AWS IAM authentication, it has to be enabled and configured in Vault. To perform this authenticaion:

- You have to specify `spec.secret` in the [AppBinding](/docs/concepts/appbinding-crds/appbinding.md).

- The type of the specified secret must be `"kubevault.com/aws"`.

- The specified secret data can have the following key:
    - `Secret.Data["access_key_id"]` : `Required`. Specifies AWS access key.
    - `Secret.Data["secret_access_key"]` : `Required`. Specifies AWS access secret.
    - `Secret.Data["security_token"]` : `Optional`. Specifies AWS security token.
    
- The specified secret annotation can have the following key:
    - `Secret.Annotations["kubevault.com/aws.header-value"]` : `Optional`. Specifies the header value that required if X-Vault-AWS-IAM-Server-ID Header is set in Vault.
    - `Secret.Annotations["kubevault.com/auth-path"]` : `Optional`. Specifies the path where AWS auth is enabled in Vault. If AWS auth is enabled in different path (not `aws`), then you have to specify it.
    
- The specified secret must be in AppBinding's namespace.

- You have to specify IAM auth type [role](https://www.vaultproject.io/api/auth/aws/index.html#create-role) name in `spec.parameters` of the [AppBinding](/docs/concepts/appbinding-crds/appbinding.md).
    ```yaml
    spec:
      parameters:
        policyControllerRole: demo # role name against which login will be done
    ```
Sample AppBinding and Secret is given below:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  secret:
    name: aws-cred
  parameters:
    policyControllerRole: demo # role name against which login will be done
  clientConfig:
    service:
      name: vault
      scheme: http
      port: 8200
    insecureSkipTLSVerify: true
```

```yaml
apiVersion: v1
data:
  access_key_id: cm9vdA==
  secret_access_key: cm9vdA==
kind: Secret
metadata:
  name: aws-cred
  namespace: demo
  annotations:
    kubevault.com/aws.header-value: hello
    kubevault.com/auth-path: my-aws
type: kubevault.com/aws
```

## Userpass Auth Method

For Userpass authentication, it has to be enabled and configured in Vault. To perform it, 

- You have to specify `spec.secret` in the [AppBinding](/docs/concepts/appbinding-crds/appbinding.md).

- The type of the specified secret must be `kubernetes.io/basic-auth`.

- The specified secret data must have the following key:
    - `Secret.Data["username"]` : `Required`. Specifies the username used for authentication.
    - `Secret.Data["password"]` : `Required`. Specifies the password used for authentication.

- The specified secret annotation can have the following key:
    - `Secret.Annotations["kubevault.com/auth-path"]` : `Optional`. Specifies the path where userpass auth is enabled in Vault. If userpass auth is enabled in different path (not `userpass`), then you have to specify it.
  
- The specified secret must be in AppBinding's namespace.

Sample AppBinding and Secret is given below:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  secret:
    name: userpass-cred
  clientConfig:
    service:
      name: vault
      scheme: http
      port: 8200
    insecureSkipTLSVerify: true
```

```yaml
apiVersion: v1
data:
  username: cm9vdA==
  password: cm9vdA==
kind: Secret
metadata:
  name: userpass-cred
  namespace: demo
  annotations:
    kubevault.com/auth-path: my-userpass
type: kubernetes.io/basic-auth
```

## TLS Certificates Auth Method

For TLS Certificates authentication, it has to be enabled and configured in Vault. To perform it, 

- You have to specify `spec.secret` in the [AppBinding](/docs/concepts/appbinding-crds/appbinding.md).

- The type of the specified secret must be `kubernetes.io/tls`.

- The specified secret data must have the following key:
    - `Secret.Data["tls.crt"]` : `Required`. Specifies the tls certificate.
    - `Secret.Data["tls.key"]` : `Required`. Specifies the tls private key.

- The specified secret annotation can have the following key:
    - `Secret.Annotations["kubevault.com/auth-path"]` : `Optional`. Specifies the path where TLS certificate auth is enabled in Vault. If TLS certificate auth is enabled in different path (not `cert`), then you have to specify it.

- You have to specify [role](https://www.vaultproject.io/api/auth/cert/index.html#create-ca-certificate-role) name in `spec.parameters` of the [AppBinding](/docs/concepts/appbinding-crds/appbinding.md).
    ```yaml
    spec:
      parameters:
        policyControllerRole: demo # role name against which login will be done
    ```

- The specified secret must be in AppBinding's namespace.

Sample AppBinding and Secret is given below:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  secret:
    name: tls
  clientConfig:
    service:
      name: vault
      scheme: http
      port: 8200
    insecureSkipTLSVerify: true
  parameters:
    policyControllerRole: demo
```

```yaml
apiVersion: v1
data:
  tls.crt: cm9vdA==
  tls.key: cm9vdA==
kind: Secret
metadata:
  name: tls
  namespace: demo
  annotations:
    kubevault.com/auth-path: my-cert
type: kubernetes.io/tls
```
