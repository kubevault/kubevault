---
title: Configure TLS Certificates Auth Method for Vault Server
menu:
  docs_0.2.0:
    identifier: tls-auth-methods
    name: TLS Certificates
    parent: auth-methods-vault-server-crds
    weight: 25
menu_name: docs_0.2.0
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
      scheme: HTTPS
      port: 8200
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9ERXlNamN3TkRVNU1qVmFGdzB5T0RFeU1qUXdORFU1TWpWYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMVhid2wyQ1NNc2VQTU5RRzhMd3dUVWVOCkI1T05oSTlDNzFtdUoyZEZjTTlUc1VDQnlRRk1weUc5dWFvV3J1ZDhtSWpwMVl3MmVIUW5udmoybXRmWGcrWFcKSThCYkJUaUFKMWxMMFE5MlV0a1BLczlXWEt6dTN0SjJUR1hRRDhhbHZhZ0JrR1ViOFJYaUNqK2pnc1p6TDRvQQpNRWszSU9jS0xnMm9ldFZNQ0hwNktpWTBnQkZiUWdJZ1A1TnFwbksrbU02ZTc1ZW5hWEdBK2V1d09FT0YwV0Z2CmxGQmgzSEY5QlBGdTJKbkZQUlpHVDJKajBRR1FNeUxodEY5Tk1pZTdkQnhiTWhRVitvUXp2d1EvaXk1Q2pndXQKeDc3d29HQ2JtM0o4cXRybUg2Tjl6Tlc3WlR0YTdLd05PTmFoSUFEMSsrQm5rc3JvYi9BYWRKT0tMN2dLYndJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFXeWFsdUt3Wk1COWtZOEU5WkdJcHJkZFQyZnFTd0lEOUQzVjN5anBlaDVCOUZHN1UKSS8wNmpuRVcyaWpESXNHNkFDZzJKOXdyaSttZ2VIa2Y2WFFNWjFwZHRWeDZLVWplWTVnZStzcGdCRTEyR2NPdwpxMUhJb0NrekVBMk5HOGRNRGM4dkQ5WHBQWGwxdW5veWN4Y0VMeFVRSC9PRlc4eHJxNU9vcXVYUkxMMnlKcXNGCmlvM2lJV3EvU09Yajc4MVp6MW5BV1JSNCtSYW1KWjlOcUNjb1Z3b3R6VzI1UWJKWWJ3QzJOSkNENEFwOUtXUjUKU2w2blk3NVMybEdSRENsQkNnN2VRdzcwU25seW5mb3RaTUpKdmFzbStrOWR3U0xtSDh2RDNMMGNGOW5SOENTSgpiTjBiZzczeVlWRHgyY3JRYk0zcko4dUJnY3BsWlRpUy91SXJ2QT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
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
