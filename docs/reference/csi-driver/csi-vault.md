---
title: Csi-Vault
menu:
  docs_{{ .version }}:
    identifier: csi-vault
    name: Csi-Vault
    parent: reference-csi-driver
    weight: 0

menu_name: docs_{{ .version }}
section_menu_id: reference
url: /docs/{{ .version }}/reference/csi-driver/
aliases:
- /docs/{{ .version }}/reference/csi-driver/csi-vault/
---
## csi-vault

Vault CSI by Appscode - Start farms

### Synopsis

Vault CSI by Appscode - Start farms

### Options

```
      --alsologtostderr                  log to standard error as well as files
      --enable-analytics                 Send analytical events to Google Analytics (default true)
  -h, --help                             help for csi-vault
      --log-flush-frequency duration     Maximum number of seconds between log flushes (default 5s)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files (default true)
      --stderrthreshold severity         logs at or above this threshold go to stderr
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [csi-vault run](/docs/reference/csi-driver/csi-vault_run.md)	 - Run Vault CSI driver
* [csi-vault version](/docs/reference/csi-driver/csi-vault_version.md)	 - Prints binary version number.

