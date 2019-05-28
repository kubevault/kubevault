---
title: Monitoring Overview | KubeVault
description: A General Overview of Monitoring KubeVault Components
menu:
  docs_0.2.0:
    identifier: overview-monitoring
    name: Overview
    parent: monitoring-guides
    weight: 5
menu_name: docs_0.2.0
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Monitoring Vault Server

Vault operator has native support for monitoring via [Prometheus](https://prometheus.io/). You can use builtin [Prometheus](https://github.com/prometheus/prometheus) scrapper or [CoreOS Prometheus Operator](https://github.com/coreos/prometheus-operator) to monitor Vault operator. This tutorial will show you how this monitoring works with Vault operator and how to enable them.

## Overview

By default the Vault operator will configure each vault pod to publish [statsd](https://www.vaultproject.io/docs/configuration/telemetry.html) metrics.
The Vault operator runs a [statsd-exporter](https://github.com/kubevault/vault_exporter) container as sidecar to convert and expose those metrics in the format for Prometheus.
Following diagram shows the logical structure of Vault operator monitoring flow.

<p align="center">
  <img alt="Monitoring Structure"  src="/docs/images/vault-prometheus.jpg">
</p>

Each pod provides metrics at `/metrics` endpoint on port `9102`

## Operator Metrics

Following metrics are available for Vault server. These metrics are accessible through `api` endpoint of `vault-operator` service.

- vault_audit
- vault_audit_file
- vault_barrier
- vault_core
- vault_runtime
- vault_expire
- vault_merkle_flushdirty
- vault_merkle_savecheckpoint
- vault_policy
- vault_token
- vault_wal
- vault_rollback_attempt
- logshipper_streamWALs
- replication
- database
- database_error
- database_name
- database_named_error
- vault_storage_backend
- vault_provider_lock
- vault_consul
- vault_route
- vault_expire_num_leases
- vault_runtime_alloc_bytes
- vault_runtime_free_count
- vault_runtime_heap_objects
- vault_runtime_malloc_count
- vault_runtime_num_goroutines
- vault_runtime_sys_bytes
- vault_runtime_total_gc_pause_ns
- vault_runtime_total_gc_runs
- vault_runtime_gc_pause_ns

## How to Enable Monitoring

You can enable monitoring through some flags while installing or upgrading or updating. Vault operator via both `script` and `Helm`. You can chose which monitoring agent to use for monitoring. Vault operator will configure respective resources accordingly. Here, are the list of available flags and their uses,


|       Script Flag        |            Helm Values             |                     Acceptable Values                      |                                                         Default                                                         |                                                                                    Uses                                                                                    |
| ------------------------ | ---------------------------------- | ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--monitoring-agent`     | `monitoring.agent`                 | `prometheus.io/builtin` or `prometheus.io/coreos-operator` | `none`                                                                                                                  | Specify which monitoring agent to use for monitoring Vault operator.                                                                                                                |
| `--monitor-operator`  | `monitoring.operator`              | `true` or `false`                                          | `false`                                                                                                                 | Specify whether to monitor Vault operator.                                                                                                                                 |
| `--prometheus-namespace` | `monitoring.prometheus.namespace`  | any namespace                                              | same namespace as Vault operator                                                                                        | Specify the namespace where Prometheus server is running or will be deployed                                                                                               |
| `--servicemonitor-label` | `monitoring.serviceMonitor.labels` | any label                                                  | For Helm installation, `app: <generated app name>` and `release: <release name>`. For script installation, `app: vault-operator` | Specify the labels for ServiceMonitor. Prometheus crd will select ServiceMonitor using these labels. Only usable when monitoring agent is `prometheus.io/coreos-operator`. |

You have to provides these flags while installing or upgrading or updating Vault operator. Here, are examples for both script and Helm installation process are given which enable monitoring with `prometheus.io/coreos-operator` Prometheuse server for `operator` metrics.

**Helm:**
```console
$ helm install appscode/vault-operator --name vault-operator --version 0.2.0 --namespace kube-system \
  --set monitoring.agent=prometheus.io/coreos-operator \
  --set monitoring.operator=true \
  --set monitoring.prometheus.namespace=demo \
  --set monitoring.serviceMonitor.labels.k8s-app=prometheus
```

**Script:**
```console
$ curl -fsSL https://raw.githubusercontent.com/kubevault/operator/0.2.0/hack/deploy/install.sh  | bash -s -- \
  --monitoring-agent=prometheus.io/coreos-operator \
  --monitor-operator=true \
  --prometheus-namespace=demo \
  --servicemonitor-label=k8s-app=prometheus
```

## Next Steps

- Learn how to monitor Vault operator using built-in Prometheus from [here](/docs/guides/monitoring/builtin.md).
- Learn how to monitor Vault operator using CoreOS Prometheus operator from [here](/docs/guides/monitoring/coreos.md).
- Learn how to use Grafana dashboard to visualize monitoring data from [here](/docs/guides/monitoring/grafana.md).
