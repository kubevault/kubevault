---
title: Configure Kubernetes Auth Method for Vault Server
menu:
  docs_0.1.0:
    identifier: kubernetes-auth-methods
    name: Kubernetes
    parent: auth-methods-vault-server-crds
    weight: 20
menu_name: docs_0.1.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Configure Kubernetes Auth Method for Vault Server

In Vault operator, usually Vault connection information are handled by [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md). For [Kubernetes Authentication](https://www.vaultproject.io/docs/auth/kubernetes.html), it has to be enabled and configured in Vault. In Vault operator, it can be performed in two ways:

- Using ServiceAccount Name
- Using ServiceAccount Token Secret

### Kubernetes Authentication using ServiceAccount Name

To perform Kubernetes Authentication using ServiceAccount Name,

- You have to specify serviceaccount name and [role](https://www.vaultproject.io/api/auth/kubernetes/index.html#create-role) name in `spec.parameters` of the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md). If Kubernetes auth method is enabled in different path (not `kubernetes`), then you also have to specify it in `spec.parameters` of AppBinding.

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

- You have to specify serviceaccount token secret in `spec.secret` of the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md). Kubernetes create token secret for every serviceaccount. You can use that in `spec.secret`.

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

- You have to specify [role](https://www.vaultproject.io/api/auth/kubernetes/index.html#create-role) name in `spec.parameters` of the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).
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
