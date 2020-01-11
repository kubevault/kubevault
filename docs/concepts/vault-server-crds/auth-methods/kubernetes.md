---
title: Connect to Vault using Kubernetes Auth Method
menu:
  docs_{{ .version }}:
    identifier: kubernetes-auth-methods
    name: Kubernetes
    parent: auth-methods-vault-server-crds
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Connect to Vault using Kubernetes Auth Method

The KubeVault operator uses an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) to connect to an externally provisioned Vault server. For [Kubernetes Authentication](https://www.vaultproject.io/docs/auth/kubernetes.html), it has to be enabled and configured in the Vault server. In KubeVault operator, it can be performed in two ways:

- Using ServiceAccount Name
- Using ServiceAccount Token Secret

## Kubernetes Authentication using ServiceAccount Name

To perform Kubernetes Authentication using ServiceAccount Name,

- You have to specify serviceaccount name and [role](https://www.vaultproject.io/api/auth/kubernetes/index.html#create-role) name in `spec.parameters` of the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md). If Kubernetes auth method is enabled in different path (not `kubernetes`), then you also have to specify it in `spec.parameters` of AppBinding.

  ```yaml
  spec:
    parameters:
      apiVersion: config.kubevault.com/v1alpha1
      kind: VaultServerConfiguration
      path: k8s # Kubernetes auth is enabled in this path
      vaultRole: demo  # role name against which login will be done
      kubernetes:
        serviceAccountName: vault-sa # service account name
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
    apiVersion: config.kubevault.com/v1alpha1
    kind: VaultServerConfiguration
    path: k8s
    vaultRole: demo
    kubernetes:
      serviceAccountName: vault-sa
  clientConfig:
    service:
      name: vault
      scheme: HTTPS
      port: 8200
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9ERXlNamN3TkRVNU1qVmFGdzB5T0RFeU1qUXdORFU1TWpWYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMVhid2wyQ1NNc2VQTU5RRzhMd3dUVWVOCkI1T05oSTlDNzFtdUoyZEZjTTlUc1VDQnlRRk1weUc5dWFvV3J1ZDhtSWpwMVl3MmVIUW5udmoybXRmWGcrWFcKSThCYkJUaUFKMWxMMFE5MlV0a1BLczlXWEt6dTN0SjJUR1hRRDhhbHZhZ0JrR1ViOFJYaUNqK2pnc1p6TDRvQQpNRWszSU9jS0xnMm9ldFZNQ0hwNktpWTBnQkZiUWdJZ1A1TnFwbksrbU02ZTc1ZW5hWEdBK2V1d09FT0YwV0Z2CmxGQmgzSEY5QlBGdTJKbkZQUlpHVDJKajBRR1FNeUxodEY5Tk1pZTdkQnhiTWhRVitvUXp2d1EvaXk1Q2pndXQKeDc3d29HQ2JtM0o4cXRybUg2Tjl6Tlc3WlR0YTdLd05PTmFoSUFEMSsrQm5rc3JvYi9BYWRKT0tMN2dLYndJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFXeWFsdUt3Wk1COWtZOEU5WkdJcHJkZFQyZnFTd0lEOUQzVjN5anBlaDVCOUZHN1UKSS8wNmpuRVcyaWpESXNHNkFDZzJKOXdyaSttZ2VIa2Y2WFFNWjFwZHRWeDZLVWplWTVnZStzcGdCRTEyR2NPdwpxMUhJb0NrekVBMk5HOGRNRGM4dkQ5WHBQWGwxdW5veWN4Y0VMeFVRSC9PRlc4eHJxNU9vcXVYUkxMMnlKcXNGCmlvM2lJV3EvU09Yajc4MVp6MW5BV1JSNCtSYW1KWjlOcUNjb1Z3b3R6VzI1UWJKWWJ3QzJOSkNENEFwOUtXUjUKU2w2blk3NVMybEdSRENsQkNnN2VRdzcwU25seW5mb3RaTUpKdmFzbStrOWR3U0xtSDh2RDNMMGNGOW5SOENTSgpiTjBiZzczeVlWRHgyY3JRYk0zcko4dUJnY3BsWlRpUy91SXJ2QT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
```

## Kubernetes Authentication using ServiceAccount Token Secret

To perform Kubernetes Authentication using ServiceAccount Token Secret,

- You have to specify serviceaccount token secret in `spec.secret` of the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md). Kubernetes create a token secret for every serviceaccount. You can use that in `spec.secret`.

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

- The specified secret must be in AppBinding's namespace.

- The specified token secret must have the following key:
  - `Secret.Data["token"]` : `Required`. Specifies the serviceaccount token.

- The type of the specified token secret must be `kubernetes.io/service-account-token`.

- The additional information required for the Kubernetes authentication method can be provided as AppBinding's `spec.parameters`.
  
  ```yaml
  spec:
    parameters:
      apiVersion: config.kubevault.com/v1alpha1
      kind: VaultServerConfiguration
      path: k8s
      vaultRole: demo-role
  ```

  - `path` : `optional`. Specifies the path where Kubernetes auth is enabled in Vault. If this path is not provided, the path will be set by default path `kubernetes`.
  - `vaultRole` : `required`. Specifies the name of the Vault auth [role](https://www.vaultproject.io/api/auth/kubernetes/index.html#create-role) against which login will be performed.

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
    apiVersion: config.kubevault.com/v1alpha1
    kind: VaultServerConfiguration
    path: k8s
    vaultRole: demo-role
  clientConfig:
    service:
      name: vault
      scheme: HTTPS
      port: 8200
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9ERXlNamN3TkRVNU1qVmFGdzB5T0RFeU1qUXdORFU1TWpWYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMVhid2wyQ1NNc2VQTU5RRzhMd3dUVWVOCkI1T05oSTlDNzFtdUoyZEZjTTlUc1VDQnlRRk1weUc5dWFvV3J1ZDhtSWpwMVl3MmVIUW5udmoybXRmWGcrWFcKSThCYkJUaUFKMWxMMFE5MlV0a1BLczlXWEt6dTN0SjJUR1hRRDhhbHZhZ0JrR1ViOFJYaUNqK2pnc1p6TDRvQQpNRWszSU9jS0xnMm9ldFZNQ0hwNktpWTBnQkZiUWdJZ1A1TnFwbksrbU02ZTc1ZW5hWEdBK2V1d09FT0YwV0Z2CmxGQmgzSEY5QlBGdTJKbkZQUlpHVDJKajBRR1FNeUxodEY5Tk1pZTdkQnhiTWhRVitvUXp2d1EvaXk1Q2pndXQKeDc3d29HQ2JtM0o4cXRybUg2Tjl6Tlc3WlR0YTdLd05PTmFoSUFEMSsrQm5rc3JvYi9BYWRKT0tMN2dLYndJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFXeWFsdUt3Wk1COWtZOEU5WkdJcHJkZFQyZnFTd0lEOUQzVjN5anBlaDVCOUZHN1UKSS8wNmpuRVcyaWpESXNHNkFDZzJKOXdyaSttZ2VIa2Y2WFFNWjFwZHRWeDZLVWplWTVnZStzcGdCRTEyR2NPdwpxMUhJb0NrekVBMk5HOGRNRGM4dkQ5WHBQWGwxdW5veWN4Y0VMeFVRSC9PRlc4eHJxNU9vcXVYUkxMMnlKcXNGCmlvM2lJV3EvU09Yajc4MVp6MW5BV1JSNCtSYW1KWjlOcUNjb1Z3b3R6VzI1UWJKWWJ3QzJOSkNENEFwOUtXUjUKU2w2blk3NVMybEdSRENsQkNnN2VRdzcwU25seW5mb3RaTUpKdmFzbStrOWR3U0xtSDh2RDNMMGNGOW5SOENTSgpiTjBiZzczeVlWRHgyY3JRYk0zcko4dUJnY3BsWlRpUy91SXJ2QT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
```
