# Unsealer

[Unsealer](https://github.com/kubevault/unsealer) automates the process of [initializing](https://www.vaultproject.io/docs/commands/operator/init.html) and [unsealing](https://www.vaultproject.io/docs/concepts/seal.html#unsealing) Vault running in Kubernetes cluster. Also it provides facilities to store unseal keys and root token in a secure way.


## spec.unsealer
To use Unsealer specify `spec.unsealer` in [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD .

```yaml
spec:
  unsealer:
    secretShares: <num_of_secret_shares>
    secretThresold: <num_of_secret_threshold>
    retryPeriodSeconds: <retry_period>
    overwriteExisting: <true/false>
    insecureSkipTLSVerify: <true/false>
    caBundle: <pem_encoded_ca_cert>
    mode:
      ...
```

#### unsealer.secretShares

`unsealer.secretShares` is an optional field that specifies the number of shares to split the master key into. It accepts integer value. Default vault is `5`.

```yaml
spec:
  unsealer:
    secretShares: 4
```

> Note: `unsealer.secretShares` must be greater than 1.

#### unsealer.secretThreshold

`unsealer.secretThreshold` is an optional field that specifies the number of keys required to unseal vault. It accepts integer value. Default vault is `3`.

```yaml
spec:
  unsealer:
    secretThreshold: 2
```
> Note: `unsealer.secretThreshold` must be a positive interger and less than or equal to `unsealer.secretShares`.

#### unsealer.retryPeriodSeconds

`unsealer.retryPeriodSeconds` is an optional field that specifies how often Unsealer will attempt to unseal the vault instance. It accepts integer value. Default vault is `10`.

```yaml
spec:
  unsealer:
    retryPeriodSeconds: 15
```

#### unsealer.insecureSkipTLSVerify

`unsealer.insecureSkipTLSVerify` is an optional field that specifies whether skip tls verification when Unsealer communicating with Vault server. It accepts boolean value. Default vault is `false`.

```yaml
spec:
  unsealer:
    insecureSkipTLSVerify: true
```

#### unsealer.caBundle

`unsealer.caBundle` is an optional field that specifies the CA cert that will be used to verify Vault server certificates when Unsealer communicating with it.

```yaml
spec:
  unsealer:
    caBundle: <pem_encoded_ca_cert>
```

#### unsealer.overwriteExisting

`unsealer.overwriteExisting` is an optional field that specifies Unsealer will overwrite existing unseal keys and root token(if have any). It accepts boolean value. Default vault is `false`.

```yaml
spec:
  unsealer:
    overwriteExisting: true
```

#### unsealer.mode

`unsealer.mode` is a required field that specifies which mode to use to store unseal keys and root token.

```yaml
spec:
  unsealer:
    mode:
    ...
```

List of supported mode:

- [kubernetesSecret](/docs/concepts/vault-server-crds/unsealer/kubernetes_secret.md)
- [googleKmsGcs](/docs/concepts/vault-server-crds/unsealer/google_kms_gcs.md)
- [awsKmsSsm](/docs/concepts/vault-server-crds/unsealer/aws_kms_ssm.md)
- [azureKeyVault](/docs/concepts/vault-server-crds/unsealer/azure_key_vault.md)
