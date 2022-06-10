---
title: Overview
menu:
  docs_{{ .version }}:
    identifier: unsealer-overview
    name: Overview
    parent: unsealer-vault-server-crds
    weight: 1
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Unsealer

[Unsealer](https://github.com/kubevault/unsealer) automates the process of [initializing](https://www.vaultproject.io/docs/commands/operator/init.html) and [unsealing](https://www.vaultproject.io/docs/concepts/seal.html#unsealing) Vault running in Kubernetes cluster. Also, it provides facilities to store unseal keys and root token in a secure way.

## Configuring Unsealer

To use Unsealer, configure `spec.unsealer` field in [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD .

```yaml
spec:
  unsealer:
    secretShares: <num_of_secret_shares>
    secretThresold: <num_of_secret_threshold>
    retryPeriodSeconds: <retry_period>
    overwriteExisting: <true/false>
    mode:
      ...
```

Here, we are going to describe the various attributes of the `spec.unsealer` field.

### unsealer.secretShares

`unsealer.secretShares` is an optional field that specifies the number of shares to split the master key into. It accepts integer value. The default vault is `5`.

```yaml
spec:
  unsealer:
    secretShares: 5
```

> Note: `unsealer.secretShares` must be greater than 1.

### unsealer.secretThreshold

`unsealer.secretThreshold` is an optional field that specifies the number of keys required to unseal vault. It accepts integer value. The default vault is `3`.

```yaml
spec:
  unsealer:
    secretThreshold: 2
```

> Note: `unsealer.secretThreshold` must be a positive integer and less than or equal to `unsealer.secretShares`.

### unsealer.retryPeriodSeconds

`unsealer.retryPeriodSeconds` is an optional field that specifies how often Unsealer will attempt to unseal the vault instance. It accepts integer value. The default vault is `10`.

```yaml
spec:
  unsealer:
    retryPeriodSeconds: 15
```

### unsealer.overwriteExisting

`unsealer.overwriteExisting` is an optional field that specifies Unsealer will overwrite existing unseal keys and root token(if have any). It accepts boolean value. Default vault is `false`.

```yaml
spec:
  unsealer:
    overwriteExisting: true
```

### unsealer.mode

`unsealer.mode` is a required field that specifies which mode to use to store unseal keys and root token.

```yaml
spec:
  unsealer:
    mode:
    ...
```

List of supported modes:

- [kubernetesSecret](/docs/concepts/vault-server-crds/unsealer/kubernetes_secret.md)
- [googleKmsGcs](/docs/concepts/vault-server-crds/unsealer/google_kms_gcs.md)
- [awsKmsSsm](/docs/concepts/vault-server-crds/unsealer/aws_kms_ssm.md)
- [azureKeyVault](/docs/concepts/vault-server-crds/unsealer/azure_key_vault.md)
