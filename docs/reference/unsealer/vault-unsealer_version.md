---
title: Vault-Unsealer Version
menu:
  docs_{{ .version }}:
    identifier: vault-unsealer-version
    name: Vault-Unsealer Version
    parent: reference-unsealer
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## vault-unsealer version

Prints binary version number.

```
vault-unsealer version [flags]
```

### Options

```
      --check string   Check version constraint
  -h, --help           help for version
      --short          Print just the version number.
```

### Options inherited from parent commands

```
      --enable-analytics                 Send analytical events to Google Analytics (default true)
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
```

### SEE ALSO

* [vault-unsealer](/docs/reference/unsealer/vault-unsealer.md)	 - Automates initialisation and unsealing of Hashicorp Vault

