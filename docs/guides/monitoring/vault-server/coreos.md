---
title: Monitor Vault Server using CoreOS Prometheus Operator
menu:
  docs_0.1.0:
    identifier: coreos-vault-server-monitoring
    name: Prometheus Operator
    parent: vault-server-monitoring
    weight: 15
menu_name: docs_0.1.0
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Monitoring Vault Server Using CoreOS Prometheus Operator

The prometheus server is needed to configure so that it can discover endpoints of services. If a Prometheus server is already running in cluster and if it is configured in a way that it can discover service endpoints, no extra configuration will be needed.

If there is no existing Prometheus server running, [read this tutorial](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/coreos-operator/README.md) to see how to install [CoreOS Prometheus Operator](https://github.com/coreos/prometheus-operator).

This tutorial will show you how to monitor Vault server using Prometheus via [CoreOS Prometheus Operator](https://github.com/coreos/prometheus-operator).

## Monitor Vault server

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: exampleco
  namespace: demo
spec:
  nodes: 1
  version: "0.11.1"
  serviceTemplate:
    spec:
      type: NodePort
  backend:
    inmem: {}
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      kubernetesSecret:
        secretName: vault-keys
  monitor:
    agent: prometheus.io/coreos-operator
    prometheus:
      namespace: demo
      labels:
        app: vault
      interval: 10s

```

Here,

- `monitor.agent` indicates the monitoring agent. Currently only valid value currently is `coreos-prometheus-operator`
- `monitor.prometheus` specifies the information for monitoring by prometheus
  - `prometheus.namespace` specifies the namespace where ServiceMonitor is created.
  - `prometheus.labels` specifies the labels applied to ServiceMonitor.
  - `prometheus.port` indicates the port for PostgreSQL exporter endpoint (default is `56790`)
  - `prometheus.interval` indicates the scraping interval (eg, '10s')

Now create Vault server with monitoring spec

```console
$ kubectl create -f https://raw.githubusercontent.com/kubevault/docs/master/docs/examples/monitoring/vault-server/vault-server-coreos.yaml

```

Vault operator will create a ServiceMonitor object once the Vault server is successfully running.

```console
$ kubectl get servicemonitor -n demo
NAME                   AGE
vault-demo-exampleco   23s
```

Now, if you go the Prometheus Dashboard, you should see that this database endpoint as one of the targets.

<p align="center">
  <kbd>
    <img alt="prometheus-coreos"  src="/docs/images/monitoring/coreos-prom-vault.png">
  </kbd>
</p>

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete -n demo vs/coreos-prom-postgres

$ kubectl delete ns demo
```