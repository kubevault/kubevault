---
title: Configure Token Auth Method for Vault Server
menu:
  docs_0.1.0:
    identifier: token-auth-methods
    name: Token
    parent: auth-methods-vault-server-crds
    weight: 30
menu_name: docs_0.1.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Configure Token Auth Method for Vault Server

In Vault operator, usually Vault connection information are handled by [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md). To perform [Token Authentication](https://www.vaultproject.io/docs/auth/token.html#token-auth-method),

- You have to specify `spec.secret` in the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).

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
