---
title: Vault-Unsealer
menu:
  docs_0.1.0:
    identifier: vault-unsealer
    name: Vault-Unsealer
    parent: reference-unsealer
    weight: 0

menu_name: docs_0.1.0
section_menu_id: reference
url: /docs/0.1.0/reference/unsealer/
aliases:
- /docs/0.1.0/reference/unsealer/vault-unsealer/
---
## vault-unsealer

Automates initialisation and unsealing of Hashicorp Vault

### Synopsis

Automates initialisation and unsealing of Hashicorp Vault

### Options

```
      --alsologtostderr                  log to standard error as well as files
      --enable-analytics                 Send analytical events to Google Analytics (default true)
  -h, --help                             help for vault-unsealer
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [vault-unsealer run](/docs/reference/unsealer/vault-unsealer_run.md)	 - Launch Vault unsealer
* [vault-unsealer version](/docs/reference/unsealer/vault-unsealer_version.md)	 - Prints binary version number.

