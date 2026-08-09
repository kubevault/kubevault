---
title: ZooKeeper | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: zookeeper-storage
    name: ZooKeeper
    parent: storage-vault-server-crds
    weight: 65
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# ZooKeeper

In the ZooKeeper storage backend, Vault data will be stored in [Apache ZooKeeper](https://zookeeper.apache.org/). This backend was removed from upstream OpenBao and is only available in this fork; Vault documentation for the ZooKeeper storage backend can be found [here](https://developer.hashicorp.com/vault/docs/configuration/storage/zookeeper).

> `spec.backend.zookeeper` is only available on the `kubevault.com/v1alpha2` `VaultServer` API. Unlike most other database-backed backends, ZooKeeper has no `haEnabled` toggle: it is unconditionally usable in Vault's high availability mode, the same as `raft`/`consul`.

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault-with-zookeeper
  namespace: demo
spec:
  version: "2.0.0"
  replicas: 1
  backend:
    zookeeper:
      address: "zookeeper-0.zookeeper.demo.svc:2181,zookeeper-1.zookeeper.demo.svc:2181"
      path: "vault/"
```

## spec.backend.zookeeper

To use ZooKeeper as backend storage in Vault, specify `spec.backend.zookeeper` in the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    zookeeper:
      address: <comma_separated_address_list>
      path: <znode_path>
      znodeOwner: <acl_scheme:id>
      authInfoSecretRef:
        name: <secret_name>
      tlsEnabled: <true/false>
      tlsSecretRef:
        name: <secret_name>
      tlsSkipVerify: <true/false>
      tlsVerifyIP: <true/false>
      tlsMinVersion: <tls10|tls11|tls12|tls13>
```

Here, we are going to describe the various attributes of the `spec.backend.zookeeper` field.

### zookeeper.address

`zookeeper.address` is an optional field that specifies the addresses of the ZooKeeper instances as a comma-separated list. Defaults to `localhost:2181`.

```yaml
spec:
  backend:
    zookeeper:
      address: "zookeeper-0.zookeeper.demo.svc:2181,zookeeper-1.zookeeper.demo.svc:2181"
```

### zookeeper.path

`zookeeper.path` is an optional field that specifies the path in ZooKeeper's tree where Vault data will be stored. Defaults to `vault/`.

```yaml
spec:
  backend:
    zookeeper:
      path: "vault/"
```

### zookeeper.znodeOwner

`zookeeper.znodeOwner` is an optional field that specifies the ACL `scheme:id` applied to every znode Vault creates. Defaults to `world:anyone`, i.e. unrestricted access.

```yaml
spec:
  backend:
    zookeeper:
      znodeOwner: "digest:vault:base64EncodedHash"
```

### zookeeper.authInfoSecretRef

`zookeeper.authInfoSecretRef` is an optional field that specifies the name of the `Secret`, in the same namespace as the `VaultServer`, containing a `scheme:auth` pair passed to ZooKeeper's `AddAuth` API immediately after connecting, so the client authenticates as a specific principal. The secret contains the following key:

- `authInfo`

```yaml
spec:
  backend:
    zookeeper:
      authInfoSecretRef:
        name: zookeeper-auth
```

### zookeeper.tlsEnabled

`zookeeper.tlsEnabled` is an optional field that enables a TLS connection to ZooKeeper. This field accepts a boolean value. The default value is `false`.

```yaml
spec:
  backend:
    zookeeper:
      tlsEnabled: true
```

### zookeeper.tlsSecretRef

`zookeeper.tlsSecretRef` is an optional field that specifies the name of the `Secret`, in the same namespace as the `VaultServer`, that contains `ca.crt`, `tls.crt`, and `tls.key` for ZooKeeper communication. The secret contains the following keys:

- `ca.crt`
- `tls.crt`
- `tls.key`

```yaml
spec:
  backend:
    zookeeper:
      tlsSecretRef:
        name: zookeeper-tls
```

### zookeeper.tlsSkipVerify

`zookeeper.tlsSkipVerify` is an optional field that disables verification of the ZooKeeper server's certificate chain and host name. Not recommended for production use. This field accepts a boolean value. The default value is `false`.

```yaml
spec:
  backend:
    zookeeper:
      tlsSkipVerify: false
```

### zookeeper.tlsVerifyIP

`zookeeper.tlsVerifyIP` is an optional field that, when set, verifies the server's IP address against the certificate instead of its DNS name. Only consulted when `tlsSkipVerify` is `false`. This field accepts a boolean value. The default value is `false`.

```yaml
spec:
  backend:
    zookeeper:
      tlsVerifyIP: false
```

### zookeeper.tlsMinVersion

`zookeeper.tlsMinVersion` is an optional field that specifies the minimum acceptable TLS version. One of `tls10`, `tls11`, `tls12`, or `tls13`.

```yaml
spec:
  backend:
    zookeeper:
      tlsMinVersion: "tls12"
```
