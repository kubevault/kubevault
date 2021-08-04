---
title: Vault Server | KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: vaultserver-vault-server-crds
    name: Vault Server
    parent: vault-server-crds-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# VaultServer

## What is VaultServer

A `VaultServer` is a Kubernetes `CustomResourceDefinition` (CRD) which is used to deploy a HashiCorp Vault server on Kubernetes clusters in a Kubernetes native way.

When a `VaultServer` is created, the KubeVault operator will deploy a Vault server and create necessary Kubernetes resources required for the Vault server.

![VaultServer CRD](/docs/images/concepts/vault_server.svg)

## VaultServer CRD Specification

Like any official Kubernetes resource, a `VaultServer` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample `VaultServer` object is shown below:

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: example
  namespace: default
spec:
  replicas: 1
  version: "1.7.2"
  backend:
    inmem: {}
  tls:
    certificates:
      - alias: ca
      - alias: server
      - alias: client
      - alias: storage
  serviceTemplates:
    - alias: vault
      metadata:
        annotations:
          name: vault
      spec:
        type: NodePort
    - alias: stats
      spec:
        type: ClusterIP
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      kubernetesSecret:
        secretName: vault-keys
  monitor:
    agent: prometheus.io
    prometheus:
      exporter:
        resources: {}
  terminationPolicy: WipeOut
```

Here, we are going to describe the various sections of the `VaultServer` crd.

### VaultServer Spec

VaultServer Spec contains the configuration about how to deploy Vault in the Kubernetes cluster. It also covers automate unsealing of Vault.

The `spec` section has following parts:

#### spec.replicas

`spec.replicas` specifies the number of Vault nodes to deploy. It has to be a positive number.

```yaml
spec:
  replicas: 3 # 3 vault server will be deployed in Kubernetes cluster
```

#### spec.version

Specifies the name of the `VaultServerVersion` CRD. This CRD holds the image name and version of the Vault, Unsealer, and Exporter. To know more information about `VaultServerVersion` CRD see [here](/docs/concepts/vault-server-crds/vaultserverversion.md).

```yaml
spec:
  version: "1.7.2"
```

#### spec.tls

`spec.tls` is an optional field that specifies the TLS policy of Vault nodes. If this is not specified, the KubeVault operator will run in `insecure` mode. `spec.tls.certificates` provides `server`, `client`, `storage` certificate options used by application pods, where the `alias` represents the identifier of the certificate.  

```yaml
spec:
  tls:
    certificates:
      - alias: ca
      - alias: server
      - alias: client
      - alias: storage
```

The server certificate must allow the following wildcard domains:
- `localhost`
- `*.<namespace>.pod`
- `<vaultServer-name>.<namespace>.svc`

  The server certificate must allow the following IP:
- `127.0.0.1`

#### spec.configSecret

`spec.configSecret` is an optional field that allows the user to provide extra configuration for Vault. This field accepts a [VolumeSource](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). You can use any Kubernetes supported volume source such as configMap, secret, azureDisk, etc.

> Please note that the config file name must be `vault.hcl` to work.

```yaml
spec:
  configSecret:
    <type of volume>: # for example `configSecret`
      name: <name of volume>
```

### spec.dataSources

`spec.dataSources` is an `optional` field that allows the user to provide a list of [VolumeSources](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes) (i.e. secrets, configmaps, etc.) which will be mounted into the VaultServer pods. These volumes will be mounted into `/etc/vault/data/<name>` directory. The first data will be named as `data-0`, the second one will be named as `data-1` and so on.

```yaml
spec:
  dataSources:
  - secret:  # mounted on /etc/vault/data/data-0
      secretName: custom-cert
  - configMap: # mounted on /etc/vault/data/data-1
      name: special-config
```

### spec.backend

`spec.backend` is a required field that specifies the Vault backend storage configuration. KubeVault operator generates storage configuration according to this `spec.backend`.

```yaml
spec:
  backend:
    ...
```

### spec.monitor
`spec.monitor` is an optional field that is used to monitor the `vaultserver` instances.
```yaml
monitor:
    agent: prometheus.io
    prometheus:
      exporter:
        resources: {}
```

###spec.terminationPolicy
`spec.terminationPolicy` is an optional field that gives flexibility whether to nullify(reject) the delete operation of VaultServer crd or which resources KubeVault operator should keep or delete when you delete VaultServer crd. KubeVault provides following four termination policies:
- DoNotTerminate
- Halt
- Delete (Default)
- WipeOut

When, `terminationPolicy` is `DoNotTerminate`, KubeVault takes advantage of ValidationWebhook feature in Kubernetes 1.9.0 or later clusters to provide safety from accidental deletion of VaultServer. If admission webhook is enabled, KubeVault prevents users from deleting the VaultServer as long as the spec.terminationPolicy is set to DoNotTerminate.


List of supported backends:

- [Azure](/docs/concepts/vault-server-crds/storage/azure.md)
- [Consul](/docs/concepts/vault-server-crds/storage/consul.md)
- [DynamoDB](/docs/concepts/vault-server-crds/storage/dynamodb.md)
- [Etcd](/docs/concepts/vault-server-crds/storage/etcd.md)
- [GCS](/docs/concepts/vault-server-crds/storage/gcs.md)
- [In Memory](/docs/concepts/vault-server-crds/storage/inmem.md)
- [MySQL](/docs/concepts/vault-server-crds/storage/mysql.md)
- [PosgreSQL](/docs/concepts/vault-server-crds/storage/postgresql.md)
- [AWS S3](/docs/concepts/vault-server-crds/storage/s3.md)
- [Swift](/docs/concepts/vault-server-crds/storage/swift.md)
- [Filesystem](/docs/concepts/vault-server-crds/storage/filesystem.md)

#### spec.unsealer

`spec.unsealer` is an optional field that specifies [Unsealer](https://github.com/kubevault/unsealer) configuration. Unsealer handles automatic initializing and unsealing of Vault. See [here](/docs/concepts/vault-server-crds/unsealer/unsealer.md) for Unsealer documentation.

```yaml
spec:
  unsealer:
    secretShares: <num_of_secret_shares>
    secretThresold: <num_of_secret_threshold>
    retryPeriodSeconds: <retry_period>
    overwriteExisting: <true/false>
    mode:
      ...
