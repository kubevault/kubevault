---
title: Monitor Vault Server using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: coreos-vault-server-monitoring
    name: Prometheus Operator
    parent: vault-server-monitoring
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Monitoring Vault Server Using Prometheus Operator

CoreOS [prometheus-operator](https://github.com/coreos/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to monitor Vault server using Prometheus via Prometheus Operator).

## Monitor Vault server

To enable monitoring, configure `spec.monitor` field in a `VaultServer` custom resource. Below is an example:

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: exampleco
  namespace: demo
spec:
  replicas: 1
  version: "1.2.0"
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

- `monitor.agent` indicates the monitoring agent `coreos-prometheus-operator`.
- `monitor.prometheus` specifies the information for monitoring by Prometheus.
  - `prometheus.namespace` specifies the namespace where ServiceMonitor is created.
  - `prometheus.labels` specifies the labels applied to ServiceMonitor.
  - `prometheus.port` indicates the port for Vault statsd exporter endpoint (default is `56790`)
  - `prometheus.interval` indicates the scraping interval (eg, '10s')

Now create Vault server with the monitoring spec

```console
$ kubectl create -f https://github.com/kubevault/docs/raw/{{< param "info.version" >}}/docs/examples/monitoring/vault-server/vault-server-coreos.yaml

```

KubeVault operator will create a ServiceMonitor object once the Vault server is successfully running.

```console
$ kubectl get servicemonitor -n demo
NAME                   AGE
vault-demo-exampleco   23s
```

Now, if you go the Prometheus Dashboard, you should see that this Vault endpoint as one of the targets.

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