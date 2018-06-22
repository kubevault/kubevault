# Vault Operator
[Vault Operator by AppsCode](https://github.com/kubevault/operator) - HashiCorp Vault Operator for Kubernetes

## TL;DR;

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm install appscode/vault-operator
```

## Introduction

This chart bootstraps a [HashiCorp Vault controller](https://github.com/kubevault/operator) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.9+

## Installing the Chart
To install the chart with the release name `my-release`:

```console
$ helm install appscode/vault-operator --name my-release
```

The command deploys Vault operator on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release`:

```console
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Vault chart and their default values.


| Parameter                             | Description                                                        | Default            |
| ------------------------------------- | ------------------------------------------------------------------ | ------------------ |
| `replicaCount`                        | Number of Vault operator replicas to create (only 1 is supported)  | `1`                |
| `operator.registry`                   | Docker registry used to pull Vault operator image                  | `vault-operator`   |
| `operator.repository`                 | Vault operator container image                                     | `pack-operator`    |
| `operator.tag`                        | Vault operator container image tag                                 | `canary`           |
| `imagePullPolicy`                     | container image pull policy                                        | `IfNotPresent`     |
| `criticalAddon`                       | If true, installs Vault operator as critical addon                 | `false`            |
| `logLevel`                            | Log level for operator                                             | `3`                |
| `nodeSelector`                        | Node labels for pod assignment                                     | `{}`               |
| `rbac.create`                         | If `true`, create and use RBAC resources                           | `true`             |
| `serviceAccount.create`               | If `true`, create a new service account                            | `true`             |
| `serviceAccount.name`                 | Service account to be used. If not set and `serviceAccount.create` is `true`, a name is generated using the fullname template | `` |
| `apioperator.groupPriorityMinimum`    | The minimum priority the group should have.                        | 10000              |
| `apioperator.versionPriority`         | The ordering of this API inside of the group.                      | 15                 |
| `apioperator.enableValidatingWebhook` | Enable validating webhooks for Kubernetes workloads                | false              |
| `apioperator.enableMutatingWebhook`   | Enable mutating webhooks for Kubernetes workloads                  | false              |
| `apioperator.ca`                      | CA certificate used by main Kubernetes api operator                | ``                 |
| `enableAnalytics`                     | Send usage events to Google Analytics                              | `true`             |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example:

```console
$ helm install --name my-release --set image.tag=v0.2.1 appscode/vault-operator
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while
installing the chart. For example:

```console
$ helm install --name my-release --values values.yaml appscode/vault-operator
```

## RBAC
By default the chart will not install the recommended RBAC roles and rolebindings.

You need to have the flag `--authorization-mode=RBAC` on the api operator. See the following document for how to enable [RBAC](https://kubernetes.io/docs/admin/authorization/rbac/).

To determine if your cluster supports RBAC, run the following command:

```console
$ kubectl api-versions | grep rbac
```

If the output contains "beta", you may install the chart with RBAC enabled (see below).

### Enable RBAC role/rolebinding creation

To enable the creation of RBAC resources (On clusters with RBAC). Do the following:

```console
$ helm install --name my-release appscode/vault-operator --set rbac.create=true
```
