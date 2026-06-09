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
      --burst int                                               The maximum burst for throttle (default 1000000)
      --cluster-name string                                     Name of cluster used in a multi-cluster setup
      --docker-registry string                                  Docker image registry for sidecar, init-container, check-job, recovery-job and kubectl-job (default "kubevault")
      --gen-rotate-tls-recommendation-before-expiry-day int     Rotate TLS recommendation will be generated before given day of expiration.
      --gen-rotate-tls-recommendation-before-expiry-month int   Rotate TLS recommendation will be generated before given month of expiration. (default 1)
      --gen-rotate-tls-recommendation-before-expiry-year int    Rotate TLS recommendation will be generated before given year of expiration.
      --health-probe-bind-address string                        The address the probe endpoint binds to. (default ":8081")
  -h, --help                                                    help for operator
      --kubeconfig string                                       Path to kubeconfig file with authorization information (the master location is set by the master flag).
      --leader-elect                                            Enable leader election for controller manager.
      --license-file string                                     Path to license file
      --master string                                           The address of the Kubernetes API server (overrides any value in kubeconfig)
      --metrics-bind-address string                             The address the metric endpoint binds to. (default ":8080")
      --qps float                                               The maximum QPS to the master from this client (default 1e+06)
      --recommendation-resync-period duration                   Recommendation will be generated after every given duration based on the resource status at that moment. Default value is one hour (default 1h0m0s)
      --register-crds                                           If true, the operator applies its CustomResourceDefinitions on startup. Normally CRDs are installed by the helm chart and this should be left false.
      --resync-period duration                                  If non-zero, will re-list this often. (default 10m0s)
```

### Options inherited from parent commands

```
      --default-seccomp-profile-type string   Default seccomp profile
      --use-kubeapiserver-fqdn-for-aks        if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
```

### SEE ALSO

* [vault-operator](/docs/reference/operator/vault-operator.md)	 - Vault Operator by AppsCode - HashiCorp Vault Operator for Kubernetes

