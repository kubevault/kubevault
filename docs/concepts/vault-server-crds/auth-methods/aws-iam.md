---
title: Connect to Vault using AWS IAM Auth Method
menu:
  docs_{{ .version }}:
    identifier: aws-iam-auth-methods
    name: AWS IAM
    parent: auth-methods-vault-server-crds
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Connect to Vault using AWS IAM Auth Method

The KubeVault operator uses an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) to connect to an externally provisioned Vault server. For [AWS IAM authentication](https://www.vaultproject.io/docs/auth/aws.html#iam-auth-method), it has to be enabled and configured in the Vault server. Follow the steps below to create an appropriate AppBinding:

- You have to specify `spec.secret` in the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).

- The type of the specified secret must be `"kubevault.com/aws"`.

- The specified secret data can have the following key:
  - `Secret.Data["access_key_id"]` : `Required`. Specifies AWS access key.
  - `Secret.Data["secret_access_key"]` : `Required`. Specifies AWS access secret.
  - `Secret.Data["security_token"]` : `Optional`. Specifies AWS security token.

- The specified secret annotation can have the following key:
  - `Secret.Annotations["kubevault.com/aws.header-value"]` : `Optional`. Specifies the header value that required if X-Vault-AWS-IAM-Server-ID Header is set in Vault.
  - `Secret.Annotations["kubevault.com/auth-path"]` : `Optional`. Specifies the path where AWS auth is enabled in Vault. If AWS auth is enabled in a different path (not `aws`), then you have to specify it.

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
      scheme: HTTPS
      port: 8200
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9ERXlNamN3TkRVNU1qVmFGdzB5T0RFeU1qUXdORFU1TWpWYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMVhid2wyQ1NNc2VQTU5RRzhMd3dUVWVOCkI1T05oSTlDNzFtdUoyZEZjTTlUc1VDQnlRRk1weUc5dWFvV3J1ZDhtSWpwMVl3MmVIUW5udmoybXRmWGcrWFcKSThCYkJUaUFKMWxMMFE5MlV0a1BLczlXWEt6dTN0SjJUR1hRRDhhbHZhZ0JrR1ViOFJYaUNqK2pnc1p6TDRvQQpNRWszSU9jS0xnMm9ldFZNQ0hwNktpWTBnQkZiUWdJZ1A1TnFwbksrbU02ZTc1ZW5hWEdBK2V1d09FT0YwV0Z2CmxGQmgzSEY5QlBGdTJKbkZQUlpHVDJKajBRR1FNeUxodEY5Tk1pZTdkQnhiTWhRVitvUXp2d1EvaXk1Q2pndXQKeDc3d29HQ2JtM0o4cXRybUg2Tjl6Tlc3WlR0YTdLd05PTmFoSUFEMSsrQm5rc3JvYi9BYWRKT0tMN2dLYndJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFXeWFsdUt3Wk1COWtZOEU5WkdJcHJkZFQyZnFTd0lEOUQzVjN5anBlaDVCOUZHN1UKSS8wNmpuRVcyaWpESXNHNkFDZzJKOXdyaSttZ2VIa2Y2WFFNWjFwZHRWeDZLVWplWTVnZStzcGdCRTEyR2NPdwpxMUhJb0NrekVBMk5HOGRNRGM4dkQ5WHBQWGwxdW5veWN4Y0VMeFVRSC9PRlc4eHJxNU9vcXVYUkxMMnlKcXNGCmlvM2lJV3EvU09Yajc4MVp6MW5BV1JSNCtSYW1KWjlOcUNjb1Z3b3R6VzI1UWJKWWJ3QzJOSkNENEFwOUtXUjUKU2w2blk3NVMybEdSRENsQkNnN2VRdzcwU25seW5mb3RaTUpKdmFzbStrOWR3U0xtSDh2RDNMMGNGOW5SOENTSgpiTjBiZzczeVlWRHgyY3JRYk0zcko4dUJnY3BsWlRpUy91SXJ2QT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-cred
  namespace: demo
  annotations:
    kubevault.com/aws.header-value: hello
    kubevault.com/auth-path: my-aws
type: kubevault.com/aws
data:
  access_key_id: cm9vdA==
  secret_access_key: cm9vdA==
```
