---
title: Vault-Operator Run
menu:
  docs_{{ .version }}:
    identifier: vault-operator-run
    name: Vault-Operator Run
    parent: reference-operator
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## vault-operator run

Launch KubeVault Webhook Server

```
vault-operator run [flags]
```

### Options

```
      --burst int                          The maximum burst for throttle (default 1000000)
      --cert-dir string                    Directory containing tls.crt + tls.key for the webhook server. Empty disables HTTPS (manager will self-sign).
      --cluster-name string                Name of cluster used in a multi-cluster setup
      --enable-mutating-webhook            If true, registers mutating webhooks for KubeVault CRDs. (default true)
      --enable-validating-webhook          If true, registers validating webhooks for KubeVault CRDs. (default true)
      --health-probe-bind-address string   The address the probe endpoint binds to. (default ":8081")
  -h, --help                               help for run
      --kubeconfig string                  Path to kubeconfig file with authorization information (the master location is set by the master flag).
      --label-key-blacklist strings        list of keys that are not propagated from a CRD object to its offshoots (default [app.kubernetes.io/name,app.kubernetes.io/version,app.kubernetes.io/instance,app.kubernetes.io/managed-by])
      --leader-elect                       Enable leader election for the webhook manager.
      --master string                      The address of the Kubernetes API server (overrides any value in kubeconfig)
      --metrics-bind-address string        The address the metric endpoint binds to. '0' disables.
      --qps float                          The maximum QPS to the master from this client (default 1e+06)
```

### Options inherited from parent commands

```
      --default-seccomp-profile-type string   Default seccomp profile
      --use-kubeapiserver-fqdn-for-aks        if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
```

### SEE ALSO

* [vault-operator](/docs/reference/operator/vault-operator.md)	 - Vault Operator by AppsCode - HashiCorp Vault Operator for Kubernetes