```

#### spec.serviceTemplates

You can also provide a list of templates for the services created by KubeVault operator for VaultServer through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services. `spec.serviceTemplates` is an optional field.

```yaml
spec:
  serviceTemplates:
    - alias: stats
      spec:
        type: ClusterIP
```

VaultServer allows following fields to be set in `spec.serviceTemplates`:

- metadata:
  - annotations (set as annotations on Vault service)
- spec:
  - type
  - ports
  - clusterIP
  - externalIPs
  - loadBalancerIP
  - loadBalancerSourceRanges
  - externalTrafficPolicy
  - healthCheckNodePort
  - sessionAffinityConfig

#### spec.podTemplate

VaultServer allows providing a template for Vault pod through `spec.podTemplate`. KubeVault operator will pass the information provided in `spec.podTemplate` to the Deployment created for Vault. `spec.podTemplate` is an optional field.

```yaml
spec:
  podTemplate:
    spec:
      resources:
        requests:
          memory: "64Mi"
          cpu: "250m"
        limits:
          memory: "128Mi"
          cpu: "500m"
```

VaultServer accepts the following fields to set in `spec.podTemplate:`

- metadata:
  - annotations (set as annotations on Vault pods)
- controller:
  - annotations (set as annotations on Vault statefulset)
- spec:
  - resources
  - imagePullSecrets
  - nodeSelector
  - affinity
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext

You can find the full list of fields [here](https://github.com/kmodules/offshoot-api/blob/kubernetes-1.16.3/api/v1/types.go). Some of the fields of `spec.podTemplate` are described below:

##### spec.podTemplate.spec.imagePullSecret

`spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker images if you are using a private docker registry.

##### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

##### spec.podTemplate.spec.resources

`spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by Vault pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

#### spec.authMethods

`spec.authMethods` is an optional field that specifies the list of auth methods to enable in Vault.

```yaml
spec:
  authMethods:
    - type: kubernetes
      path: k8s
    - type: aws
      path: aws
```

`spec.authMethods` has following fields:

##### spec.authMethods[].type

`spec.authMethods[].type` is a required field that specifies the name of the authentication method type.

##### spec.authMethods[].path

`spec.authMethods[].path` is a required field that specifies the path where to enable the auth method.

##### spec.authMethods[].description

`spec.authMethods[].description` is an optional field that specifies a human-friendly description of the auth method.

##### spec.authMethods[].pluginName

`spec.authMethods[].pluginName` is an optional field that specifies the name of the auth plugin to use based on the name in the plugin catalog.

##### spec.authMethods[].local

`spec.authMethods[].local` is an optional field that specifies if the auth method is local only. Local auth methods are not replicated nor (if a secondary) removed by replication.

##### spec.authMethods[].config

`spec.authMethods[].config` is an optional field that specifies configuration options for auth method.

`spec.authMethods[].config` has following fields:

- `defaultLeaseTTL` : `Optional`. Specifies the default lease duration.

- `maxLeaseTTL` : `Optional`. Specifies the maximum lease duration.

- `pluginName` : `Optional`. Specifies the name of the plugin in the plugin catalog to use.

- `auditNonHMACRequestKeys` : `Optional`. Specifies the list of keys that will not be HMAC'd by audit devices in the request data object.

- `auditNonHMACResponseKeys`: `Optional`. Specifies the list of keys that will not be HMAC'd by audit devices in the response data object.

- `listingVisibility`: `Optional`.  Specifies whether to show this is mount in the UI-specific listing endpoint.

- `passthroughRequestHeaders`: `Optional`. Specifies a list of headers to whitelist and pass from the request to the backend.

### VaultServer Status

VaultServer Status shows the status of a Vault deployment. The status of the Vault is monitored and updated by the KubeVault operator.

```yaml
status:
  phase: <phase>
  initialized: <true/false>
  serviceName: <service_name>
  clientPort: <client_port>
  vaultStatus:
    active: <active_vault_pod_name>
    standby: <names_of_the_standby_vault_pod>
    sealed: <names_of_the_sealed_vault_pod>
    unsealed: <names_of_the_unsealed_vault_pod>
```

- `phase`: Indicates the phase Vault is currently in.

- `initialized`: Indicates whether Vault is initialized or not.

- `serviceName`: Name of the service by which Vault can be accessed.

- `clientPort`: Indicates the port client will use to communicate with Vault.

- `vaultStatus`: Indicates the status of Vault pods. It has the following fields:

  - `active`: Name of the active vault pod.

  - `standby`: Names of the standby vault pods.

  - `sealed`: Names of the sealed vault pods.

  - `unsealed`: Names of the unsealed vault pods.

- `authMethodStatus` : Indicates the status of the auth methods specified in `spec.authMethods`. It has the following fields:

  - `type`: Specifies the name of the authentication method type

  - `path`: Specifies the path in which the auth method is enabled.

  - `status`: Specifies whether the auth method is enabled or not.

  - `reason`: Specifies the reason why failed to enable the auth method.
