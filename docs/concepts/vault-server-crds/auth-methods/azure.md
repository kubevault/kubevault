---
title: Connect to Vault using Azure Auth Method
menu:
  docs_{{ .version }}:
    identifier: azure-auth-methods
    name: Azure
    parent: auth-methods-vault-server-crds
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Connect to Vault using Azure Auth Method

The KubeVault operator uses an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) to connect to an externally provisioned Vault server. For [Azure authentication](https://www.vaultproject.io/docs/auth/azure.html), it has to be [enabled](https://www.vaultproject.io/docs/auth/azure.html#via-the-cli-1) and [configured](https://www.vaultproject.io/docs/auth/azure.html#via-the-cli-1) in the Vault server. Follow the steps below to create an appropriate AppBinding:

- You have to specify the k8s secret name in `spec.secret` in the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).

- The specified secret must be in AppBinding's namespace.

- The type of the specified secret must be `"kubevault.com/azure"`.

- The specified secret data can have the following key:
  - `Secret.Data["msiToken"]` : `Required`. Signed JSON Web Token (JWT) from Azure MSI. Documentation can be found in [here](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview)

- The additional information required for the Azure authentication method can be provided as AppBinding's `spec.parameters`.
  
  ```yaml
  spec:
    parameters:
      path: my-azure
      vaultRole: demo-role
      azure:
        subscriptionID: 1bfc9f66-316d-433e-b13d-c55589f642ca
        resourceGroupName: vault-test
        vmName: test
        vmssName: test-set
  ```

  - `path` : `optional`. Specifies the path where Azure auth is enabled in Vault. If this path is not provided, the path will be set by default path `azure`.
  - `vaultRole` : `required`. Specifies the name of the Vault auth [role](https://www.vaultproject.io/api/auth/azure/index.html#create-role) against which login will be performed.
  - `azure.subscriptionID` : `optional`. Specifies the subscription ID for the machine that generated the MSI token. This information can be obtained through instance metadata.
  - `azure.resourceGroupName` : `optional`. Specifies the resource group for the machine that generated the MSI token. This information can be obtained through instance metadata.
  - `azure.vmName` : `optional`. Specifies the virtual machine name for the machine that generated the MSI token. This information can be obtained through instance metadata. If `vmssName` is provided, this value is ignored.
  - `azure.vmssName` : `optional`. Specifies the virtual machine scale set name for the machine that generated the MSI token. This information can be obtained through instance metadata.

Sample AppBinding and Secret is given below:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  secret:
    name: azure-cred
  parameters:
    path: my-azure
    vaultRole: demo-role
    azure:
      subscriptionID: 1bfc9f66-316d-433e-b13d-c55589f642ca
      resourceGroupName: vault-test
      vmName: test
      vmssName: test-set
  clientConfig:
    service:
      name: vault
      scheme: HTTPS
      port: 8200
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9UQTFNREl3T1RNeU5UaGFGdzB5T1RBME1qa3dPVE15TlRoYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdkVwTGVUT05ZRGxuTDkyMDlVQWY0Njc1CnR4MXpGL2J2UlZkejZjcVlYYWI1d0o0TmJoNGUxRFJDbmp6WXg1aDd0N1RLQXpvN3BWMWlzOFVHTTJUWGdPcloKa0hLQVJjOEFjekgxekE5Sk9mWkdGVk4xaXZBOTJHZ0xvVHNURDh4VTk0OWZ3Um8rYm5RemFqL2tLeHA5Z0puZQppYzBUenV5SEw2UTVseXRTVkoxWWhHSGdBdGM0eTBOcGZXZTZRekk1RnFXM2t1THFyaEVmRGV5TnR0UDUzaVV4ClpJN1IrbHBlVWsrY0NKZ2U0cTU1eGUvVmFpdEN6VVFIQVhLd0czRlNoYTc5cTBtb3J3N0ZIMW5YRDQwa0EwUXkKRlUxQTZtaDd5QVovb3BydWF0NFZjSDNwakdweGFvMzdLMGZQOXZXVjBvUnI5bm85YnM1MlI5L0I4cTVwYndJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFsd09iT3luNzZNMFRYMm53d1hlUkRreVdmNVczT3U4WGJQS3FEeXF1VnBIVUY2dHMKMmhuZFcyTXFJTEx3OUFYOXp1cFV4bzVKNEZadVNYYkNXUEt1a3VJbGtKYTRWMzAyeEtLNUJnT0R0TkdlQmNJQwpqNkFxZzUrU3dBVm9Va1FzcG52KzFBYVBOM0V1MldGNXlKaElaN0FnbjZERUYyWGlUQ3UvNm5lZVFLWC85QXVpCnNyTEZ2eE9TMEZuUTZ5YnpURGordHdTSGM4TkpHbWhKaGVOTVg5MmlYbHkrWG5VSmdRYU1tQzg5Y3JaRGRXczQKTnoyY0hMZFNZNnFNaHlSZGJ6NHZqdzlnVllMYXN5WjVFSTh2K1FmRTNOVE5SdlVJdlNwTVNQN09DT09UbTlteAprSlRVYjdCeW9DV0tzUDBGeWQ5RkdSVnZRQ0xFZi85MVNOeU5Sdz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K```
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: azure-cred
  namespace: demo
type: kubevault.com/azure
data:
  msiToken: ZXlKMGVYQWlPaUcFpDSTZJa2hDQ0o5LmV5SmhkV1FpT2lKpPaTh2YzNSekxuZHBibVJ2ZDNNdWJtVjBM=
```
