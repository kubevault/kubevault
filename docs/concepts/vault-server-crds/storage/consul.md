---
title: Consul | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: consul-storage
    name: Consul
    parent: storage-vault-server-crds
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Consul

To use Consul as storage backend, first we need to deploy consul. Documentations can be found [here](https://www.consul.io/docs/platform/k8s/run.html)

Here is an example:

Pod yaml:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: consul-example
  namespace: demo
  labels:
    app: consul
spec:
  containers:
    - name: example
      image: "consul:latest"
  restartPolicy: Never
```

Service yaml:

```yaml
kind: Service
apiVersion: v1
metadata:
  name: my-service
  namespace: demo
spec:
  selector:
    app: consul
  ports:
  - protocol: TCP
    port: 8500
  type: NodePort
```

## Consul Backend

In Consul storage backend, data will be stored in consul storage container. Vault documentation for Consul storage backend can be found in [here](https://www.vaultproject.io/docs/configuration/storage/consul.html)

[VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) yaml:

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  nodes: 1
  version: "1.0.1"
  serviceTemplate:
    spec:
      type: NodePort
  backend:
    consul:
      address: "http://my-service.demo.svc:8500"
      path: "vault"
  configSource:
    configMap:
      name: extra-config
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      kubernetesSecret:
        secretName: vault-keys
```

If you need to disable the server from executing the `mlock` syscall, you can provide [disable_mlock](https://www.vaultproject.io/docs/configuration/#disable_mlock) in a ConfigMap and mention the name in `spec.configSource.configMap.name`.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: extra-config
  namespace: demo
data:
  vault.hcl: |
    disable_mlock = true
```

## spec.backend.consul

To use Consul as backend storage in Vault, we need to specify `spec.backend.consul` in [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.
More information about the Consul backend storage can be found in [here](https://www.vaultproject.io/docs/configuration/storage/consul.html)

```yaml
spec:
  backend:
    consul:
      address: <address_of_consul_agent>
      checkTimeout: <check_interval>
      consistencyMode: <consul_consistency_mode>
      disableRegistration: <disable_registration>
      maxParallel: <max_number_of_concurrent_request>
      path: <key-value_store_path>
      scheme: <scheme_type>
      service: <service_name>
      serviceTags: <list_of_tags>
      serviceAddress: <service_specific_address>
      aclTokenSecretName: <aclToken_secret_name>
      sessionTTL: <minimum_allowed_session_TTL>
      lockWaitTime: <time_before_lock_acquisition>
      tlsSecretName: <secret_name>
      tlsMinVersion: <minimum_TLS_version>
      tlsSkipVerify: <boolean_value>
```

`spec.backend.consul` has the following fields:

### consul.address

Specifies the address of the Consul agent to communicate with. This can be an IP address, DNS record, or unix socket. It is recommended that you communicate with a local Consul agent; do not communicate directly with a server.

```yaml
spec:
  backend:
    consul:
      address: "127.0.0.1:8500"
```

### consul.checkTimeout

Specifies the check interval used to send health check information back to Consul. This is specified using a label suffix like "30s" or "1h"

```yaml
spec:
  backend:
    consul:
      checkTimeout: "5s"
```

### consul.consistencyMode

Specifies the Consul consistency mode. Possible values are "default" or "strong".

```yaml
spec:
  backend:
    consul:
      consistencyMode: "default"
```

### consul.disableRegistration

Specifies whether Vault should register itself with Consul.

```yaml
spec:
  backend:
    consul:
      disableRegistration: "false"
```

### consul.maxParallel

Specifies the maximum number of concurrent requests to Consul.

```yaml
spec:
  backend:
    consul:
      maxParallel: "128"
```

### consul.path

Specifies the path in Consul's key-value store where Vault data will be stored.

```yaml
spec:
  backend:
    consul:
      path: "vault"
```

### consul.scheme

Specifies the scheme to use when communicating with Consul. This can be set to "http" or "https". It is highly recommended you communicate with Consul over https over non-local connections. When communicating over a unix socket, this option is ignored.

```yaml
spec:
  backend:
    consul:
      scheme: "http"
```

### consul.service

Specifies the name of the service to register in Consul.

```yaml
spec:
  backend:
    consul:
      service: "vault"
```

### consul.serviceTags

Specifies a comma-separated list of tags to attach to the service registration in Consul.

```yaml
spec:
  backend:
    consul:
      serviceTags: ""
```

### consul.serviceAddress

Specifies a service-specific address to set on the service registration in Consul. If unset, Vault will use what it knows to be the HA redirect address - which is usually desirable. Setting this parameter to "" will tell Consul to leverage the configuration of the node the service is registered on dynamically.

```yaml
spec:
  backend:
    consul:
      serviceAddress: ""
```

### consul.aclTokenSecretName

Specifies the secret name that contains ACL token with permission to read and write from the path in Consul's key-value store. ACL Token should be stored in `data["aclToken"]=<value>`

```yaml
spec:
  backend:
    consul:
      aclTokenSecretName: aclcred
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aclcred
  namespace: demo
data:
  aclToken: |-
   ZGF0YQ==
```

### consul.sessionTTL

Specifies the minimum allowed session TTL. Consul server has a lower limit of 10s on the session TTL by default. The value of session_ttl here cannot be lesser than 10s unless the session_ttl_min on the consul server's configuration has a lesser value.

```yaml
spec:
  backend:
    consul:
      sessionTTL: "15s"
```

### consul.lockWaitTime

Specifies the wait time before a lock lock acquisition is made. This affects the minimum time it takes to cancel a lock acquisition.

```yaml
spec:
  backend:
    consul:
      lockWaitTime: "15s"
```

### consul.tlsSecretName

Specifies the secret name that contains tls_ca_file, tls_cert_file and tls_key_file for consul communication.

```yaml
spec:
  backend:
    consul:
      tlsSecretName: tls
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: tls
  namespace: demo
data:
  ca.crt: <tls_ca_file>
  client.crt: <tls_cert_file>
  client.key: <tls_key_file>
```

### consul.tlsMinVersion

Specifies the minimum TLS version to use. Accepted values are "tls10", "tls11" or "tls12".

```yaml
spec:
  backend:
    consul:
      tlsMinVersion: "tls12"
```

### consul.tlsSkipVerify

Disable verification of TLS certificates. Using this option is highly discouraged. It is a `boolean` type variable.

```yaml
spec:
  backend:
    consul:
      tlsSkipVerify: false
```
