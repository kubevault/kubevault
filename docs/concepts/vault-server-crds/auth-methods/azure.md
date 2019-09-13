---
title: Configure Azure Auth Method for Vault Server
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

# Configure Azure Auth Method for Vault Server

In Vault operator, usually Vault connection information are handled by [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md). For [Azure authentication](https://www.vaultproject.io/docs/auth/azure.html), it has to be [enabled](https://www.vaultproject.io/docs/auth/azure.html#via-the-cli-1) and [configured](https://www.vaultproject.io/docs/auth/azure.html#via-the-cli-1) in Vault. To perform this authenticaion:

- You have to specify `spec.secret` in the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).

- The type of the specified secret must be `"kubevault.com/azure"`.

- The specified secret data can have the following key:
    - `Secret.Data["msiToken"]` : `Required` - Signed JSON Web Token (JWT) from Azure MSI. Documentation can be found in [here](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview)

- The specified secret annotation can have the following key:
    - `Secret.Annotations["kubevault.com/azure.subscription-id"]`: `Optional`  - The subscription ID for the machine that generated the MSI token. This information can be obtained through instance metadata.
    - `Secret.Annotations["kubevault.com/azure.resource-group-name"]` : `Optional` - The resource group for the machine that generated the MSI token. This information can be obtained through instance metadata.
    - `Secret.Annotations["kubevault.com/azure.vm-name"]` : `Optional` - The virtual machine name for the machine that generated the MSI token. This information can be obtained through instance metadata. If vmss_name is provided, this value is ignored.
    - `Secret.Annotations["kubevault.com/azure.vmss-name"]` : `Optional` - The virtual machine scale set name for the machine that generated the MSI token. This information can be obtained through instance metadata.
    - `Secret.Annotations["kubevault.com/auth-path"]` : `Optional` - Specifies the path where Azure auth is enabled in Vault. If this path is not provided, the path will be set by default path "azure". If your Azure auth is enable some other path but "azure", you have to specify it here.

- The specified secret must be in AppBinding's namespace.

- You have to specify the name of Azure [role](https://www.vaultproject.io/api/auth/azure/index.html#create-role) name in `spec.parameters` of the [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md).
    ```yaml
    spec:
      parameters:
        policyControllerRole: my-azure-role # role name against which login will be done
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
    name: azure-cred
  parameters:
    policyControllerRole: my-azure-role
  clientConfig:
    service:
      name: vault
      scheme: HTTPS
      port: 8200
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9UQTFNREl3T1RNeU5UaGFGdzB5T1RBME1qa3dPVE15TlRoYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdkVwTGVUT05ZRGxuTDkyMDlVQWY0Njc1CnR4MXpGL2J2UlZkejZjcVlYYWI1d0o0TmJoNGUxRFJDbmp6WXg1aDd0N1RLQXpvN3BWMWlzOFVHTTJUWGdPcloKa0hLQVJjOEFjekgxekE5Sk9mWkdGVk4xaXZBOTJHZ0xvVHNURDh4VTk0OWZ3Um8rYm5RemFqL2tLeHA5Z0puZQppYzBUenV5SEw2UTVseXRTVkoxWWhHSGdBdGM0eTBOcGZXZTZRekk1RnFXM2t1THFyaEVmRGV5TnR0UDUzaVV4ClpJN1IrbHBlVWsrY0NKZ2U0cTU1eGUvVmFpdEN6VVFIQVhLd0czRlNoYTc5cTBtb3J3N0ZIMW5YRDQwa0EwUXkKRlUxQTZtaDd5QVovb3BydWF0NFZjSDNwakdweGFvMzdLMGZQOXZXVjBvUnI5bm85YnM1MlI5L0I4cTVwYndJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFsd09iT3luNzZNMFRYMm53d1hlUkRreVdmNVczT3U4WGJQS3FEeXF1VnBIVUY2dHMKMmhuZFcyTXFJTEx3OUFYOXp1cFV4bzVKNEZadVNYYkNXUEt1a3VJbGtKYTRWMzAyeEtLNUJnT0R0TkdlQmNJQwpqNkFxZzUrU3dBVm9Va1FzcG52KzFBYVBOM0V1MldGNXlKaElaN0FnbjZERUYyWGlUQ3UvNm5lZVFLWC85QXVpCnNyTEZ2eE9TMEZuUTZ5YnpURGordHdTSGM4TkpHbWhKaGVOTVg5MmlYbHkrWG5VSmdRYU1tQzg5Y3JaRGRXczQKTnoyY0hMZFNZNnFNaHlSZGJ6NHZqdzlnVllMYXN5WjVFSTh2K1FmRTNOVE5SdlVJdlNwTVNQN09DT09UbTlteAprSlRVYjdCeW9DV0tzUDBGeWQ5RkdSVnZRQ0xFZi85MVNOeU5Sdz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K```

```
```yaml
apiVersion: v1
data:
  msiToken: ZXlKMGVYQWlPaUcFpDSTZJa2hDQ0o5LmV5SmhkV1FpT2lKpPaTh2YzNSekxuZHBibVJ2ZDNNdWJtVjBM=
kind: Secret
metadata:
  name: azure-cred
  namespace: demo
  annotations:
      kubevault.com/auth-path: azure
      kubevault.com/azure.subscription-id: 1bfc9f66-316d-433e-b13d-c55589f642ca
      kubevault.com/azure.resource-group-name: vault-test
      kubevault.com/azure.vm-name: test
type: kubevault.com/azure
```
