---
title: Manage Apache Solr credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-solr
    name: Overview
    parent: solr-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Apache Solr credentials using the KubeVault operator

OpenBao's [`solr-database-plugin`](https://github.com/sigilr/openbao/pull/11) is a **dynamic-credentials** database plugin for [Apache Solr](https://solr.apache.org/). The plugin provisions credentials by talking to Solr's [Security Plugin API](https://solr.apache.org/guide/solr/latest/deployment-guide/authentication-and-authorization-plugins.html): each issued credential becomes a Basic Auth Plugin user and is bound to one or more pre-existing roles on the configured Rule-Based Authorization Plugin via Solr's `set-user-role` API. The plugin **does not create roles**; it only manages users and their role bindings, so every role referenced from `creationStatements` must already exist on the authorizer.

The same CRD shape is used both for the in-process `solr-database-plugin` and for the hub-spoke `remote-solr-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-solr-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [SolrRole](/docs/concepts/secret-engine-crds/database-secret-engine/solrrole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Run an Apache Solr cluster with the Basic Auth Plugin and Rule-Based Authorization Plugin enabled (`security.json` in ZooKeeper / SolrCloud, or `server/solr/security.json` for standalone). The configured admin principal must have permission to add users and assign roles.
- Pre-create the Solr roles you want to bind credentials to (e.g. `admin`, `read`). The plugin only binds — it does not create roles.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Apache Solr

Create an `AppBinding` pointing at the Solr HTTP endpoint. Unlike most database engines, the URL here is **not** a JDBC/connection URI — it is the HTTP(S) base URL of a Solr node (the Security Plugin REST API lives under `/solr/admin/authentication` and `/solr/admin/authorization`). The referenced Secret carries the username and password of a Solr Basic Auth principal that has permission to manage users and assign roles.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: solr
  namespace: demo
spec:
  clientConfig:
    url: http://solr.demo.svc:8983
  secret:
    name: solr-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: solr-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: solr
  password: change-me
```

> If you front the Solr endpoint with a self-signed TLS cert, set `SecretEngine.spec.solr.insecure: true` below. Drop the knob once you front Solr with a real CA-issued certificate.

## Enable and Configure Solr Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on a specified path and configure the secret engine with the given configuration.

A sample `SecretEngine` for Solr:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: solr-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  solr:
    databaseRef:
      name: solr
      namespace: demo
    pluginName: solr-database-plugin   # optional; this is the default
    allowedRoles:
      - "*"
    insecure: false                    # set true only for self-signed dev clusters
```

Apply it and wait for `STATUS=Success`:

```bash
$ kubectl apply -f solr-engine.yaml
secretengine.engine.kubevault.com/solr-engine created

$ kubectl get secretengines -n demo
NAME          STATUS    AGE
solr-engine   Success   10s
```

Use `kubectl describe secretengine -n demo solr-engine` to inspect error events, if any.

## Create a SolrRole

A [`SolrRole`](/docs/concepts/secret-engine-crds/database-secret-engine/solrrole.md) describes how the plugin should mint a dynamic credential. `creationStatements` is a single-element string slice holding a JSON role document of the form `'{"roles":["admin","read"]}'`. The listed roles **must already exist** on the Rule-Based Authorization Plugin — the plugin only binds via `set-user-role`, it does not create roles.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SolrRole
metadata:
  name: solr-readonly
  namespace: demo
spec:
  secretEngineRef:
    name: solr-engine
  creationStatements:
    - '{"roles":["admin","read"]}'
  defaultTTL: 1h
  maxTTL: 24h
```

Apply and verify:

```bash
$ kubectl apply -f solr-role.yaml
solrrole.engine.kubevault.com/solr-readonly created

$ kubectl get solrrole -n demo
NAME            STATUS    AGE
solr-readonly   Success   12s
```

The role name in Vault follows the format `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`, so you can verify directly with the Vault CLI:

```bash
$ vault read your-database-path/roles/k8s.-.demo.solr-readonly
Key                      Value
---                      -----
creation_statements      [{"roles":["admin","read"]}]
db_name                  k8s.-.demo.solr
default_ttl              1h
max_ttl                  24h
```

Deleting the `SolrRole` removes the role from Vault.

## Issue Solr credentials

Request a dynamic credential by creating a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: solr-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: SolrRole
    name: solr-readonly
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Approve it through the KubeVault CLI:

```bash
$ kubectl vault approve secretaccessrequest solr-cred-rqst -n demo
approved
```

Once approved, the operator issues the credential, stores it in a `Secret`, and binds the listed subjects via a `Role`/`RoleBinding`. The plugin creates a new Basic Auth Plugin user on Solr and binds it to `admin` and `read` on the configured Rule-Based Authorization Plugin. The credential lives on the lease until you delete the `SecretAccessRequest` or it expires; on lease revocation the plugin removes the Basic Auth user (Solr's `delete-user` is idempotent, so no separate revocation statements are required).

```bash
$ kubectl get secretaccessrequest solr-cred-rqst -n demo -o json | jq '.status'
{
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.solr-readonly/abc...",
    "renewable": true
  },
  "secret": {
    "name": "solr-cred-rqst-xxxxxx"
  }
}

$ kubectl get secret -n demo solr-cred-rqst-xxxxxx -o jsonpath='{.data.username}' | base64 -d
v-kubernetes-demo-XXXXXXXX

$ kubectl get secret -n demo solr-cred-rqst-xxxxxx -o jsonpath='{.data.password}' | base64 -d
xxxxxxxxxxxxxxxxxx
```

Use the issued `username` / `password` as HTTP Basic-Auth credentials when querying Solr; the credential is revoked when the `SecretAccessRequest` is deleted.

## Further reading

- Solr authentication and authorization plugins: https://solr.apache.org/guide/solr/latest/deployment-guide/authentication-and-authorization-plugins.html
- OpenBao Solr plugin: https://github.com/sigilr/openbao/pull/11
