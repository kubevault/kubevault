---
title: Connect to Vault using GCP IAM Auth Method
menu:
  docs_{{ .version }}:
    identifier: gcp-iam-auth-methods
    name: GCP IAM
    parent: auth-methods-vault-server-crds
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Connect to Vault using GCP IAM Auth Method

The KubeVault operator uses an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) to connect to an externally provisioned Vault server. For [GCP IAM authentication](https://www.vaultproject.io/docs/auth/gcp.html#configuration), it has to be enabled and configured in the Vault server. Follow the steps below to create an appropriate AppBinding:

- You have to specify `spec.secret` in the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).

- The specified secret must be in AppBinding's namespace.

- The type of the specified secret must be `"kubevault.com/gcp"`.

- The specified secret data can have the following key:
  - `Secret.Data["sa.json"]` : `Required`. Specifies the google application credentials.

- The additional information required for the GCP IAM authentication method can be provided as AppBinding's `spec.parameters`.
  
  ```yaml
    spec:
      parameters:
        apiVersion: config.kubevault.com/v1alpha1
        kind: VaultServerConfiguration
        path: my-gcp
        vaultRole: demo-role
  ```

  - `path` : `optional`. Specifies the path where GCP auth is enabled in Vault. If this path is not provided, the path will be set by default path `gcp`.
  - `vaultRole` : `required`. Specifies the name of the Vault auth [role](https://www.vaultproject.io/api/auth/gcp/index.html#create-role) against which login will be performed.

Sample AppBinding and Secret is given below:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  secret:
    name: gcp-cred
  parameters:
    apiVersion: config.kubevault.com/v1alpha1
    kind: VaultServerConfiguration
    path: my-gcp
    vaultRole: demo-role
  clientConfig:
    service:
      name: vault
      scheme: HTTPS
      port: 8200
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9UQTBNVGN3TkRFM05UVmFGdzB5T1RBME1UUXdOREUzTlRWYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBNGxKeTNsMThKMjZ2WWlZakNYZm54M1dvCnpaNUh4WUUvWmVNT1BKTFZyN0kwb2Rrc3N6bXhieWROaGJaN2kvQ2xOUzRvclB1eVFJZ29Ncng1bHRvTDhEd1cKRmZQZ0RGbFM4WjArcHNwRU00WEtVYnpBQk9lY0JaUnhZYTJPdmJqeFZjTE1PYzI5VGw4TzYzc2ZFeTlDcWhrRApEaUZDeFQ2bFd1MjZ0YmNzZEwwNFdBVzZDN1pyakhtaUMvWHhGcnl6STllRUVhb0xkVTdHMDJhTmFmOVBZM0RaCjBNRTJtOUNXMDYzOFZMeCtZMjR3cXMrQVJrUmg3cUVVKy9qK2lId2Y5N0hKMi8vcGdNMEFRcklOSk1kQ1ZNZEsKZ2hGYUs4NjZBdDNsa1R0U0FOclVtM3pCN1lEN20rYjNTRlJ5clMzM2RDZG5zNlp3NDdTQjFjWEJxcCs1SFFJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFXeUZTR1drSkV3U25nTW5zQXJ3c2xYckIyQm1NNHdNdzJtSkwyWk9CSDVlWm9ka0QKVmpsS1ltSVlKRE9rS2pIR2JuQys3KzVJd1J0ais4Um9uT0lSSnp3Vy9PZnUyUFJML1JmQmxmVmwxKzJJZlNWVQprbEVsRnlHNHRQL202ZjhWU2U5ZEpSZkFOWGRkcGdlOUd3dFlTbGsyaGI1aE5RTzVFSTAyVVYwdVVpWGcwNWRECkgyYkppQ1FQcHBxc3NiL09yNWQ5YXBSV3FMMzliQ0Z5Zi9GTzhZVVNYL0NEM1ZlZzhic24yWWc2bU14b2tUTGIKM1EvWll0NGthS0t0UVNreDV3NXh6bmZGNVBHenRIVmtSMkc0SVNRdDBVK2t1TVpRZTEyVCszS2ZqVitSVkZkZApWRkFpekxPOC82ZEFvRk5mQ3M4c2xsUDVEYXRLWnNXT2hROFJMZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gcp-cred
  namespace: demo
type: kubevault.com/gcp
data:
  sa.json: ZXlKMGVYQWlPaUcFpDSTZJa2hDQ0o5LmV5SmhkV1FpT2lKpPaTh2YzNSekxuZHBibVJ2ZDNNdWJtVjBM
```
