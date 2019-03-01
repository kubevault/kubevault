---
title: Vault Server | KubeVault Concepts
menu:
  docs_0.2.0:
    identifier: vaultserver-vault-server-crds
    name: Vault Server
    parent: vault-server-crds-concepts
    weight: 10
menu_name: docs_0.2.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# VaultServer CRD

Vault operator will deploy Vault according to `VaultServer` CRD (CustomResourceDefinition) specification.

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: <name>
spec:
  ...
status:
  ...
```

## VaultServer Spec

VaultServer Spec contains the configuration about how to deploy Vault in Kubernetes cluster. It also covers automate unsealing of Vault.

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: example
  namespace: default
spec:
  nodes: 1
  version: "0.11.1"
  backend:
    inmem: {}
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      kubernetesSecret:
        secretName: vault-keys
```

The `spec` section has following parts:

### spec.nodes

`spec.nodes` specifies the number of vault nodes to deploy. It has to be a positive number.

```yaml
spec:
  nodes: 3 # 3 vault server will be deployed in Kubernetes cluster
```

### spec.version

Specifies the name of the `VaultServerVersion` CRD. This CRD holds the image name and version of the Vault, Unsealer and Exporter. To know more information about `VaultServerVersion` CRD see [here](/docs/concepts/vault-server-crds/vaultserverversion.md).

```yaml
spec:
  version: "1.0.0"
```

### spec.tls

`spec.tls` is an optional field that specifies TLS policy of Vault nodes. If this is not specified, Vault operator will auto generate TLS assets and secrets.

```yaml
spec:
  tls:
    tlsSecret: <tls_assets_secret_name> # name of the secret containing TLS assets
    caBundle: <pem_coded_ca>
```

- **`tls.tlsSecret`**: Specifies the name of the secret containing TLS assets. The secret must contain following keys:
  - `tls.crt`
  - `tls.key`

  The server certificate must allow the following wildcard domains:
  - `localhost`
  - `*.<namespace>.pod`
  - `<vaultServer-name>.<namespace>.svc`

  The server certificate must allow the following ip:
  - `127.0.0.1`

- **`tls.caBundle`**: Specifies the PEM encoded CA bundle which will be used to validate the serving certificate.

### spec.configSource

`spec.configSource` is an optional field that allows the user to provide extra configuration for Vault. This field accepts a [VolumeSource](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). You can use any Kubernetes supported volume source such as configMap, secret, azureDisk etc.

> Please note that the config file name needs to be `vault.hcl` for Vault.

### spec.backend

`spec.backend` is a required field that specifies Vault backend storage configuration. Vault operator generates storage configuration according to this `spec.backend`.

```yaml
spec:
  backend:
    ...
```

List of supported backends:

- [Azure](/docs/concepts/vault-server-crds/storage/azure.md)
- [S3](/docs/concepts/vault-server-crds/storage/s3.md)
- [GCS](/docs/concepts/vault-server-crds/storage/gcs.md)
- [DynamoDB](/docs/concepts/vault-server-crds/storage/dynamodb.md)
- [PosgreSQL](/docs/concepts/vault-server-crds/storage/postgresql.md)
- [MySQL](/docs/concepts/vault-server-crds/storage/mysql.md)
- [In Memory](/docs/concepts/vault-server-crds/storage/inmem.md)
- [Swift](/docs/concepts/vault-server-crds/storage/swift.md)
- [Etcd](/docs/concepts/vault-server-crds/storage/etcd.md)

### spec.unsealer

`spec.unsealer` is an optional field that specifies [Unsealer](https://github.com/kubevault/unsealer) configuration. Unsealer handles automatic initializing and unsealing of Vault. See [here](/docs/concepts/vault-server-crds/unsealer/unsealer.md) for Unsealer fields information.

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

### spec.serviceTemplate

You can also provide a template for the services created by Vault operator for VaultServer through `spec.serviceTemplate`. This will allow you to set the type and other properties of the services. `spec.serviceTemplate` is an optional field.

```yaml
spec:
  serviceTemplate:
    spec:
      type: NodePort
```

VaultServer allows following fields to set in `spec.serviceTemplate`:

- metadata:
  - annotations
- spec:
  - type
  - ports
  - clusterIP
  - externalIPs
  - loadBalancerIP
  - loadBalancerSourceRanges
  - externalTrafficPolicy
  - healthCheckNodePort

### spec.podTemplate

VaultServer allows providing a template for Vault pod through `spec.podTemplate`. Vault operator will pass the information provided in `spec.podTemplate` to the Deployment created for Vault. `spec.podTemplate` is an optional field.

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

VaultServer accept following fields to set in `spec.podTemplate:`

- metadata:
  - annotations (pod's annotation)
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

Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.imagePullSecret

`spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker image if you are using a private docker registry.

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.resources

`spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by Vault pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.authMethods

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

#### spec.authMethods[].type

`spec.authMethods[].type` is a required field that specifies the name of the authentication method type.

#### spec.authMethods[].path

`spec.authMethods[].path` is a required field that specifies the path in which to enable the auth method.

#### spec.authMethods[].description

`spec.authMethods[].description` is an optional field that specifies a human-friendly description of the auth method.

#### spec.authMethods[].pluginName

`spec.authMethods[].pluginName` is an optional field that specifies the name of the auth plugin to use based from the name in the plugin catalog.

#### spec.authMethods[].local

`spec.authMethods[].local` is an optional field that specifies if the auth method is a local only. Local auth methods are not replicated nor (if a secondary) removed by replication.

#### spec.authMethods[].config

`spec.authMethods[].config` is an optional field that specifies configuration options for auth method.

`spec.authMethods[].config` has following fields:

- `defaultLeaseTTL` : `Optional`. Specifies the default lease duration.

- `maxLeaseTTL` : `Optional`. Specifies the maximum lease duration.

- `pluginName` : `Optional`. Specifies the name of the plugin in the plugin catalog to use.

- `auditNonHMACRequestKeys` : `Optional`. Specifies the list of keys that will not be HMAC'd by audit devices in the request data object.

- `auditNonHMACResponseKeys`: `Optional`. Specifies the list of keys that will not be HMAC'd by audit devices in the response data object.

- `listingVisibility`: `Optional`.  Specifies whether to show this mount in the UI-specific listing endpoint.

- `passthroughRequestHeaders`: `Optional`. Specifies list of headers to whitelist and pass from the request to the backend.

## VaultServer Status

VaultServer Status shows the status of Vault deployment. Status of vault is monitored and updated by Vault operator.

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

- `phase` : Indicates the phase Vault is currently in.

- `initialized` : Indicates whether vault is initialized or not.

- `serviceName` : Name of the service by which vault can be accessed.

- `clientPort` : Indicates the port client will use to communicate with vault.

- `vaultStatus` : Indicates the status of vault pods. It has following fields:

  - `active` : Name of the active vault pod.

  - `standby` : Names of the standby vault pods.

  - `sealed` : Names of the sealed vault pods.

  - `unsealed` : Names of the unsealed vault pods.

- `authMethodStatus` : Indicates the status of the auth methods specified in `spec.authMethods`. It has following fields:

  - `type` : Specifies the name of the authentication method type

  - `path` : Specifies the path in which the auth method is enabled.

  - `status` : Specifies whether auth method is enabled or not.

  - `reason` : Specifies the reason why failed to enable auth method.
