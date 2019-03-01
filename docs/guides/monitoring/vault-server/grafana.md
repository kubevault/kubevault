---
title: Grafana dashboard for Vault Server
menu:
  docs_0.2.0:
    identifier: grafana-vault-server-monitoring
    name: Grafana Dashboard
    parent: vault-server-monitoring
    weight: 20
menu_name: docs_0.2.0
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Visualize Vault server data using Grafana dashboard

Grafana provides an elegant graphical user interface to visualize data. You can create beautiful dashboard easily with a meaningful representation of your Prometheus metrics.

If there is no grafana instance running on your cluster, then you can [read this tutorial](https://github.com/appscode/third-party-tools/blob/master/monitoring/grafana/README.md) to deploy one.


## Add Prometheus Data Source

We have to add our Prometheus server `prometheus-prometheus-0` as data source of grafana. We are going to use a `ClusterIP` service to connect Prometheus server with grafana. Let's create a service to select Prometheus server `prometheus-prometheus-0`,

```console
$ kubectl apply -f https://raw.githubusercontent.com/kubevault/docs/master/docs/examples/monitoring/vault-server/prometheus-service.yaml
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

## Import Vault server Dashboard

Vault server comes with a pre-configured Grafana dashboard. You can download json configuration of the dashboard from [here](/docs/examples/monitoring/grafana/dashboard.json).

Follow these steps to import the preconfigured stash dashboard,

1. From Grafana UI, go to `Create` option from sidebar and click on `import`.

    <p align="center">
        <img alt="Grafana: Import dashboard"  src="/docs/images/monitoring/grafana-import-1.png" style="padding: 10px;">
    </p>

2. Then, paste `json` from [here](/docs/examples/monitoring/grafana/dashboard.json) or upload `json` configuration file of the dashboard using `Upload .json File` button.

    <p align="center">
      <img alt="Grafana: Provide dashboard ID"  src="/docs/images/monitoring/grafana-import-2.png" style="padding: 10px;">
    </p>

3. Now on `prometheus-infra` field, select the data source name that we have given to our Prometheus data source earlier. Then click on `Import` button.

    <p align="center">
        <img alt="Grafana: Select data source"  src="/docs/images/monitoring/grafana-import-4.png" style="padding: 10px;">
    </p>

Once you have imported the dashboard successfully, you will be greeted with dashboard.

<p align="center">
      <img alt="Grafana: Stash dashboard"  src="/docs/images/monitoring/grafana-import-3.png" style="padding: 10px;">
</p>


## Cleanup
To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl delete -n demo service prometheus
```