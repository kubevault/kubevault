---
title: Vault Server Version | KubeVault Concepts
menu:
  docs_0.2.0:
    identifier: vaultserverversion-vault-server-crds
    name: Vault Server Version
    parent: vault-server-crds-concepts
    weight: 15
menu_name: docs_0.2.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# VaultServerVersion

`VaultServerVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It is a **non-namespaced** CRD. The name of this CRD will be used in `.spec.version` field of [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD. It provides a way to specify the docker images of Vault, Unsealer and Exporter.

Using a separate CRD for specifying respective docker images allow us to modify the images independently of Vault operator. This will also allow users to use their custom images.

```yaml
apiVersion: catalog.kubevault.com/v1alpha1
kind: VaultServerVersion
metadata:
  name: <name>
spec:
  ...
```

## VaultServerVersion Spec

VaultServerVersion `.spec` contains image information.

```yaml
apiVersion: catalog.kubevault.com/v1alpha1
kind: VaultServerVersion
metadata:
  name: 0.11.1
spec:
  vault:
    image: vault:0.11.1
  exporter:
    image: kubevault/vault-exporter: 0.1.0
  unsealer:
    image: kubevault/vault-unsealer: 0.2.0
  version: 0.11.1
```

`.spec` contains following fields:

### spec.version

`spec.version` is a required field that specifies the original version of Vault that has been used to build the docker image specified in `spec.vault.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the specified docker images are supported by the current Vault operator. The default value of this field is false.

### spec.vault.image

`spec.vault.image` is a required field that specifies the docker image which will be used for Vault.

```yaml
spec:
  vault:
    image: vault:0.11.1
```

### spec.unsealer.image

`spec.unsealer.image` is a required field that specifies the docker image which will be used for Unsealer.

```yaml
spec:
  unsealer:
    image: kubevault/vault-unsealer: 0.2.0
```

### spec.exporter.image

`spec.exporter.image` is a required field that specifies the docker image which will be used to export Prometheus metrics.

```yaml
spec:
  exporter:
    image: kubevault/vault-exporter: 0.1.0
```
