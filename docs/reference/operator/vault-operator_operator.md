---
title: Vault-Operator Operator
menu:
  docs_{{ .version }}:
    identifier: vault-operator-operator
    name: Vault-Operator Operator
    parent: reference-operator
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## vault-operator operator

Launch Vault operator

```
vault-operator operator [flags]
```

### Options

```
      --burst int                          The maximum burst for throttle (default 1000000)
      --cluster-name string                Name of cluster used in a multi-cluster setup
      --docker-registry string             Docker image registry for sidecar, init-container, check-job, recovery-job and kubectl-job (default "kubevault")
      --health-probe-bind-address string   The address the probe endpoint binds to. (default ":8081")
  -h, --help                               help for operator
      --kubeconfig string                  Path to kubeconfig file with authorization information (the master location is set by the master flag).
      --leader-elect                       Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.
      --license-file string                Path to license file
      --master string                      The address of the Kubernetes API server (overrides any value in kubeconfig)
      --metrics-bind-address string        The address the metric endpoint binds to. (default ":8080")
      --qps float                          The maximum QPS to the master from this client (default 1e+06)
      --resync-period duration             If non-zero, will re-list this often. Otherwise, re-list will be delayed aslong as possible (until the upstream source closes the watch or times out. (default 10m0s)
```

### Options inherited from parent commands

```
      --bypass-validating-webhook-xray   if true, bypasses validating webhook xray checks
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
```

### SEE ALSO

* [vault-operator](/docs/reference/operator/vault-operator.md)	 - Vault Operator by AppsCode - HashiCorp Vault Operator for Kubernetes

