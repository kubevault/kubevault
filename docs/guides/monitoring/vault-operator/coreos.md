---
title: Monitor Vault Operator using Coreos Prometheus Operator
menu:
  product_vault-operator_0.1.0:
    identifier: vault-operator-using-coreos-prometheus-monitoring
    name: Coreos Prometheus Discovery
    parent: vault-monitor
    weight: 10
product_name: vault-operator
menu_name: product_vault-operator_0.1.0
section_menu_id: monitor
---

# Monitoring Vault Operator Using CoreOS Prometheus Operator

CoreOS [prometheus-operator](https://github.com/coreos/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use CoreOS Prometheus operator for monitoring Vault operator.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

- To keep Prometheus resources isolated, we are going to use a separate namespace to deploy Prometheus operator and respective resources.

```console
$ kubectl create ns monitoring
namespace/monitoring created
```

- We need a CoreOS prometheus-operator instance running. If you already don't have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/coreos-operator/README.md).

## Enable Monitoring in Vault operator

Enable Prometheus monitoring using `prometheus.io/coreos-operator` agent while installing Vault operator. To know details about how to enable monitoring see [here](/docs/guides/monitoring/overview.md#how-to-enable-monitoring)

Here, we are going to enable monitoring for `operator` metrics.

<b> Using Helm: </b>

```console
$ helm install appscode/vault-operator --name vault-operator --version 0.1.0 --namespace kube-system \
  --set monitoring.agent=prometheus.io/coreos-operator \
  --set monitoring.operator=true \
  --set monitoring.prometheus.namespace=monitoring \
  --set monitoring.serviceMonitor.labels.k8s-app=prometheus
```

<b> Using Script: </b>

```console
$ curl -fsSL https://raw.githubusercontent.com/kubevault/vault-operator/0.1.0/hack/deploy/install.sh  | bash -s -- \
  --monitoring-agent=prometheus.io/coreos-operator \
  --monitoring-operator=true \
  --prometheus-namespace=monitoring \
  --servicemonitor-label=k8s-app=prometheus
```

This will create a `ServiceMonitor` crd with name `vault-operator-servicemonitor` in `monitoring` namespace for monitoring endpoints of `vault-operator` service. This ServiceMonitor will have label `k8s-app: prometheus` provided by `--servicemonitor-label` flag. This label will be used by Prometheus crd to select this ServiceMonitor.

Let's check the ServiceMonitor crd using following command,

```yaml
$ kubectl get servicemonitors -n monitoring vault-operator-servicemonitor -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2018-12-26T11:13:25Z"
  generation: 1
  labels:
    k8s-app: prometheus
  name: vault-operator-servicemonitor
  namespace: monitoring
  resourceVersion: "32902"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/monitoring/servicemonitors/vault-operator-servicemonitor
  uid: 438a7cb5-08ff-11e9-852c-080027857726
spec:
  endpoints:
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    port: api
    scheme: https
    tlsConfig:
      caFile: /etc/prometheus/secrets/vault-operator-apiserver-cert/tls.crt
      serverName: vault-operator.kube-system.svc
  namespaceSelector:
    matchNames:
    - kube-system
  selector:
    matchLabels:
      app: vault-operator
      release: vault-operator
```

Here, `api` endpoint exports operator metrics.

Vault operator exports operator metrics via TLS secured `api` endpoint. So, Prometheus server need to provide certificate while scrapping metrics from this endpoint. Vault operator has created a secret named `vault-operator-apiserver-certs` with this certificate in `monitoring` namespace as we have specified that we are going to deploy Prometheus in that namespace through `--prometheus-namespace` flag. We have to specify this secret in Prometheus crd through `spec.secrets` field. Prometheus operator will mount this secret at `/etc/prometheus/secrets/vault-operator-apiserver-cert` directory of respective Prometheus pod. So, we need to configure `tlsConfig` field to use that certificate. Here, `caFile` indicates the certificate to use and serverName is used to verify hostname. In our case, the certificate is valid for hostname server and `vault-operator.kube-system.svc`.

Let's check secret vault-operator-apiserver-cert has been created in demo namespace.

```console
$ kubectl get secret -n monitoring -l=app=vault-operator
NAME                            TYPE                DATA   AGE
vault-operator-apiserver-cert   kubernetes.io/tls   2      8m27s
```

Also note that, there is a bearerTokenFile field. This file is token for the serviceaccount that will be created while creating RBAC stuff for Prometheus crd. This is required for authorizing Prometheus to scrape Vault operator API server.

Now, we are ready to deploy Prometheus server.

## Deploy Prometheus Server

In order to deploy Prometheus server, we have to create Prometheus crd. Prometheus crd defines a desired Prometheus server setup. For more details about Prometheus crd, please visit [here](https://github.com/coreos/prometheus-operator/blob/master/Documentation/design.md#prometheus).

If you are using a RBAC enabled cluster, you have to give necessary permissions to Prometheus. Check the documentation to see required RBAC permission from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/coreos-operator/README.md#deploy-prometheus-server).

#### Create Prometheus:

Below is the YAML of Prometheus crd that we are going to create for this tutorial,

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
  labels:
    k8s-app: prometheus
spec:
  replicas: 1
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      k8s-app: prometheus
  secrets:
  - vault-operator-apiserver-cert
  resources:
    requests:
      memory: 400Mi
```

Here, `spec.serviceMonitorSelector` is used to select the ServiceMonitor crd that is created by Vault operator. We have provided `vault-operator-apiserver-cert` secret in `spec.secrets` field. This will be mounted in Prometheus pod.

Let's create the Prometheus object we have shown above,

```console
$ kubectl apply -f https://raw.githubusercontent.com/kubevault/docs/master/docs/examples/monitoring/vault-operator/prom-coreos-crd.yaml
prometheus.monitoring.coreos.com/prometheus created
```

Prometheus operator watches for Prometheus `crd`. Once a Prometheus crd is created, Prometheus operator generates respective configuration and creates a `StatefulSet` to run Prometheus server.

Let's check `StatefulSet` has been created,

```console
$ kubectl get statefulset -n monitoring
NAME                    READY   AGE
prometheus-prometheus   1/1     31m
```

Check StatefulSet's pod is running,

```console
$ kubectl get pod prometheus-prometheus-0 -n monitoring
NAME                      READY   STATUS    RESTARTS   AGE
prometheus-prometheus-0   3/3     Running   1          31m
```

Now, we are ready to access Prometheus dashboard.

#### Verify Monitoring Metrics

Prometheus server is running on port 9090. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard. Run following commands on a separate terminal,

```console
$ kubectl port-forward -n monitoringprometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at localhost:9090. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see the configured jobs as target and they are in UP state which means Prometheus is able collect metrics from them.

<p align="center">
  <kbd>
    <img alt="builtin-prom-vault"  src="/docs/images/monitoring/vault-operator-coreos.png">
  </kbd>
</p>

## Cleaning up

To uninstall Vault operator follow [this](https://github.com/kubevault/docs/blob/master/docs/setup/operator/uninstall.md#uninstall-vault-operator).

To cleanup the Kubernetes resources created by this tutorial, run:

```console
# cleanup Prometheus resources
kubectl delete -n monitoring prometheus prometheus
kubectl delete -n monitoring secret vault-operator-apiserver-cert

$ kubectl delete ns monitoring
```