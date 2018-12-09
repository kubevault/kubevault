---
title: Monitor Vault Server using Builtin Prometheus Discovery
menu:
  product_vault-operator_0.1.0:
    identifier: vault-srever-using-builtin-prometheus-monitoring
    name: Builtin Prometheus Discovery
    parent: vault-monitor
    weight: 10
product_name: vault-operator
menu_name: product_vault-operator_0.1.0
section_menu_id: monitor
---

# Using Prometheus with Vault Server

This tutorial will show you how to monitor Vault server using [Prometheus](https://prometheus.io/).

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

## Monitor with builtin Prometheus

Below is the Vault server object created in this tutorial.

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: example
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
    agent: prometheus.io/builtin
    prometheus:
      namespace: demo
      labels:
        app: vault
      interval: 10s

```

Here,

- `spec.monitor` specifies that built-in [prometheus](https://github.com/prometheus/prometheus) is used to monitor this database instance.

Run following command to create example above.

```console
$ kubectl create -f https://raw.githubusercontent.comkubevault/operator/docs/examples/vault-server-builtin.yaml
vaultserver.kubevault.com/example created
```

Vault operator will configure its service once the Vault server is successfully running.

```console
$ kubectl get vs -n demo
NAME      NODES     VERSION   STATUS    AGE
example   1         0.11.1    Running   3h
```

Let's describe Service `example`

```console
$ kubectl get svc -n demo example -o yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    monitoring.appscode.com/agent: prometheus.io/builtin
    prometheus.io/path: /metrics
    prometheus.io/port: "9102"
    prometheus.io/scrape: "true"
  creationTimestamp: 2018-10-25T06:34:22Z
  labels:
    app: vault
    vault_cluster: example
  name: example
  namespace: demo
  ownerReferences:
  - apiVersion: kubevault.com/v1alpha1
    controller: true
    kind: VaultServer
    name: example
    uid: 00cb40e5-d820-11e8-b571-0800277d74b2
  resourceVersion: "7552"
  selfLink: /api/v1/namespaces/demo/services/example
  uid: 0268913f-d820-11e8-b571-0800277d74b2
spec:
  clusterIP: 10.97.239.21
  externalTrafficPolicy: Cluster
  ports:
  - name: client
    nodePort: 30922
    port: 8200
    protocol: TCP
    targetPort: 8200
  - name: cluster
    nodePort: 30321
    port: 8201
    protocol: TCP
    targetPort: 8201
  selector:
    app: vault
    vault_cluster: example
  sessionAffinity: None
  type: NodePort
status:
  loadBalancer: {}

```

You can see that the service contains following annotations.

```console
prometheus.io/path: /metrics
prometheus.io/port: "9102"
prometheus.io/scrape: "true"
```

The prometheus server will discover the service endpoint using these specifications and will scrape metrics from exporter.

## Deploy and configure Prometheus Server

The prometheus server is needed to configure so that it can discover endpoints of services. If a Prometheus server is already running in cluster and if it is configured in a way that it can discover service endpoints, no extra configuration will be needed.

If there is no existing Prometheus server running, rest of this tutorial will create a Prometheus server with appropriate configuration.

The configuration file of Prometheus server will be provided by ConfigMap. Create following ConfigMap with Prometheus configuration.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-server-conf
  labels:
    name: prometheus-server-conf
  namespace: demo
data:
  prometheus.yml: |-
    global:
      scrape_interval: 5s
      evaluation_interval: 5s
    scrape_configs:
    - job_name: 'kubernetes-service-endpoints'

      kubernetes_sd_configs:
      - role: endpoints

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: replace
        target_label: __scheme__
        regex: (https?)
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: kubernetes_namespace
      - source_labels: [__meta_kubernetes_service_name]
        action: replace
        target_label: kubernetes_name
```

Create above ConfigMap

```console
$ kubectl create -f https://raw.githubusercontent.comkubevault/operator/docs/examples/prom-server-conf.yaml
configmap/prometheus-server-conf created
```

Now, the below YAML is used to deploy Prometheus in kubernetes :

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus-server
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus-server
  template:
    metadata:
      labels:
        app: prometheus-server
    spec:
      containers:
        - name: prometheus
          image: prom/prometheus:v2.1.0
          args:
            - "--config.file=/etc/prometheus/prometheus.yml"
            - "--storage.tsdb.path=/prometheus/"
          ports:
            - containerPort: 9090
          volumeMounts:
            - name: prometheus-config-volume
              mountPath: /etc/prometheus/
            - name: prometheus-storage-volume
              mountPath: /prometheus/
      volumes:
        - name: prometheus-config-volume
          configMap:
            defaultMode: 420
            name: prometheus-server-conf
        - name: prometheus-storage-volume
          emptyDir: {}
```

Run the following command to deploy prometheus-server

```console
$ kubectl create -f https://raw.githubusercontent.comkubevault/operator/docs/examples/prometheus-builtin.yaml
clusterrole.rbac.authorization.k8s.io/prometheus-server created
serviceaccount/prometheus-server created
clusterrolebinding.rbac.authorization.k8s.io/prometheus-server created
deployment.apps/prometheus-server created
service/prometheus-service created
```

Wait until pods of the Deployment is running.

```console
$ kubectl get pods -n demo --selector=app=prometheus-server
NAME                                READY     STATUS    RESTARTS   AGE
prometheus-server                   1/1       Running   0          1m
```


And also verify RBAC stuffs

```console
$ kubectl get clusterrole prometheus-server -n demo
NAME                AGE
prometheus-server   1m
```

```console
$ kubectl get clusterrolebinding prometheus-server -n demo
NAME                AGE
prometheus-server   2m
```

### Prometheus Dashboard

Now open prometheus dashboard on browser by running `minikube service prometheus-service -n demo`.

Or you can get the URL of `prometheus-service` Service by running following command

```console
$ minikube service prometheus-service -n demo --url
http://192.168.99.100:30901
```

If you are not using minikube, browse prometheus dashboard using following address `http://{Node's ExternalIP}:{NodePort of prometheus-service}`.

Now, if you go the Prometheus Dashboard, you should see that this database endpoint as one of the targets.

<p align="center">
  <kbd>
    <img alt="builtin-prom-vault"  src="/docs/images/builtin-prom-vault.jpg">
  </kbd>
</p>

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete -n demo vs/example

$ kubectl delete clusterrole prometheus-server
$ kubectl delete clusterrolebindings  prometheus-server
$ kubectl delete serviceaccounts -n demo  prometheus-server
$ kubectl delete configmap -n demo prometheus-server-conf

$ kubectl delete ns demo
```

