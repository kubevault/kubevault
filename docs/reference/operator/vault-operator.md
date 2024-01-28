---
title: Vault-Operator
menu:
  docs_{{ .version }}:
    identifier: vault-operator
    name: Vault-Operator
    parent: reference-operator
    weight: 0

menu_name: docs_{{ .version }}
section_menu_id: reference
url: /docs/{{ .version }}/reference/operator/
aliases:
- /docs/{{ .version }}/reference/operator/vault-operator/
---
## vault-operator

Vault Operator by AppsCode - HashiCorp Vault Operator for Kubernetes

### Options

```
      --bypass-validating-webhook-xray        if true, bypasses validating webhook xray checks
      --default-seccomp-profile-type string   Default seccomp profile
  -h, --help                                  help for vault-operator
      --use-kubeapiserver-fqdn-for-aks        if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
```

### SEE ALSO

* [vault-operator operator](/docs/reference/operator/vault-operator_operator.md)	 - Launch Vault operator
* [vault-operator run](/docs/reference/operator/vault-operator_run.md)	 - Launch KubeVault Webhook Server
* [vault-operator version](/docs/reference/operator/vault-operator_version.md)	 - Prints binary version number.

