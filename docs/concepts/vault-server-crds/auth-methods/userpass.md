---
title: Configure Userpass Auth Method for Vault Server
menu:
  docs_0.1.0:
    identifier: userpass-auth-methods
    name: Userpass
    parent: auth-methods-vault-server-crds
    weight: 35
menu_name: docs_0.1.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Configure Userpass Auth Method for Vault Server

In Vault operator, usually Vault connection information are handled by [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md). For [Userpass authentication](https://www.vaultproject.io/docs/auth/userpass.html), it has to be enabled and configured in Vault. To perform it,

- You have to specify `spec.secret` in the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).

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
