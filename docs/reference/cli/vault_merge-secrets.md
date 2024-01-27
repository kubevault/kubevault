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

merge-secrets merges secrets

### Synopsis

merge-secrets command merges two given secrets. Both the src & dst secrets must exist for successful merge operation.

Example: 
 # merge two secret name1 & name2 from ns1 & ns2 namespaces respectively
 $ kubectl vault merge-secrets --src=<ns1>/<name1> --dst=<ns2>/<name2>

 # --overwrite-keys flag will overwrite keys in destination if set to true.
 $ kubectl vault merge-secrets --src=<ns1>/<name1> --dst=<ns2>/<name2> --overwrite-keys=true


```
vault merge-secrets [flags]
```

### Options

```
      --dst string       destination secret.
  -h, --help             help for merge-secrets
      --overwrite-keys   will overwrite keys in destination if set to true.
      --src string       source secret.
```

### Options inherited from parent commands

```
      --as string                             Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray                  Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string                         UID to impersonate for the operation.
      --cache-dir string                      Default cache directory (default "/home/runner/.kube/cache")
      --certificate-authority string          Path to a cert file for the certificate authority
      --client-certificate string             Path to a client certificate file for TLS
      --client-key string                     Path to a client key file for TLS
      --cluster string                        The name of the kubeconfig cluster to use
      --context string                        The name of the kubeconfig context to use
      --default-seccomp-profile-type string   Default seccomp profile
      --disable-compression                   If true, opt-out of response compression for all requests to the server
      --insecure-skip-tls-verify              If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string                     Path to the kubeconfig file to use for CLI requests.
      --match-server-version                  Require server version to match client version
  -n, --namespace string                      If present, the namespace scope for this CLI request
      --request-timeout string                The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                         The address and port of the Kubernetes API server
      --tls-server-name string                Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                          Bearer token for authentication to the API server
      --user string                           The name of the kubeconfig user to use
```

### SEE ALSO

* [vault](/docs/reference/cli/vault.md)	 - KubeVault cli by AppsCode

