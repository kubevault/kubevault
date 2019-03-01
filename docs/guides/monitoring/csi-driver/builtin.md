---
title: Monitor Vault CSI Driver using Builtin Prometheus Discovery
menu:
  docs_0.2.0:
    identifier: builtin-prometheus-csi-driver-monitoring
    name: Builtin Prometheus
    parent: csi-driver-monitoring
    weight: 10
menu_name: docs_0.2.0
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Monitor Vault CSI Driver with builtin Prometheus

This tutorial will show you how to configure builtin [Prometheus](https://github.com/prometheus/prometheus) scrapper to monitor Vault CSI driver.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

To keep Prometheus resources isolated, we are going to use a separate namespace to deploy Prometheus server.

```console
$ kubectl create ns monitoring
namespace/monitoring created
```

## Enable Monitoring in Vault CSI driver

Enable Prometheus monitoring using `prometheus.io/builtin` agent while install Vault CSI driver.  To know details about how to enable monitoring see [here](/docs/guides/monitoring/overview.md#how-to-enable-monitoring)

Here, we are going to enable monitoring for `operator` metrics.

<b> Using Helm: </b>

```console
$ helm install appscode/csi-vault --name csi-vault --version 0.2.0 --namespace kube-system \
  --set monitoring.agent=prometheus.io/builtin \
  --set monitoring.controller=true \
  --set monitoring.node=true \
  --set monitoring.prometheus.namespace=monitoring

```

<b> Using Script: </b>

```console
$ curl -fsSL https://raw.githubusercontent.com/kubevault/csi-driver/0.2.0/hack/deploy/install.sh | bash -s -- \
  --monitoring-agent=prometheus.io/builtin \
  --monitor-controller-plugin=true \
  --monitor-node-plugin=true \
  --prometheus-namespace=monitoring
```

This will add necessary annotations to `csi-vault-controller`, `csi-vault-node` services. Prometheus server will scrap metrics using those annotations. Let's check which annotations are added to the services,

```yaml
$ kubectl get svc csi-vault-controller -n kube-system -o yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "8443"
    prometheus.io/scheme: https
    prometheus.io/scrape: "true"
  creationTimestamp: "2018-12-28T06:32:51Z"
  labels:
    app: csi-vault
    chart: csi-vault-0.2.0
    component: csi-vault-attacher
    heritage: Tiller
    release: csi-vault
  name: csi-vault-controller
  namespace: kube-system
  resourceVersion: "8017"
  selfLink: /api/v1/namespaces/kube-system/services/csi-vault-controller
  uid: 66b553a8-0a6a-11e9-ae90-02e34d62ed30
spec:
  clusterIP: 10.101.189.99
  ports:
  - name: api
    port: 443
    protocol: TCP
    targetPort: 8443
  selector:
    app: csi-vault
    component: controller
    release: csi-vault
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}

$ kubectl get svc csi-vault-node -n kube-system -o yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "8443"
    prometheus.io/scheme: https
    prometheus.io/scrape: "true"
  creationTimestamp: "2018-12-28T06:32:51Z"
  labels:
    app: csi-vault
    chart: csi-vault-0.2.0
    component: csi-vault-plugin
    heritage: Tiller
    release: csi-vault
  name: csi-vault-node
  namespace: kube-system
  resourceVersion: "8018"
  selfLink: /api/v1/namespaces/kube-system/services/csi-vault-node
  uid: 66b67d38-0a6a-11e9-ae90-02e34d62ed30
spec:
  clusterIP: None
  ports:
  - name: api
    port: 443
    protocol: TCP
    targetPort: 8443
  selector:
    app: csi-vault
    component: csi-vault-node
    release: csi-vault
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}

$ kubectl get svc csi-vault-provisioner -n kube-system -o yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "8443"
    prometheus.io/scheme: https
    prometheus.io/scrape: "true"
  creationTimestamp: "2018-12-28T06:32:51Z"
  labels:
    app: csi-vault
    chart: csi-vault-0.2.0
    component: csi-vault-provisioner
    heritage: Tiller
    release: csi-vault
  name: csi-vault-provisioner
  namespace: kube-system
  resourceVersion: "8020"
  selfLink: /api/v1/namespaces/kube-system/services/csi-vault-provisioner
  uid: 66b8d309-0a6a-11e9-ae90-02e34d62ed30
spec:
  clusterIP: 10.111.29.109
  ports:
  - name: api
    port: 443
    protocol: TCP
    targetPort: 8443
  selector:
    app: csi-vault
    component: csi-vault-provisioner
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}
```

Here, `prometheus.io/scrape: "true"` annotation indicates that Prometheus should scrap metrics for this service.

The following three annotations point to api endpoints which provides operator specific metrics.

```console
prometheus.io/path: /metrics
prometheus.io/port: "8443"
prometheus.io/scheme: https
```

Now, we are ready to configure our Prometheus server to scrap those metrics.

## Deploy Prometheus Server

We have deployed Vault CSI driver in `kube-system` namespace. Vault exports driver metrics via TLS secured `api` endpoint. So, Prometheus server need to provide certificate while scrapping metrics from this endpoint. Vault CSI driver has created a secret named `csi-vault-apiserver-cert` with this certificate in `monitoring` namespaces as we have specified that we are going to deploy Prometheus in that namespace through `--prometheus-namespace` or `monitoring.prometheus.namespace` flag. We have to mount this secret in Prometheus deployment.

Let's check `csi-vault-apiserver-cert` secret has been created in `monitoring` namespace.

```console
$ kubectl get secret -n monitoring -l=app=csi-vault
NAME                       TYPE                DATA   AGE
csi-vault-apiserver-cert   kubernetes.io/tls   2      3h4m
```

#### Create `RBAC`

If you are using a `RBAC` enabled cluster, you have to provide necessary `RBAC` permissions for Prometheus. Following [this](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/README.md#deploy-prometheus-server), let's create `RBAC` stuffs for Prometheus by running:

```console
$ kubectl apply -f https://raw.githubusercontent.com/appscode/third-party-tools/master/monitoring/prometheus/builtin/artifacts/rbac.yaml
clusterrole.rbac.authorization.k8s.io/prometheus created
serviceaccount/prometheus created
clusterrolebinding.rbac.authorization.k8s.io/prometheus created
```
> YAML for RBAC resources can be found [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/artifacts/rbac.yaml).

#### Create `ConfigMap`

As we are monitoring Vault CSI driver, we should follow [this](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/README.md#kubernetes-apiservers) to create a ConfigMap. Bellow the YAML of ConfigMap that we are going to create in this tutorial

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  labels:
    name: prometheus-config
  namespace: monitoring
data:
  prometheus.yml: |-
    global:
      scrape_interval: 5s
      evaluation_interval: 5s
    scrape_configs:
    - job_name: 'csi-vault-controller'
      honor_labels: true
      kubernetes_sd_configs:
      - role: endpoints
      # Kubernetes apiserver serve metrics on a TLS secure endpoints. so, we have to use "https" scheme
      scheme: https
      # we have to provide certificate to establish tls secure connection
      tls_config:
        ca_file: /etc/prometheus/secret/csi-vault-apiserver-cert/tls.crt
        server_name: csi-vault-controller.kube-system.svc
      #  bearer_token_file is required for authorizating prometheus server to Kubernetes apiserver
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_label_app]
        separator: ;
        regex: csi-vault
        replacement: $1
        action: keep
      - source_labels: [__meta_kubernetes_service_label_component]
        separator: ;
        regex: controller
        replacement: $1
        action: keep
      - source_labels: [__meta_kubernetes_endpoint_address_target_kind, __meta_kubernetes_endpoint_address_target_name]
        separator: ;
        regex: Node;(.*)
        target_label: node
        replacement: ${1}
        action: replace
      - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: kube-system;csi-vault-controller;api
      - separator: ;
        regex: (.*)
        target_label: endpoint
        replacement: api
        action: replace
      - source_labels: [__meta_kubernetes_service_name]
        separator: ;
        regex: (.*)
        target_label: job
        replacement: ${1}
        action: replace

    - job_name: 'csi-vault-node'
      honor_labels: true
      kubernetes_sd_configs:
      - role: endpoints
      # Kubernetes apiserver serve metrics on a TLS secure endpoints. so, we have to use "https" scheme
      scheme: https
      # we have to provide certificate to establish tls secure connection
      tls_config:
        ca_file: /etc/prometheus/secret/csi-vault-apiserver-cert/tls.crt
        server_name: csi-vault-node.kube-system.svc
      #  bearer_token_file is required for authorizating prometheus server to Kubernetes apiserver
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_label_app]
        separator: ;
        regex: csi-vault
        replacement: $1
        action: keep
      - source_labels: [__meta_kubernetes_service_label_component]
        separator: ;
        regex: node
        replacement: $1
        action: keep
      - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: kube-system;csi-vault-node;api
      - source_labels: [__meta_kubernetes_endpoint_address_target_kind, __meta_kubernetes_endpoint_address_target_name]
        separator: ;
        regex: Node;(.*)
        target_label: node
        replacement: ${1}
        action: replace
      - source_labels: [__meta_kubernetes_service_name]
        separator: ;
        regex: (.*)
        target_label: job
        replacement: ${1}
        action: replace
      - separator: ;
        regex: (.*)
        target_label: endpoint
        replacement: api
        action: replace
```

Look at the `tls_config` field of `vault-apiservers` job. We have provided certificate file through `ca_file` field. This certificate comes from `csi-vault-apiserver-cert` that we are going to mount in Prometheus deployment. Here, `server_name` is used to verify hostname. In our case, the certificate is valid for hostname server, `csi-vault-controller.kube-system.svc`, `csi-vault-node.kube-system.svc`.

Let's create the ConfigMap we have shown above,

```console
$ kubectl apply -f https://raw.githubusercontent.com/kubevault/docs/master/docs/examples/monitoring/csi-driver/prom-builtin-conf.yaml
configmap/prometheus-config created

```

#### Deploy Prometheus

Now, we are ready to deploy Prometheus server. YAML for the deployment that we are going to create for Prometheus is shown below.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: prometheus-demo
  name: prometheus
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      containers:
      - args:
        - --config.file=/etc/prometheus/prometheus.yml
        - --storage.tsdb.path=/prometheus/
        image: prom/prometheus:v2.5.0
        imagePullPolicy: IfNotPresent
        name: prometheus
        ports:
        - containerPort: 9090
          protocol: TCP
        volumeMounts:
        - mountPath: /etc/prometheus/
          name: prometheus-config
        - mountPath: /prometheus/
          name: prometheus-storage
        - mountPath: /etc/prometheus/secret/csi-vault-apiserver-cert
          name: csi-vault-apiserver-cert
      serviceAccountName: prometheus
      volumes:
      - configMap:
          defaultMode: 420
          name: prometheus-config
        name: prometheus-config
      - emptyDir: {}
        name: prometheus-storage
      - name: csi-vault-apiserver-cert
        secret:
          defaultMode: 420
          secretName: csi-vault-apiserver-cert
          items:
          - path: tls.crt
            key: tls.crt
```

Notice that, we have mounted csi-vault-apiserver-cert secret as a volume at `/etc/prometheus/secret/csi-vault-apiserver-cert` directory.

Now, let's create the deployment,

```console
$ kubectl apply -f https://raw.githubusercontent.com/kubevault/docs/master/docs/examples/monitoring/csi-driver/prom-builtin-deployment.yaml
deployment.apps "prometheus" deleted
```

#### Verify Monitoring Metrics

Prometheus server is running on port 9090. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard. Run following commands on a separate terminal,

```console
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                          READY   STATUS    RESTARTS   AGE
prometheus-8568c86d86-vpzx5   1/1     Running   0          102s

$ kubectl port-forward -n monitoring prometheus-8568c86d86-vpzx5  9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at localhost:9090. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see the configured jobs as target and they are in UP state which means Prometheus is able collect metrics from them.

<p align="center">
  <kbd>
    <img alt="builtin-prom-vault"  src="/docs/images/monitoring/csi-vault-prom-builtin.png">
  </kbd>
</p>

## Cleaning up

To uninstall Prometheus server follow [this](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/README.md#cleanup)

To uninstall Vault CSI driver follow [this](https://github.com/kubevault/docs/blob/master/docs/setup/csi-driver/uninstall.md#uninstall-vault-csi-driver)

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete ns monitoring
```