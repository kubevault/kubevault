## steward run

Run operator

### Synopsis


Run operator

```
steward run [flags]
```

### Options

```
      --ca-cert-file string           File containing CA certificate used by Vault server.
      --cluster-name string           Name of Kubernetes cluster used to create backends (default "kubernetes")
  -h, --help                          help for run
      --kubeconfig string             Path to kubeconfig file with authorization information (the master location is set by the master flag).
      --master string                 The address of the Kubernetes API server (overrides any value in kubeconfig)
      --resync-period duration        If non-zero, will re-list this often. Otherwise, re-list will be delayed aslong as possible (until the upstream source closes the watch or times out. (default 5m0s)
      --token-renew-period duration   Interval between consecutive attempts at renewing vault tokens. (default 1h0m0s)
      --vault-address string          Address of Vault server
      --vault-token string            Vault token used by operator.
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --analytics                        Send analytical events to Google Guard (default true)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [steward](steward.md)	 - Steward by AppsCode - HashiCorp Vault Operator for Kubernetes

