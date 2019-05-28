---
title: Install Vault operator
menu:
  docs_0.2.0:
    identifier: install-operator
    name: Install
    parent: operator-setup
    weight: 10
menu_name: docs_0.2.0
section_menu_id: setup
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Installation Guide

Vault operator can be installed via a script or as a Helm chart.

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="true">Script</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="helm-tab" data-toggle="tab" href="#helm" role="tab" aria-controls="helm" aria-selected="false">Helm</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="script" role="tabpanel" aria-labelledby="script-tab">

## Using Script

To install Vault operator in your Kubernetes cluster, run the following command:

```console
$ curl -fsSL https://github.com/kubevault/operator/raw/0.2.0/hack/deploy/install.sh | bash
```

After successful installation, you should have a `vault-operator-***` pod running in the `kube-system` namespace.

```console
$ kubectl get pods -n kube-system | grep vault-operator
vault-operator-846d47f489-jrb58       1/1       Running   0          48s
```

#### Customizing Installer

The installer script and associated yaml files can be found in the [/hack/deploy](https://github.com/kubevault/operator/tree/0.2.0/hack/deploy) folder. You can see the full list of flags available to installer using `-h` flag.

```console
$ curl -fsSL https://github.com/kubevault/operator/raw/0.2.0/hack/deploy/install.sh | bash -s -- -h
install.sh - install Vault operator

install.sh [options]

options:
-h, --help                             show brief help
-n, --namespace=NAMESPACE              specify namespace (default: kube-system)
    --docker-registry                  docker registry used to pull Vault images (default: kubevault)
    --image-pull-secret                name of secret used to pull Vault images
    --run-on-master                    run Vault operator on master
    --enable-mutating-webhook          enable/disable mutating webhooks for Kubernetes workloads
    --enable-validating-webhook        enable/disable validating webhooks for Vault CRDs
    --bypass-validating-webhook-xray   if true, bypasses validating webhook xray checks
    --enable-status-subresource        if enabled, uses status sub resource for crds
    --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
    --enable-analytics                 send usage events to Google Analytics (default: true)
    --uninstall                        uninstall Vault operator
    --purge                            purges Vault CRD objects and crds
    --install-catalog                  installs Vault server version catalog (default: all)
    --monitoring-agent                 specify which monitoring agent to use (default: none)
    --monitor-operator                 specify whether to monitor Vault operator (default: false)
    --prometheus-namespace             specify the namespace where Prometheus server is running or will be deployed (default: same namespace as vault-operator)
    --servicemonitor-label             specify the label for ServiceMonitor crd. Prometheus crd will use this label to select the ServiceMonitor. (default: 'app: vault-operator')
    --cluster-name                     Name of cluster used in a multi-cluster setup
```

If you would like to run Vault operator pod in `master` instances, pass the `--run-on-master` flag:

```console
$ curl -fsSL https://github.com/kubevault/operator/raw/0.2.0/hack/deploy/install.sh \
    | bash -s -- --run-on-master
```

Vault operator will be installed in a `kube-system` namespace by default. If you would like to run Vault operator pod in `vault` namespace, pass the `--namespace=vault` flag:

```console
$ kubectl create namespace vault
$ curl -fsSL https://github.com/kubevault/operator/raw/0.2.0/hack/deploy/install.sh \
    | bash -s -- --namespace=vault [--run-on-master]
```

If you are using a private Docker registry, you need to pull the following images:

 - [kubevault/vault-operator](https://hub.docker.com/r/kubevault/vault-operator)
 - [kubevault/vault-unsealer](https://hub.docker.com/r/kubevault/vault-unsealer)
 - [kubevault/csi-vault](https://hub.docker.com/r/kubevault/csi-vault)

To pass the address of your private registry and optionally a image pull secret use flags `--docker-registry` and `--image-pull-secret` respectively.

```console
$ kubectl create namespace vault
$ curl -fsSL https://github.com/kubevault/operator/raw/0.2.0/hack/deploy/install.sh \
    | bash -s -- --docker-registry=MY_REGISTRY [--image-pull-secret=SECRET_NAME]
```

Vault operator implements [validating admission webhooks](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#validatingadmissionwebhook) to validate KubeVault CRDs and **mutating webhooks** for KubeVault crds. This is enabled by default for Kubernetes 1.9.0 or later releases. To disable this feature, pass the `--enable-validating-webhook=false` and `--enable-mutating-webhook=false` flag respectively.

```console
$ curl -fsSL https://github.com/kubevault/operator/raw/0.2.0/hack/deploy/install.sh \
    | bash -s -- --enable-validating-webhook=false --enable-mutating-webhook=false
```

</div>
<div class="tab-pane fade" id="helm" role="tabpanel" aria-labelledby="helm-tab">

## Using Helm
Vault operator can be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/kubevault/operator/tree/0.2.0/chart/vault-operator) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `my-release`:

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search appscode/vault
NAME                    CHART VERSION APP VERSION   DESCRIPTION
appscode/vault-operator 0.2.0         0.2.0         Vault Operator by AppsCode - HashiCorp Vault Operator for Kubernetes
appscode/vault-catalog  0.2.0         0.2.0         Vault Catalog by AppsCode - Catalog for vault versions

# Step 1: Install vault-operator chart
$ helm install appscode/vault-operator --name vault-operator --version 0.2.0 \
  --namespace kube-system

# Step 2: wait until crds are registered
$ kubectl get crds -l app=vault -w
NAME                                        AGE
vaultservers.kubevault.com                  12s
vaultserverversions.catalog.kubevault.com    8s

# Step 3: Install catalog of Vault versions
$ helm install appscode/vault-catalog --name vault-catalog

# Step 3(a): Install catalog of Vault versions
$ helm install appscode/vault-catalog --name vault-catalog --version 0.2.0 \
  --namespace kube-system

# Step 3(b): Or, if previously installed, upgrade catalog of Vault versions
$ helm upgrade vault-catalog appscode/vault-catalog --version 0.2.0 \
  --namespace kube-system
```

To see the detailed configuration options, visit [here](https://github.com/kubevault/operator/tree/master/chart/vault-operator).

</div>
</div>

### Installing in GKE Cluster

If you are installing Vault operator on a GKE cluster, you will need cluster admin permissions to install Vault operator. Run the following command to grant admin permision to the cluster.

```console
$ kubectl create clusterrolebinding "cluster-admin-$(whoami)" \
  --clusterrole=cluster-admin \
  --user="$(gcloud config get-value core/account)"
```


## Verify installation
To check if Vault operator pods have started, run the following command:
```console
$ kubectl get pods --all-namespaces -l app=vault-operator --watch

NAMESPACE     NAME                              READY   STATUS    RESTARTS   AGE
kube-system   vault-operator-746d568685-m2w65   1/1     Running   0          5m44s
```

Once the operator pods are running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm CRD groups have been registered by the operator, run the following command:
```console
$ kubectl get crd -l app=vault

NAME                                        CREATED AT
awsaccesskeyrequests.engine.kubevault.com   2019-01-08T05:57:21Z
awsroles.engine.kubevault.com               2019-01-08T05:57:21Z
vaultpolicies.policy.kubevault.com          2019-01-08T05:57:21Z
vaultpolicybindings.policy.kubevault.com    2019-01-08T05:57:21Z
vaultservers.kubevault.com                  2019-01-08T05:57:17Z
vaultserverversions.catalog.kubevault.com   2019-01-08T05:57:21Z

```

Now, you are ready to [deploy and manage Vault](/docs/guides/README.md) using Vault operator.


## Configuring RBAC
Vault operator creates multiple CRDs. Vault operator installer will create 2 user facing cluster roles:

| ClusterRole          | Aggregates To | Desription                            |
|----------------------|---------------|---------------------------------------|
| kubevault:core:admin | admin         | Allows admin access to Vault operator CRDs, intended to be granted within a namespace using a RoleBinding. |
| kubevault:core:edit  | admin, edit   | Allows edit access to Vault operator CRDs, intended to be granted within a namespace using a RoleBinding. |
| kubevault:core:view  | view          | Allows read-only access to Vault operator CRDs, intended to be granted within a namespace using a RoleBinding. |

These user facing roles supports [ClusterRole Aggregation](https://kubernetes.io/docs/admin/authorization/rbac/#aggregated-clusterroles) feature in Kubernetes 1.9 or later clusters.


## Using kubectl for VaultServer
```console
# List all VaultServer objects
$ kubectl get vaultserver --all-namespaces

# List VaultServer objects for a namespace
$ kubectl get vaultserver -n <namespace>

# Get VaultServer YAML
$ kubectl get vaultserver -n <namespace> <name> -o yaml

# Describe VaultServer. Very useful to debug problems.
$ kubectl describe vaultserver -n <namespace> <name>
```

## Using kubectl for VaultPolicy
```console
# List all VaultPolicy objects
$ kubectl get vaultpolicy --all-namespaces

# List VaultPolicy objects for a namespace
$ kubectl get vaultpolicy -n <namespace>

# Get VaultPolicy YAML
$ kubectl get vaultpolicy -n <namespace> <name> -o yaml

# Describe VaultPolicy. Very useful to debug problems.
$ kubectl describe vaultpolicy -n <namespace> <name>
```

## Using kubectl for VaultPolicyBinding
```console
# List all VaultPolicyBinding objects
$ kubectl get vaultpolicybinding --all-namespaces

# List VaultPolicyBinding objects for a namespace
$ kubectl get vaultpolicybinding -n <namespace>

# Get VaultPolicyBinding YAML
$ kubectl get vaultpolicybinding -n <namespace> <name> -o yaml

# Describe VaultPolicyBinding. Very useful to debug problems.
$ kubectl describe vaultpolicybinding -n <namespace> <name>
```

## Using kubectl for AWSRole
```console
# List all AWSRole objects
$ kubectl get awsrole --all-namespaces

# List AWSRole objects for a namespace
$ kubectl get awsrole -n <namespace>

# Get AWSRole YAML
$ kubectl get awsrole -n <namespace> <name> -o yaml

# Describe AWSRole. Very useful to debug problems.
$ kubectl describe awsrole -n <namespace> <name>
```

## Using kubectl for AWSAccessKeyRequest
```console
# List all AWSAccessKeyRequest objects
$ kubectl get awsaccesskeyrequest --all-namespaces

# List AWSAccessKeyRequest objects for a namespace
$ kubectl get awsaccesskeyrequest -n <namespace>

# Get AWSAccessKeyRequest YAML
$ kubectl get awsaccesskeyrequest -n <namespace> <name> -o yaml

# Describe AWSAccessKeyRequest. Very useful to debug problems.
$ kubectl describe awsaccesskeyrequest -n <namespace> <name>
```

## Using kubectl for DatabaseAccessRequest
```console
# List all DatabaseAccessRequest objects
$ kubectl get databaseaccessrequest --all-namespaces

# List DatabaseAccessRequest objects for a namespace
$ kubectl get databaseaccessrequest -n <namespace>

# Get DatabaseAccessRequest YAML
$ kubectl get databaseaccessrequest -n <namespace> <name> -o yaml

# Describe DatabaseAccessRequest. Very useful to debug problems.
$ kubectl describe databaseaccessrequest -n <namespace> <name>
```

## Using kubectl for PostgresRole
```console
# List all PostgresRole objects
$ kubectl get postgresrole --all-namespaces

# List PostgresRole objects for a namespace
$ kubectl get postgresrole -n <namespace>

# Get PostgresRole YAML
$ kubectl get postgresrole -n <namespace> <name> -o yaml

# Describe PostgresRole. Very useful to debug problems.
$ kubectl describe postgresrole -n <namespace> <name>
```

## Using kubectl for MySQLRole
```console
# List all MySQLRole objects
$ kubectl get mysqlrole --all-namespaces

# List MySQLRole objects for a namespace
$ kubectl get mysqlrole -n <namespace>

# Get MySQLRole YAML
$ kubectl get mysqlrole -n <namespace> <name> -o yaml

# Describe MySQLRole. Very useful to debug problems.
$ kubectl describe mysqlrole -n <namespace> <name>
```

## Using kubectl for MongoDBRole
```console
# List all MongoDBRole objects
$ kubectl get mongodbrole --all-namespaces

# List MongoDBRole objects for a namespace
$ kubectl get mongodbrole -n <namespace>

# Get MongoDBRole YAML
$ kubectl get mongodbrole -n <namespace> <name> -o yaml

# Describe MongoDBRole. Very useful to debug problems.
$ kubectl describe mongodbrole -n <namespace> <name>
```

## Detect Vault operator version
To detect Vault operator version, exec into the operator pod and run `vault version` command.

```console
$ POD_NAMESPACE=kube-system
$ POD_NAME=$(kubectl get pods -n $POD_NAMESPACE -l app=vault-operator -o jsonpath={.items[0].metadata.name})
$ kubectl exec -it $POD_NAME -n $POD_NAMESPACE vault-operator version

Version = 0.2.0
VersionStrategy = tag
Os = alpine
Arch = amd64
CommitHash = 85b0f16ab1b915633e968aac0ee23f877808ef49
GitBranch = release-0.5
GitTag = 0.2.0
CommitTimestamp = 2017-10-10T05:24:23

```
