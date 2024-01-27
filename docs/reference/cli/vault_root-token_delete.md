---
title: Vault Root-Token Delete
menu:
  docs_{{ .version }}:
    identifier: vault-root-token-delete
    name: Vault Root-Token Delete
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## vault root-token delete

delete vault root-token

### Synopsis


$ kubectl vault root-token delete vaultserver <name> -n <namespace> [flags]

Examples:
 # delete the root-token with name set by --token-name flag
 $ kubectl vault root-token delete vaultserver vault -n demo --token-name <name>

 # default name for root-token will be used if --token-name flag is not provided
 # default root-token naming format: k8s.{cluster-name or UID}.{vault-namespace}.{vault-name}-root-token
 $ kubectl vault root-token delete vaultserver vault -n demo


```
vault root-token delete [flags]
```

### Options

```
  -h, --help                help for delete
      --token-name string   delete root-token with token-name. delete the latest root-token otherwise.
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

* [vault root-token](/docs/reference/cli/vault_root-token.md)	 - get, set, delete, sync, generate, and rotate root-token

