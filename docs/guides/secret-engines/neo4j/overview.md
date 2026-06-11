---
title: Manage Neo4j credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-neo4j
    name: Overview
    parent: neo4j-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Neo4j credentials using the KubeVault operator

OpenBao's [`neo4j-database-plugin`](https://github.com/sigilr/openbao/pull/10) is a **dynamic-credentials** database plugin for [Neo4j](https://neo4j.com/). The plugin provisions credentials as native Neo4j users created with Cypher's [`CREATE USER`](https://neo4j.com/docs/operations-manual/current/authentication-authorization/manage-users/) statement against the `system` database (via the [`neo4j-go-driver/v5`](https://github.com/neo4j/neo4j-go-driver)). Each issued credential becomes a Neo4j user and is bound to one or more pre-existing roles on the target cluster. The plugin **does not create roles**; it only manages users and their role grants, so every role referenced from `creationStatements` must already exist on the Neo4j cluster.

The same CRD shape is used both for the in-process `neo4j-database-plugin` and for the hub-spoke `remote-neo4j-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-neo4j-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [Neo4jRole](/docs/concepts/secret-engine-crds/database-secret-engine/neo4jrole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Run a Neo4j cluster reachable over the Bolt protocol. The Neo4j [Docker quickstart](https://neo4j.com/docs/operations-manual/current/docker/) exposes Bolt at `7687`.
- Pre-create the Neo4j roles you want to bind credentials to (e.g. `reader`, `publisher`, `editor`). The plugin only grants — it does not create roles. Neo4j Community ships a fixed set of built-in roles (`PUBLIC`, `reader`, `editor`, `publisher`, `architect`, `admin`); Neo4j Enterprise lets you define custom roles via `CREATE ROLE`. See [Authentication and Authorization](https://neo4j.com/docs/operations-manual/current/authentication-authorization/) for the Neo4j RBAC model.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Neo4j

Create an `AppBinding` pointing at the Neo4j Bolt endpoint. Unlike most database engines, the URL here is **not** a JDBC/connection URI — it is a [Bolt URI](https://neo4j.com/docs/bolt/current/bolt/) such as `bolt://host:7687` (plaintext) or `neo4j://host:7687` (routing for cluster deployments). The referenced Secret carries HTTP Basic Auth credentials (`username` + `password`) used by the plugin to authenticate against Neo4j when running its `CREATE USER` Cypher statements on the `system` database.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: neo4j
  namespace: demo
spec:
  clientConfig:
    url: bolt://neo4j.demo.svc:7687
  secret:
    name: neo4j-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: neo4j-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: neo4j
  password: neo4j-password
```

> For a single-instance Neo4j use `bolt://...`; for an Aura/Causal-Cluster deployment use `neo4j://...` so the driver can route to leaders for write Cyphers. If you front Neo4j with a self-signed TLS cert (`bolt+ssc://...` style), set `SecretEngine.spec.neo4j.insecure: true` below. Drop the knob once you front Neo4j with a real CA-issued certificate.

## Enable and Configure Neo4j Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on a specified path and configure the secret engine with the given configuration.

A sample `SecretEngine` for Neo4j:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: neo4j-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  neo4j:
    databaseRef:
      name: neo4j
      namespace: demo
    pluginName: neo4j-database-plugin   # optional; this is the default
    allowedRoles:
      - "*"
    insecure: false                     # set true only for self-signed dev clusters
```

Apply it and wait for `STATUS=Success`:

```bash
$ kubectl apply -f neo4j-engine.yaml
secretengine.engine.kubevault.com/neo4j-engine created

$ kubectl get secretengines -n demo
NAME           STATUS    AGE
neo4j-engine   Success   10s
```

Use `kubectl describe secretengine -n demo neo4j-engine` to inspect error events, if any.

## Create a Neo4jRole

A [`Neo4jRole`](/docs/concepts/secret-engine-crds/database-secret-engine/neo4jrole.md) describes how the plugin should mint a dynamic credential. `creationStatements` is a single-element string slice holding a JSON role document of the form `{"roles":["role1","role2"]}`. The listed roles **must already exist** on the target Neo4j cluster — the plugin only grants, it does not create roles.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: Neo4jRole
metadata:
  name: neo4j-reader
  namespace: demo
spec:
  secretEngineRef:
    name: neo4j-engine
  creationStatements:
    - '{"roles":["reader"]}'
  defaultTTL: 1h
  maxTTL: 24h
```

Apply and verify:

```bash
$ kubectl apply -f neo4j-role.yaml
neo4jrole.engine.kubevault.com/neo4j-reader created

$ kubectl get neo4jrole -n demo
NAME           STATUS    AGE
neo4j-reader   Success   12s
```

The role name in Vault follows the format `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`, so you can verify directly with the Vault CLI:

```bash
$ vault read your-database-path/roles/k8s.-.demo.neo4j-reader
Key                      Value
---                      -----
creation_statements      [{"roles":["reader"]}]
db_name                  k8s.-.demo.neo4j
default_ttl              1h
max_ttl                  24h
```

Deleting the `Neo4jRole` removes the role from Vault.

## Issue Neo4j credentials

Request a dynamic credential by creating a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: neo4j-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: Neo4jRole
    name: neo4j-reader
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Approve it through the KubeVault CLI:

```bash
$ kubectl vault approve secretaccessrequest neo4j-cred-rqst -n demo
approved
```

Once approved, the operator issues the credential, stores it in a `Secret`, and binds the listed subjects via a `Role`/`RoleBinding`. The plugin runs `CREATE USER ... SET PASSWORD ...` on the `system` database and grants the listed roles (`reader`) to that user. The credential lives on the lease until you delete the `SecretAccessRequest` or it expires; on lease revocation the plugin removes the Neo4j user with `DROP USER ... IF EXISTS` (so revocation is idempotent — no `revocation_statements` are required on the `Neo4jRole`).

```bash
$ kubectl get secretaccessrequest neo4j-cred-rqst -n demo -o json | jq '.status'
{
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.neo4j-reader/abc...",
    "renewable": true
  },
  "secret": {
    "name": "neo4j-cred-rqst-xxxxxx"
  }
}

$ kubectl get secret -n demo neo4j-cred-rqst-xxxxxx -o jsonpath='{.data.username}' | base64 -d
v-kubernetes-demo-XXXXXXXX

$ kubectl get secret -n demo neo4j-cred-rqst-xxxxxx -o jsonpath='{.data.password}' | base64 -d
xxxxxxxxxxxxxxxxxx
```

Use the issued `username` / `password` to open a Bolt session against your Neo4j cluster (e.g. via the official drivers or `cypher-shell -u $username -p $password`). The credential is revoked when the `SecretAccessRequest` is deleted.

## Further reading

- Neo4j authentication and authorization: https://neo4j.com/docs/operations-manual/current/authentication-authorization/
- OpenBao Neo4j plugin: https://github.com/sigilr/openbao/pull/10
