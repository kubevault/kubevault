---
title: Install Vault operator
menu:
  docs_{{ .version }}:
    identifier: install-operator
    name: Install
    parent: operator-setup
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: setup
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Installation Guide

Vault operator can be installed via a script or as a Helm chart.

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="helm3-tab" data-toggle="tab" href="#helm3" role="tab" aria-controls="helm3" aria-selected="true">Helm 3 (Recommended)</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="false">YAML</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="helm3" role="tabpanel" aria-labelledby="helm3-tab">

## Using Helm 3

Vault operator can be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/kubevault/installer/tree/{{< param "info.version" >}}/charts/vault-operator) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `my-release`:

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search repo appscode/vault-operator --version {{< param "info.version" >}}
NAME                    CHART VERSION APP VERSION   DESCRIPTION
appscode/vault-operator {{< param "info.version" >}}         {{< param "info.version" >}}         Vault Operator by AppsCode - HashiCorp Vault Operator for Kubernetes
appscode/vault-catalog  {{< param "info.version" >}}         {{< param "info.version" >}}         Vault Catalog by AppsCode - Catalog for vault versions

# Step 1: Install vault-operator chart
$ helm install vault-operator appscode/vault-operator --version {{< param "info.version" >}} --namespace kube-system

# Step 2: wait until crds are registered
$ kubectl get crds -l app.kubernetes.io/name=kubevault -w
NAME                                        AGE
vaultservers.kubevault.com                  12s
vaultserverversions.catalog.kubevault.com    8s

# Step 3: Install/Upgrade catalog of Vault versions

# Step 3(a): Install catalog of Vault versions
$ helm install vault-catalog appscode/vault-catalog --version {{< param "info.version" >}} --namespace kube-system

# Step 3(b): Or, if previously installed, upgrade catalog of Vault versions
$ helm upgrade vault-catalog appscode/vault-catalog --version {{< param "info.version" >}} --namespace kube-system
```

To see the detailed configuration options, visit [here](https://github.com/kubevault/installer/tree/{{< param "info.version" >}}/charts/vault-operator).

</div>
<div class="tab-pane fade" id="script" role="tabpanel" aria-labelledby="script-tab">

## Using YAML

If you prefer to not use Helm, you can generate YAMLs from Vault operator chart and deploy using `kubectl`. Here we are going to show the prodecure using Helm 3.

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search repo appscode/vault-operator --version {{< param "info.version" >}}
NAME                    CHART VERSION APP VERSION   DESCRIPTION
appscode/vault-operator {{< param "info.version" >}}         {{< param "info.version" >}}         Vault Operator by AppsCode - HashiCorp Vault Operator for Kubernetes
appscode/vault-catalog  {{< param "info.version" >}}         {{< param "info.version" >}}         Vault Catalog by AppsCode - Catalog for vault versions

# Step 1: Install vault-operator chart
$ helm template vault-operator appscode/vault-operator \
  --version {{< param "info.version" >}} \
  --namespace kube-system \
  --no-hooks | kubectl apply -f -

# Step 2: wait until crds are registered
$ kubectl get crds -l app.kubernetes.io/name=kubevault -w
NAME                                        AGE
vaultservers.kubevault.com                  12s
vaultserverversions.catalog.kubevault.com    8s

# Step 3: Install/Upgrade catalog of Vault versions
$ helm template vault-catalog appscode/vault-catalog \
  --version {{< param "info.version" >}} \
  --namespace kube-system | kubectl apply -f -
```

To see the detailed configuration options, visit [here](https://github.com/kubevault/installer/tree/{{< param "info.version" >}}/charts/vault-operator).

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
$ kubectl get pods --all-namespaces -l app.kubernetes.io/name=vault-operator --watch

NAMESPACE     NAME                              READY   STATUS    RESTARTS   AGE
kube-system   vault-operator-746d568685-m2w65   1/1     Running   0          5m44s
```

Once the operator pods are running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm CRD groups have been registered by the operator, run the following command:

```console
$ kubectl get crd -l app.kubernetes.io/name=kubevault
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

## Using kubectl for SecretEngine

```console
# List all SecretEngine objects
$ kubectl get secretengine --all-namespaces

# List SecretEngine objects for a namespace
$ kubectl get secretengine -n <namespace>

# Get SecretEngine YAML
$ kubectl get secretengine -n <namespace> <name> -o yaml

# Describe SecretEngine. Very useful to debug problems.
$ kubectl describe secretengine -n <namespace> <name>
```

## Using kubectl for GCPRole

```console
# List all GCPRole objects
$ kubectl get gcprole --all-namespaces

# List GCPRole objects for a namespace
$ kubectl get gcprole -n <namespace>

# Get GCPRole YAML
$ kubectl get gcprole -n <namespace> <name> -o yaml

# Describe GCPRole. Very useful to debug problems.
$ kubectl describe gcprole -n <namespace> <name>
```

## Using kubectl for GCPAccessKeyRequest

```console
# List all GCPAccessKeyRequest objects
$ kubectl get gcpaccesskeyrequest --all-namespaces

# List GCPAccessKeyRequest objects for a namespace
$ kubectl get gcpaccesskeyrequest -n <namespace>

# Get GCPAccessKeyRequest YAML
$ kubectl get gcpaccesskeyrequest -n <namespace> <name> -o yaml

# Describe GCPAccessKeyRequest. Very useful to debug problems.
$ kubectl describe gcpaccesskeyrequest -n <namespace> <name>
```

## Using kubectl for AzureRole

```console
# List all AzureRole objects
$ kubectl get azurerole --all-namespaces

# List AzureRole objects for a namespace
$ kubectl get azurerole -n <namespace>

# Get AzureRole YAML
$ kubectl get azurerole -n <namespace> <name> -o yaml

# Describe AzureRole. Very useful to debug problems.
$ kubectl describe azurerole -n <namespace> <name>
```

## Using kubectl for AzureAccessKeyRequest

```console
# List all AzureAccessKeyRequest objects
$ kubectl get azureaccesskeyrequest --all-namespaces

# List AzureAccessKeyRequest objects for a namespace
$ kubectl get azureaccesskeyrequest -n <namespace>

# Get AzureAccessKeyRequest YAML
$ kubectl get azureaccesskeyrequest -n <namespace> <name> -o yaml

# Describe AzureAccessKeyRequest. Very useful to debug problems.
$ kubectl describe azureaccesskeyrequest -n <namespace> <name>
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
To detect Vault operator version, exec into the operator pod and run `vault-operator version` command.

```console
$ POD_NAMESPACE=kube-system
$ POD_NAME=$(kubectl get pods -n $POD_NAMESPACE -l app.kubernetes.io/name=vault-operator -o jsonpath={.items[0].metadata.name})
$ kubectl exec -it $POD_NAME -n $POD_NAMESPACE vault-operator version

Version = {{< param "info.version" >}}
VersionStrategy = tag
Os = alpine
Arch = amd64
CommitHash = 85b0f16ab1b915633e968aac0ee23f877808ef49
GitBranch = release-0.5
GitTag = {{< param "info.version" >}}
CommitTimestamp = 2017-10-10T05:24:23
```
