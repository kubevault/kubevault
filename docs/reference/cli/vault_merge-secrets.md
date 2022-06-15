---
title: Vault Merge-Secrets
menu:
  docs_{{ .version }}:
    identifier: vault-merge-secrets
    name: Vault Merge-Secrets
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## vault merge-secrets

Merge Secrets

### Synopsis

merge-secrets command merges two given secrets. Both the src & dst secrets must exist for successful merge operation.

```bash
Example: 
 # merge two secret name1 & name2 from ns1 & ns2 namespaces respectively
 $ kubectl vault merge-secrets --src=<ns1>/<name1> --dst=<ns2>/<name2>

 # --overwrite-keys flag will overwrite keys in destination if set to true.
 $ kubectl vault merge-secrets --src=<ns1>/<name1> --dst=<ns2>/<name2> --overwrite-keys=true
```

```
vault generate [flags]
```

### Options

```
  -f, --filename strings            Filename, directory, or URL to files identifying the resource to update
  -h, --help                        help for generate
      --keys stringToString         Key/Value map used to store the keys to read and their mapping keys. secretKey=objectName (default [])
  -k, --kustomize string            Process the kustomization directory. This flag can't be used together with -f or -R.
  -o, --output string               output format yaml/json. default to yaml
  -R, --recursive                   Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.
  -b, --secretrolebinding string    secret role binding. namespace/name
  -p, --vault-ca-cert-path string   vault CA cert path in secret provider, default to Insecure mode.
  -r, --vaultrole string            vault role. RoleKind/name
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --analytics                        Send analytical events to Google Analytics (default true)
      --as string                        Username to impersonate for the operation
      --as-group stringArray             Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string                 Default cache directory (default "/home/runner/.kube/cache")
      --certificate-authority string     Path to a cert file for the certificate authority
      --client-certificate string        Path to a client certificate file for TLS
      --client-key string                Path to a client key file for TLS
      --cluster string                   The name of the kubeconfig cluster to use
      --context string                   The name of the kubeconfig context to use
      --insecure-skip-tls-verify         If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string                Path to the kubeconfig file to use for CLI requests.
      --log-backtrace-at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log-dir string                   If non-empty, write log files in this directory
      --log-flush-frequency duration     Maximum number of seconds between log flushes (default 5s)
      --logtostderr                      log to standard error instead of files
      --match-server-version             Require server version to match client version
  -n, --namespace string                 If present, the namespace scope for this CLI request
      --request-timeout string           The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                    The address and port of the Kubernetes API server
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 0)
      --tls-server-name string           Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                     Bearer token for authentication to the API server
      --user string                      The name of the kubeconfig user to use
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [vault](/docs/reference/cli/vault.md)	 - KubeVault cli by AppsCode

