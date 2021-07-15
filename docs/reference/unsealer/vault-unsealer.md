---
title: Vault-Unsealer
menu:
  docs_{{ .version }}:
    identifier: vault-unsealer
    name: Vault-Unsealer
    parent: reference-unsealer
    weight: 0

menu_name: docs_{{ .version }}
section_menu_id: reference
url: /docs/{{ .version }}/reference/unsealer/
aliases:
- /docs/{{ .version }}/reference/unsealer/vault-unsealer/
---
## vault-unsealer

Automates initialisation and unsealing of Hashicorp Vault

### Options

```
      --enable-analytics                 Send analytical events to Google Analytics (default true)
  -h, --help                             help for vault-unsealer
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
```

### SEE ALSO

* [vault-unsealer run](/docs/reference/unsealer/vault-unsealer_run.md)	 - Launch Vault unsealer
* [vault-unsealer version](/docs/reference/unsealer/vault-unsealer_version.md)	 - Prints binary version number.

