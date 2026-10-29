---
title: Cassandra | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: cassandra-storage
    name: Cassandra
    parent: storage-vault-server-crds
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Cassandra

In the Cassandra storage backend, Vault data will be stored in [Apache Cassandra](https://cassandra.apache.org/). Vault documentation for the Cassandra storage backend can be found [here](https://developer.hashicorp.com/vault/docs/configuration/storage/cassandra).

> `spec.backend.cassandra` is only available on the `kubevault.com/v1alpha2` `VaultServer` API.

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault-with-cassandra
  namespace: demo
spec:
  version: "2.0.0"
  replicas: 1
  backend:
    cassandra:
      hosts: "cassandra-0.cassandra.demo.svc,cassandra-1.cassandra.demo.svc"
      keyspace: "vault"
      credentialSecretRef:
        name: cassandra-cred
      protocolVersion: "4"
```

## spec.backend.cassandra

To use Cassandra as backend storage in Vault, specify `spec.backend.cassandra` in the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    cassandra:
      hosts: <comma_separated_host_list>
      keyspace: <keyspace_name>
      table: <table_name>
      consistency: <consistency_level>
      protocolVersion: <protocol_version>
      credentialSecretRef:
        name: <secret_name>
      simpleRetryPolicyRetries: <retries>
      initialConnectionTimeout: <seconds>
      connectionTimeout: <seconds>
      tlsEnabled: <true/false>
      tlsSecretRef:
        name: <secret_name>
      tlsSkipVerify: <true/false>
      tlsMinVersion: <tls10|tls11|tls12|tls13>
      tlsServerName: <server_name>
      databaseRef:
        name: <appbinding_name>
        namespace: <appbinding_namespace>
```

Here, we are going to describe the various attributes of the `spec.backend.cassandra` field.

### cassandra.hosts

`cassandra.hosts` specifies a comma-separated list of Cassandra hosts to connect to. All hosts must listen on the same port; include the port in each host as `<host>:<port>` if it is not the CQL native protocol default. This is required unless `databaseRef` is set, in which case the hosts are resolved from the referenced `AppBinding` instead.

```yaml
spec:
  backend:
    cassandra:
      hosts: "cassandra-0.cassandra.demo.svc,cassandra-1.cassandra.demo.svc"
```

### cassandra.databaseRef

`cassandra.databaseRef` is an optional field that references the `AppBinding` of a KubeDB-managed Cassandra. When set, the hosts and the credentials are taken from the `AppBinding`, and `credentialSecretRef` is ignored. If `tlsSecretRef` is not set and the `AppBinding` carries TLS material, that is used instead and TLS is enabled automatically.

```yaml
spec:
  backend:
    cassandra:
      databaseRef:
        name: cassandra-appbinding
        namespace: demo
```

### cassandra.keyspace

`cassandra.keyspace` is an optional field that specifies the keyspace used for storing the Vault data. The keyspace must already exist, be reachable, and writable. Defaults to `vault`.

```yaml
spec:
  backend:
    cassandra:
      keyspace: "vault"
```

### cassandra.table

`cassandra.table` is an optional field that specifies the table, in `keyspace`, used for storing the Vault data. The table must already exist. Defaults to `entries`.

```yaml
spec:
  backend:
    cassandra:
      table: "entries"
```

### cassandra.consistency

`cassandra.consistency` is an optional field that specifies the consistency level for read and write operations. Must be one of `ANY`, `ONE`, `TWO`, `THREE`, `QUORUM`, `ALL`, `LOCAL_QUORUM`, `EACH_QUORUM`, or `LOCAL_ONE`.

```yaml
spec:
  backend:
    cassandra:
      consistency: "QUORUM"
```

### cassandra.protocolVersion

`cassandra.protocolVersion` is an optional field that specifies the CQL protocol version to use. Defaults to `2`. Username/password authentication (`credentialSecretRef`) needs `2` or higher.

```yaml
spec:
  backend:
    cassandra:
      protocolVersion: "4"
```

### cassandra.credentialSecretRef

`cassandra.credentialSecretRef` is an optional field that specifies the name of the `Secret`, in the same namespace as the `VaultServer`, containing the username and password used for authentication (`PasswordAuthenticator`). Requires `protocolVersion` of `2` or higher. The secret contains the following keys:

- `username`
- `password`

```yaml
spec:
  backend:
    cassandra:
      credentialSecretRef:
        name: cassandra-cred
```

### cassandra.simpleRetryPolicyRetries

`cassandra.simpleRetryPolicyRetries` is an optional field that, when set to a positive integer, enables Cassandra's `SimpleRetryPolicy` with the given number of retries for queries that time out or fail.

```yaml
spec:
  backend:
    cassandra:
      simpleRetryPolicyRetries: "3"
```

### cassandra.initialConnectionTimeout

`cassandra.initialConnectionTimeout` is an optional field that specifies the timeout, in seconds, for the initial connection to the cluster.

```yaml
spec:
  backend:
    cassandra:
      initialConnectionTimeout: "1"
```

### cassandra.connectionTimeout

`cassandra.connectionTimeout` is an optional field that specifies the timeout, in seconds, for individual queries against the cluster.

```yaml
spec:
  backend:
    cassandra:
      connectionTimeout: "10"
```

### cassandra.tlsEnabled

`cassandra.tlsEnabled` is an optional field that enables a TLS connection to Cassandra. This field accepts a boolean value. The default value is `false`.

```yaml
spec:
  backend:
    cassandra:
      tlsEnabled: true
```

### cassandra.tlsSecretRef

`cassandra.tlsSecretRef` is an optional field that specifies the name of the `Secret`, in the same namespace as the `VaultServer`, containing a PEM-encoded certificate bundle (and, optionally, a private key) used for the TLS connection to Cassandra. Setting this field does not turn TLS on by itself: `tlsEnabled` must also be `true`, or the bundle is mounted but not used. The secret contains the following key:

- `pem_bundle`

```yaml
spec:
  backend:
    cassandra:
      tlsSecretRef:
        name: cassandra-tls
```

### cassandra.tlsSkipVerify

`cassandra.tlsSkipVerify` is an optional field that disables verification of the Cassandra server's certificate chain and host name. Not recommended for production use. This field accepts a boolean value. The default value is `false`.

```yaml
spec:
  backend:
    cassandra:
      tlsSkipVerify: false
```

### cassandra.tlsMinVersion

`cassandra.tlsMinVersion` is an optional field that specifies the minimum acceptable TLS version. One of `tls10`, `tls11`, `tls12`, or `tls13`.

```yaml
spec:
  backend:
    cassandra:
      tlsMinVersion: "tls12"
```

### cassandra.tlsServerName

`cassandra.tlsServerName` is an optional field that specifies the name to verify the certificate of every Cassandra node against.

The Cassandra driver connects to every node by IP: the `hosts` are resolved to IPs, and the other nodes are discovered from the cluster by their IPs. By default it checks each node's certificate against that IP, so a certificate that lists only DNS names, like the one KubeDB issues for all the nodes of a Cassandra, fails with `x509: certificate is valid for ..., not <ip>`. With `tlsServerName`, every node is checked against that name instead. The certificate must still be signed by the CA in `tlsSecretRef`.

If `databaseRef` is set and the TLS material comes from the `AppBinding`, it defaults to the host of the `AppBinding` (`<cassandra>.<namespace>.svc`), which the KubeDB certificate covers. It must be a name the certificate covers, not a wildcard.

```yaml
spec:
  backend:
    cassandra:
      tlsServerName: "cassandra.demo.svc"
```
