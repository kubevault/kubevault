---
title: Monitor Vault CSI Driver using Coreos Prometheus Operator
menu:
  product_vault-csi-driver_0.1.0:
    identifier: vault-csi-driver-using-coreos-prometheus-monitoring
    name: Coreos Prometheus Discovery
    parent: vault-monitor
    weight: 10
product_name: csi-vault
menu_name: product_vault-csi-driver_0.1.0
section_menu_id: monitor
---

# Monitoring Vault CSI Driver Using CoreOS Prometheus Operator

CoreOS [prometheus-operator](https://github.com/coreos/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use CoreOS Prometheus operator for monitoring Vault CSI driver.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

- To keep Prometheus resources isolated, we are going to use a separate namespace to deploy Prometheus operator and respective resources.

```console
$ kubectl create ns monitoring
namespace/monitoring created
```

- We need a CoreOS prometheus-operator instance running. If you already don't have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/coreos-operator/README.md).

## Enable Monitoring in Vault CSI driver

Enable Prometheus monitoring using `prometheus.io/coreos-operator` agent while installing Vault CSI driver. To know details about how to enable monitoring see [here](/docs/guides/monitoring/overview.md#how-to-enable-monitoring)

Here, we are going to enable monitoring for `operator` metrics.

<b> Using Helm: </b>

```console
$ helm install appscode/csi-vault --name csi-vault --version 0.1.0 --namespace kube-system \
  --set monitoring.agent=prometheus.io/coreos-operator \
  --set monitoring.attacher=true \
  --set monitoring.plugin=true \
  --set monitoring.provisioner=true \
  --set monitoring.prometheus.namespace=monitoring \
  --set monitoring.serviceMonitor.labels.k8s-app=prometheus
```

<b> Using Script: </b>

```console
$ curl -fsSL https://raw.githubusercontent.com/kubevault/csi-driver/0.1.0/hack/deploy/install.sh | bash -s -- \
  --monitoring-agent=prometheus.io/coreos-operator \
  --monitor-attacher=true \
  --monitor-plugin=true \
  --monitor-provisioner=true \
  --prometheus-namespace=monitoring \
  --servicemonitor-label=k8s-app=prometheus
```

This will create three `ServiceMonitor` crds with name `csi-vault-attacher-servicemonitor`, `csi-vault-plugin-servicemonitor` and `csi-vault-provisioner-servicemonitor`  in `monitoring` namespace for monitoring endpoints of `csi-vault-attacher`, `csi-vault-plugin` and `csi-vault-provisioner` services. These ServiceMonitor will have label `k8s-app: prometheus` provided by `--servicemonitor-label` flag. This label will be used by Prometheus crd to select this ServiceMonitor.

Let's check the ServiceMonitor crd using following command,

```yaml
$ kubectl get servicemonitors -n monitoring
NAME                                   AGE
csi-vault-attacher-servicemonitor      4m
csi-vault-plugin-servicemonitor        4m
csi-vault-provisioner-servicemonitor   4m
```

Vault CSI driver exports driver metrics via TLS secured `api` endpoint. So, Prometheus server need to provide certificate while scrapping metrics from this endpoint. Vault CSI driver has created a secret named `csi-vault-apiserver-cert` with this certificate in `monitoring` namespace as we have specified that we are going to deploy Prometheus in that namespace through `--prometheus-namespace` flag. We have to specify this secret in Prometheus crd through `spec.secrets` field. Prometheus operator will mount this secret at `/etc/prometheus/secrets/csi-vault-apiserver-cert` directory of respective Prometheus pod. So, we need to configure `tlsConfig` field to use that certificate. Here, `caFile` indicates the certificate to use and serverName is used to verify hostname. In our case, the certificate is valid for hostname server, `csi-vault-attacher.kube-system.svc`, `csi-vault-plugin.kube-system.svc` and `csi-vault-provisioner.kube-system.svc`.

Let's check secret csi-vault-apiserver-cert has been created in monitoring namespace.

```console
$ kubectl get secrets -n monitoring -l=app=csi-vault
NAME                       TYPE                DATA   AGE
csi-vault-apiserver-cert   kubernetes.io/tls   2      23m
```

Also note that, there is a bearerTokenFile field. This file is token for the serviceaccount that will be created while creating RBAC stuff for Prometheus crd. This is required for authorizing Prometheus to scrape Vault CSI driver.

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
  - csi-vault-apiserver-cert
  resources:
    requests:
      memory: 400Mi
```

Here, `spec.serviceMonitorSelector` is used to select the ServiceMonitor crd that is created by Vault CSI driver. We have provided `csi-vault-apiserver-cert` secret in `spec.secrets` field. This will be mounted in Prometheus pod.

Let's create the Prometheus object we have shown above,

```console
$ kubectl apply -f https://raw.githubusercontent.com/kubevault/docs/master/docs/examples/monitoring/csi-driver/prom-coreos-crd.yaml
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
    <img alt="builtin-prom-vault"  src="/docs/images/monitoring/csi-vault-prom-coreos.png">
  </kbd>
</p>

## Cleaning up

To uninstall Vault CSI driver follow [this](https://github.com/kubevault/docs/blob/master/docs/setup/csi-driver/uninstall.md#uninstall-vault-csi-driver).

To cleanup the Kubernetes resources created by this tutorial, run:

```console
# cleanup Prometheus resources
kubectl delete -n monitoring prometheus prometheus
kubectl delete -n monitoring secret csi-vault-apiserver-cert

$ kubectl delete ns monitoring
```