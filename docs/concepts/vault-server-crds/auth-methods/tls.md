---
title: Configure TLS Certificates Auth Method for Vault Server
menu:
  docs_0.1.0:
    identifier: tls-auth-methods
    name: TLS Certificates
    parent: auth-methods-vault-server-crds
    weight: 25
menu_name: docs_0.1.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Configure TLS Certificates Auth Method for Vault Server

In Vault operator, usually Vault connection information are handled by [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md). For [TLS Certificates authentication](https://www.vaultproject.io/docs/auth/cert.html), it has to be enabled and configured in Vault. To perform it,

- You have to specify `spec.secret` in the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).

- The type of the specified secret must be `kubernetes.io/tls`.

- The specified secret data must have the following key:
    - `Secret.Data["tls.crt"]` : `Required`. Specifies the tls certificate.
    - `Secret.Data["tls.key"]` : `Required`. Specifies the tls private key.

- The specified secret annotation can have the following key:
    - `Secret.Annotations["kubevault.com/auth-path"]` : `Optional`. Specifies the path where TLS certificate auth is enabled in Vault. If TLS certificate auth is enabled in different path (not `cert`), then you have to specify it.

- You have to specify [role](https://www.vaultproject.io/api/auth/cert/index.html#create-ca-certificate-role) name in `spec.parameters` of the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).
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
