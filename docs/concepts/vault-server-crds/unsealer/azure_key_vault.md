---
title: Azure Key Vault | Vault Unsealer
menu:
  docs_{{ .version }}:
    identifier: azure-key-vault-unsealer
    name: Azure Key Vault
    parent: unsealer-vault-server-crds
    weight: 1
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# mode.azureKeyVault

To use **azureKeyVault** mode specify `mode.azureKeyVault`. In this mode, unseal keys and root token will be stored in [Azure Key Vault](https://docs.microsoft.com/en-us/azure/key-vault/key-vault-overview) as secret.

```yaml
spec:
  unsealer:
    mode:
      azureKeyVault:
        vaultBaseUrl: <vault_base_url>
        tenantID: <tenant_id>
        clientCertSecret: <secret_name>
        aadClientSecret: <secret_name
        useManagedIdentity: <true/false>
        cloud: <cloud_environment_identifier>
```

`mode.azureKeyVault` has following field:

## azureKeyVault.vaultBaseUrl

`azureKeyVault.vaultBaseUrl` is a required field that specifies Azure key vault url.

```yaml
spec:
  unsealer:
    mode:
      azureKeyVault:
        vaultBaseUrl: "https://myvault.vault.azure.net"
```

## azureKeyVault.tenantID

`azureKeyVault.tenantID` is a required field that specifies Azure Active Directory tenant ID.

```yaml
spec:
  unsealer:
    mode:
      azureKeyVault:
        tenantID: "aaa-ddd-ffff-343455"
```

## azureKeyVault.clientCertSecret

`azureKeyVault.clientCertSecret` is an optional field that specifies the name of the secret containing client cert and client cert password. The secret contains following fields:

- `client-cert`
- `client-cert-password`

```yaml
spec:
  unsealer:
    mode:
      azureKeyVault:
        clientCertSecret: "azure-client-cert-cred"
```

## azureKeyVault.addClientSecret

`azureKeyVault.addClientSecret` is an optional field that specifies the name of the secret containing client id and client secret of AAD application. The secret contains following fields:

- `client-id`
- `client-secret`

```yaml
spec:
  unsealer:
    mode:
      azureKeyVault:
        addClientSecret: "azure-add-client-cred"
```

## azureKeyVault.useManageIdentity

`azureKeyVault.useManageIdentity` is an optional field that specifies to use managed service identity for the virtual machine. 

```yaml
spec:
  unsealer:
    mode:
      azureKeyVault:
        useManageIdentity: true
```

> Note: `azureKeyVault.clientCertSecret` or `azureKeyVault.addClientSecret` or `azureKeyVault.useManageIdentity` has to be specified.

## azureKeyVault.cloud

`azureKeyVault.cloud` is an optional field that specifies the cloud environment identifier. If it is not specified, then `AZUREPUBLICCLOUD` will be used as default.

```yaml
spec:
  unsealer:
    mode:
      azureKeyVault:
        cloud: "AZUREGERMANCLOUD"
```
