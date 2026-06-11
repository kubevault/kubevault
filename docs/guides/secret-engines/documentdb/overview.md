---
title: Manage DocumentDB credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-documentdb
    name: Overview
    parent: documentdb-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage DocumentDB credentials using the KubeVault operator

OpenBao's [`documentdb-database-plugin`](https://github.com/sigilr/openbao/pull/9) is a **dynamic-credentials** database plugin for [DocumentDB](https://github.com/documentdb/documentdb), the open-source PostgreSQL extension + gateway that speaks the MongoDB wire protocol (the same engine powers Azure Cosmos DB for MongoDB vCore). The plugin reuses the official `mongo-driver`, so KubeVault treats DocumentDB like the MongoDB engine: the `DocumentDBRole.spec.creationStatements` field is a JSON role document, and credentials are issued dynamically through a `SecretAccessRequest`.

The same CRD shape is used both for the in-process `documentdb-database-plugin` and for the hub-spoke `remote-documentdb-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-documentdb-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [DocumentDBRole](/docs/concepts/secret-engine-crds/database-secret-engine/documentdbrole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Run a DocumentDB gateway. The upstream docker quickstart at https://github.com/documentdb/documentdb publishes the gateway on `:10260` with a self-signed certificate; production deployments should terminate with a real CA-issued cert.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for DocumentDB

Create an `AppBinding` pointing at the DocumentDB gateway. The URL is a standard MongoDB connection URI; the referenced Secret carries the username and password the plugin uses to log in as the rotation principal.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: documentdb
  namespace: demo
spec:
  clientConfig:
    url: mongodb://docdb.demo.svc:10260/?tls=true&tlsInsecure=true
  secret:
    name: documentdb-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: documentdb-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: vault-root
  password: change-me
```

> The `tlsInsecure=true` query parameter pairs with `SecretEngine.spec.documentdb.insecure: true` below — both are intended for the docker quickstart with its self-signed cert. Drop both knobs once you front the gateway with a real CA.

## Enable and Configure DocumentDB Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on a specified path and configure the secret engine with the given configuration.

A sample `SecretEngine` for DocumentDB:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: documentdb-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  documentdb:
    databaseRef:
      name: documentdb
      namespace: demo
    pluginName: documentdb-database-plugin   # optional; this is the default
    allowedRoles:
      - "*"
    writeConcern: '{ "wtimeout": 5000 }'      # optional, mongo-style
    insecure: true                            # docker quickstart with self-signed cert
```

Apply it and wait for `STATUS=Success`:

```bash
$ kubectl apply -f documentdb-engine.yaml
secretengine.engine.kubevault.com/documentdb-engine created

$ kubectl get secretengines -n demo
NAME                STATUS    AGE
documentdb-engine   Success   10s
```

Use `kubectl describe secretengine -n demo documentdb-engine` to inspect error events, if any.

## Create a DocumentDBRole

A [`DocumentDBRole`](/docs/concepts/secret-engine-crds/database-secret-engine/documentdbrole.md) describes how the plugin should mint a dynamic credential. Because DocumentDB speaks the MongoDB wire protocol, `creationStatements` is a JSON role document — the same format the MongoDB engine accepts.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: DocumentDBRole
metadata:
  name: documentdb-readwrite
  namespace: demo
spec:
  secretEngineRef:
    name: documentdb-engine
  creationStatements:
    - '{ "db": "admin", "roles": [{ "role": "readWrite" }] }'
  defaultTTL: 1h
  maxTTL: 24h
```

Apply and verify:

```bash
$ kubectl apply -f documentdb-role.yaml
documentdbrole.engine.kubevault.com/documentdb-readwrite created

$ kubectl get documentdbrole -n demo
NAME                   STATUS    AGE
documentdb-readwrite   Success   12s
```

The role name in Vault follows the format `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`, so you can verify directly with the Vault CLI:

```bash
$ vault read your-database-path/roles/k8s.-.demo.documentdb-readwrite
Key                      Value
---                      -----
creation_statements      [{ "db": "admin", "roles": [{ "role": "readWrite" }] }]
db_name                  k8s.-.demo.documentdb
default_ttl              1h
max_ttl                  24h
```

Deleting the `DocumentDBRole` removes the role from Vault.

## Issue DocumentDB credentials

Request a dynamic credential by creating a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: documentdb-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: DocumentDBRole
    name: documentdb-readwrite
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Approve it through the KubeVault CLI:

```bash
$ kubectl vault approve secretaccessrequest documentdb-cred-rqst -n demo
approved
```

Once approved, the operator issues the credential, stores it in a `Secret`, and binds the listed subjects via a `Role`/`RoleBinding`. The credential lives on the lease until you delete the `SecretAccessRequest` or it expires.

```bash
$ kubectl get secretaccessrequest documentdb-cred-rqst -n demo -o json | jq '.status'
{
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.documentdb-readwrite/abc...",
    "renewable": true
  },
  "secret": {
    "name": "documentdb-cred-rqst-xxxxxx"
  }
}

$ kubectl get secret -n demo documentdb-cred-rqst-xxxxxx -o jsonpath='{.data.username}' | base64 -d
v-kubernetes-demo-XXXXXXXX

$ kubectl get secret -n demo documentdb-cred-rqst-xxxxxx -o jsonpath='{.data.password}' | base64 -d
xxxxxxxxxxxxxxxxxx
```

Use the issued `username` / `password` to log in to the DocumentDB gateway over the mongo wire protocol; the credential is revoked when the `SecretAccessRequest` is deleted.

## Further reading

- DocumentDB engine: https://github.com/documentdb/documentdb
- OpenBao DocumentDB plugin: https://github.com/sigilr/openbao/pull/9
