---
title: Use Grafana dashboard to visualize data
menu:
  product_vault-operator_0.1.0:
    identifier: use-grafana-dashboard-to-visualize-data
    name: Use Grafana Dashboard
    parent: vault-monitor
    weight: 10
product_name: vault-operator
menu_name: product_vault-operator_0.1.0
section_menu_id: monitor
---

# Visualize Vault server data using Grafana dashboard

Grafana provides an elegant graphical user interface to visualize data. You can create beautiful dashboard easily with a meaningful representation of your Prometheus metrics.

If there is no grafana instance running on your cluster, then you can [read this tutorial](https://github.com/appscode/third-party-tools/blob/master/monitoring/grafana/README.md) to deploy one.


## Add Prometheus Data Source

We have to add our Prometheus server `prometheus-prometheus-0` as data source of grafana. We are going to use a `ClusterIP` service to connect Prometheus server with grafana. Let's create a service to select Prometheus server `prometheus-prometheus-0`,

```console
$ kubectl apply -f https://raw.githubusercontent.com/kubevault/operator/docs/examples/monitoring/coreos/prometheus-service.yaml
service/prometheus created
```

Below the YAML for the service we have created above,

```yaml
apiVersion: v1
kind: Service
metadata:
  name: prometheus
  namespace: monitoring
spec:
  type: ClusterIP
  ports:
  - name: web
    port: 9090
    protocol: TCP
    targetPort: 9090
  selector:
    app: prometheus
```

Now, follow these steps to add the Prometheus server as data source of Grafana UI.

1. From Grafana UI, go to `Configuration` option from sidebar and click on `Data Sources`.

    <p align="center">
      <img alt="Grafana: Data Sources"  src="/docs/images/monitoring/grafana-data-source-1.jpg" style="padding: 10px;">
    </p>

2. Then, click on `Add data source`.

    <p align="center">
      <img alt="Grafana: Add data source"  src="/docs/images/monitoring/grafana-data-source-2.png" style="padding: 10px;">
    </p>

3. Now, configure `Name`, `Type` and `URL` fields as specified below and keep rest of the configuration to their default value then click `Save&Test` button.
    - *Name: Vault-Operator* (you can give any name)
    - *Type: Prometheus*
    - *URL: http://prometheus.monitoring.svc:9090*
      (url format: http://{prometheus service name}.{namespace}.svc:{port})

    <p align="center">
      <img alt="Grafana: Configure data source"  src="/docs/images/monitoring/grafana-data-source-3.png" style="padding: 10px;">
    </p>

Once you have added Prometheus data source successfully, you are ready to create a dashboard to visualize the metrics.


## Cleanup
To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl delete -n demo service prometheus
```