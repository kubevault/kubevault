---
title: Configure AWS IAM Auth Method for Vault Server
menu:
  docs_0.1.0:
    identifier: aws-iam-auth-methods
    name: AWS IAM
    parent: auth-methods-vault-server-crds
    weight: 15
menu_name: docs_0.1.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Configure AWS IAM Auth Method for Vault Server

In Vault operator, usually Vault connection information are handled by [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md). For [AWS IAM authentication](https://www.vaultproject.io/docs/auth/aws.html#iam-auth-method), it has to be enabled and configured in Vault. To perform this authenticaion:

- You have to specify `spec.secret` in the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).

- The type of the specified secret must be `"kubevault.com/aws"`.

- The specified secret data can have the following key:
    - `Secret.Data["access_key_id"]` : `Required`. Specifies AWS access key.
    - `Secret.Data["secret_access_key"]` : `Required`. Specifies AWS access secret.
    - `Secret.Data["security_token"]` : `Optional`. Specifies AWS security token.

- The specified secret annotation can have the following key:
    - `Secret.Annotations["kubevault.com/aws.header-value"]` : `Optional`. Specifies the header value that required if X-Vault-AWS-IAM-Server-ID Header is set in Vault.
    - `Secret.Annotations["kubevault.com/auth-path"]` : `Optional`. Specifies the path where AWS auth is enabled in Vault. If AWS auth is enabled in different path (not `aws`), then you have to specify it.

- The specified secret must be in AppBinding's namespace.

- You have to specify IAM auth type [role](https://www.vaultproject.io/api/auth/aws/index.html#create-role) name in `spec.parameters` of the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).
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
