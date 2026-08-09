---
title: Aerospike | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: aerospike-storage
    name: Aerospike
    parent: storage-vault-server-crds
    weight: 85
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Aerospike

In the Aerospike storage backend, Vault data will be stored in [Aerospike](https://aerospike.com/). This backend was removed from upstream OpenBao and is only available in this fork; Vault documentation for the Aerospike storage backend can be found [here](https://developer.hashicorp.com/vault/docs/configuration/storage/aerospike).

> `spec.backend.aerospike` is only available on the `kubevault.com/v1alpha2` `VaultServer` API. Aerospike has no `haEnabled` field and cannot be used to run Vault in high availability mode.

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault-with-aerospike
  namespace: demo
spec:
  version: "2.0.0"
  replicas: 1
  backend:
    aerospike:
      hostList: "aerospike-0.aerospike.demo.svc:3000,aerospike-1.aerospike.demo.svc:3000"
      namespace: "vault"
      credentialSecretRef:
        name: aerospike-cred
```

## spec.backend.aerospike

To use Aerospike as backend storage in Vault, specify `spec.backend.aerospike` in the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    aerospike:
      hostname: <hostname>
      port: <port>
      hostList: <comma_separated_host:port_list>
      namespace: <aerospike_namespace>
      set: <aerospike_set>
      credentialSecretRef:
        name: <secret_name>
      authMode: <INTERNAL|EXTERNAL>
      clusterName: <cluster_name>
      timeout: <milliseconds>
      idleTimeout: <milliseconds>
```

Here, we are going to describe the various attributes of the `spec.backend.aerospike` field.

### aerospike.hostname

`aerospike.hostname` is an optional field that specifies the hostname of the Aerospike server to connect to, when `hostList` is not set. Defaults to `127.0.0.1`.

```yaml
spec:
  backend:
    aerospike:
      hostname: "aerospike.demo.svc"
```

### aerospike.port

`aerospike.port` is an optional field that specifies the port of the Aerospike server to connect to, when `hostList` is not set. Defaults to `3000`.

```yaml
spec:
  backend:
    aerospike:
      port: "3000"
```

### aerospike.hostList

`aerospike.hostList` is an optional field that specifies a comma-separated list of `host[:port]` entries describing the Aerospike cluster's seed nodes. Takes precedence over `hostname`/`port` when set.

```yaml
spec:
  backend:
    aerospike:
      hostList: "aerospike-0.aerospike.demo.svc:3000,aerospike-1.aerospike.demo.svc:3000"
```

### aerospike.namespace

`aerospike.namespace` is an optional field that specifies the Aerospike namespace to store data in. The namespace must already exist on the server; Vault does not create it. Defaults to `test`.

```yaml
spec:
  backend:
    aerospike:
      namespace: "vault"
```

### aerospike.set

`aerospike.set` is an optional field that specifies the Aerospike set to store data in, within `namespace`.

```yaml
spec:
  backend:
    aerospike:
      set: "vault_kv_store"
```

### aerospike.credentialSecretRef

`aerospike.credentialSecretRef` is an optional field that specifies the name of the `Secret`, in the same namespace as the `VaultServer`, containing the username and password to authenticate to the Aerospike cluster with, if it has security enabled. The secret contains the following keys:

- `username`
- `password`

```yaml
spec:
  backend:
    aerospike:
      credentialSecretRef:
        name: aerospike-cred
```

### aerospike.authMode

`aerospike.authMode` is an optional field that specifies the authentication mode, either `INTERNAL` (Aerospike's own credential store) or `EXTERNAL` (e.g. LDAP). Defaults to `INTERNAL`.

```yaml
spec:
  backend:
    aerospike:
      authMode: "INTERNAL"
```

### aerospike.clusterName

`aerospike.clusterName` is an optional field that, if set, causes the client to verify it is connected to a cluster with this name.

```yaml
spec:
  backend:
    aerospike:
      clusterName: "vault-cluster"
```

### aerospike.timeout

`aerospike.timeout` is an optional field that specifies the socket connection timeout, in milliseconds.

```yaml
spec:
  backend:
    aerospike:
      timeout: 1000
```

### aerospike.idleTimeout

`aerospike.idleTimeout` is an optional field that specifies the idle connection timeout, in milliseconds. `0` disables idle connection trimming.

```yaml
spec:
  backend:
    aerospike:
      idleTimeout: 60000
```
