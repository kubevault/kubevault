---
title: etcd | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: etcd-storage
    name: etcd
    parent: storage-vault-server-crds
    weight: 100
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# etcd

In etcd storage backend, Vault data will be stored in [etcd](https://etcd.io/). Vault documentation for etcd storage can be found in [here](https://www.vaultproject.io/docs/configuration/storage/etcd.html).

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault-with-etcd
  namespace: demo
spec:
  replicas: 1
  version: "1.18.4"
  backend:
    etcd:
      address: "http://example.etcd.svc:2379"
      etcdApi: "v3"
```

## spec.backend.etcd

To use etcd as storage backend in Vault specify `spec.backend.etcd` in [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    etcd:
      address: "http://example.etcd.svc:2379"
      etcdApi: "v3"
```

Here, we are going to describe the various attributes of the `spec.backend.etcd` field.

### etcd.address

`etcd.address` is a required field that specifies the addresses of the etcd instances.

```yaml
spec:
  backend:
    etcd:
      address: "http://example.etcd.svc:2379"
```

### etcd.haEnable

`etcd.haEnable` is an optional field that specifies if high availability should be enabled. This field accepts boolean value. The default value is `false`.

```yaml
spec:
  backend:
    etcd:
      haEnable: true
```

> On the `kubevault.com/v1alpha2` `VaultServer` API, this field is named `haEnabled`, the same as on the other storage backends.

```yaml
spec:
  backend:
    etcd:
      haEnabled: true
```

### etcd.etcdApi

`etcd.etcdApi` is an optional field that specifies the version of the API to communicate with etcd. If this field is not specified, then Vault will derive it automatically. If the cluster version is 3.1+ and there has been no data written using the v2 API, the auto-detected default is v3.

```yaml
spec:
  backend:
    etcd:
      etcdApi: "v3"
```

### etcd.path

`etcd.path` is an optional field that specifies the path in etcd where Vault data will be stored. If this field is not specified, then Vault will set default value `/vault/`.

```yaml
spec:
  backend:
    etcd:
      path: "/data/"
```

### etcd.sync

`etcd.sync` is an optional field that specifies whether to sync a list of available etcd services on startup. This field accepts boolean value. The default value is `false`.

```yaml
spec:
  backend:
    etcd:
      sync: true
```

### etcd.discoverySrv

`etcd.discoverySrv` is an optional field that specifies the domain name to query for SRV records describing cluster endpoints. If this field is not specified, then Vault will set default value `example.com`

```yaml
spec:
  backend:
    etcd:
      discoverySrv: "example.com"
```

### etcd.credentialSecretName

`etcd.credentialSecretName` is an optional field that specifies the secret name that contains username and password to use when authenticating with the etcd server. The secret contains the following keys: 

- `username`
- `password`

```yaml
spec:
  backend:
    etcd:
      credentialSecretName: "etcd-credential"
```

### etcd.tlsSecretName

`etcd.tlsSecretName` is an optional field that specifies the secret name that contains TLS assets for etcd communication. The secret contains following keys:

- `tls_ca_file`
- `tls_cert_file`
- `tls_key_file`
