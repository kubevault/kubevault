---
title: Monitor Vault Server using Coreos Prometheus Operator
menu:
  product_vault-operator_0.1.0:
    identifier: vault-srever-using-coreos-prometheus-monitoring
    name: Coreos Prometheus Discovery
    parent: vault-monitor
    weight: 10
product_name: vault-operator
menu_name: product_vault-operator_0.1.0
section_menu_id: monitor
---

# Using Prometheus (CoreOS operator) with Vault Server

This tutorial will show you how to monitor Vault server using Prometheus via [CoreOS Prometheus Operator](https://github.com/coreos/prometheus-operator).

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [minikube](https://github.com/kubernetes/minikube).

Now, install Vault operator on your workstation by following the instructions from [here](/docs/setup/install.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace "demo" created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: Yaml files used in this tutorial are stored in [docs/examples](/docs/examples)

## Deploy CoreOS-Prometheus Operator

Run the following command to deploy CoreOS-Prometheus operator.

```console
$ kubectl create -f https://raw.githubusercontent.comkubevault/operator/docs/examples/prometheus-coreos-operator.yaml
namespace/demo created
clusterrole.rbac.authorization.k8s.io/prometheus-operator created
serviceaccount/prometheus-operator created
clusterrolebinding.rbac.authorization.k8s.io/prometheus-operator created
deployment.extensions/prometheus-operator created
```

Wait for running the Deployment’s Pods.

```console
$ kubectl get pods -n demo
NAME                                   READY     STATUS    RESTARTS   AGE
prometheus-operator-857455484c-7xwxt   1/1       Running   0          2m
```

This CoreOS-Prometheus operator will create some supported Custom Resource Definition (CRD).

```console
$ kubectl get crd
NAME                                        CREATED AT
alertmanagers.monitoring.coreos.com         2018-10-25T08:56:04Z
prometheuses.monitoring.coreos.com          2018-10-25T08:56:04Z
servicemonitors.monitoring.coreos.com       2018-10-25T08:56:04Z
```

Once the Prometheus operator CRDs are registered, run the following command to create a Prometheus.

```console
$ kubectl create -f https://raw.githubusercontent.comkubevault/operator/docs/examples/prometheus-coreos.yaml
clusterrole.rbac.authorization.k8s.io/prometheus created
serviceaccount/prometheus created
clusterrolebinding.rbac.authorization.k8s.io/prometheus created
prometheus.monitoring.coreos.com/prometheus created
service/prometheus created
```

Verify RBAC stuffs

```console
$ kubectl get clusterroles
NAME                      AGE
...
prometheus                42s
prometheus-operator       4m
...
```

```console
$ kubectl get clusterrolebindings
NAME                      AGE
...
prometheus                1m
prometheus-operator       5m
...
```

### Prometheus Dashboard

Now open prometheus dashboard on browser by running `minikube service prometheus -n demo`.

Or you can get the URL of `prometheus` Service by running following command

```console
$ minikube service prometheus -n demo --url
http://192.168.99.100:30900
```

If you are not using minikube, browse prometheus dashboard using following address `http://{Node's ExternalIP}:{NodePort of prometheus-service}`.

## Monitor Vault server with CoreOS Prometheus

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
    inmem: true
  unsealer:
    secretShares: 4
    secretThreshold: 2
    insecureTLS: true
    overwriteExisting: true
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
$ kubectl create -f https://raw.githubusercontent.comkubevault/operator/docs/examples/vault-server-coreos.yaml

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
    <img alt="prometheus-coreos"  src="/docs/images/coreos-prom-vault.png">
  </kbd>
</p>

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete -n demo vs/coreos-prom-postgres

$ kubectl delete clusterrolebindings prometheus-operator  prometheus
$ kubectl delete clusterrole prometheus-operator prometheus

$ kubectl delete ns demo
```