---
title: Vault Get Root Token
menu:
  docs_{{ .version }}:
    identifier: vault-get-root-token
    name: Vault Get Root Token
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## vault get-root-token

You don't need to install separate tools for each of the cloud providers. You can get the decrypted `root-token` for `VaultServer` for various vault auto unsealing options e.g: `GCS`, `AWS`, `Azure`, or even `K8s secret` using the `KubeVault CLI`. All you need, is to provide the `VaultServer` reference, `name` and `namespace` in the command.
Example command is shown below:

```bash
 # Get the decrypted root-token for resource vaultserver with name vault and namespace demo.
 # Prints the root-token name and the value.
 $ kubectl vault get-root-token vaultserver -n demo vault
   k8s.kubevault.com.demo.vault-root-token: s.OiObHyEa42i2DE7E6nFNZNIQ

 # Provide flag --value-only if you want to print only the root-token value.
 $ kubectl vault get-root-token vaultserver -n demo vault --value-only
   s.OiObHyEa42i2DE7E6nFNZNIQ
```

### Options

```
  -f, --filename strings   Filename, directory, or URL to files identifying the resource to update
  -h, --help               help for approve
  -k, --kustomize string   Process the kustomization directory. This flag can't be used together with -f or -R.
  -R, --recursive          Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.
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

