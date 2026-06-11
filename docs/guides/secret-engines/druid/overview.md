---
title: Manage Apache Druid credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-druid
    name: Overview
    parent: druid-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Apache Druid credentials using the KubeVault operator

OpenBao's [`druid-database-plugin`](https://github.com/sigilr/openbao/pull/12) is a **dynamic-credentials** database plugin for [Apache Druid](https://druid.apache.org/). The plugin provisions credentials by talking to Druid's [BasicSecurity](https://druid.apache.org/docs/latest/operations/security-overview/) coordinator REST API: each issued credential becomes a Druid authenticator user and is bound to one or more pre-existing roles on the configured Druid authorizer. The plugin **does not create roles**; it only manages users and their role bindings, so every role referenced from `creationStatements` must already exist on the authorizer.

The same CRD shape is used both for the in-process `druid-database-plugin` and for the hub-spoke `remote-druid-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-druid-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [DruidRole](/docs/concepts/secret-engine-crds/database-secret-engine/druidrole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Run an Apache Druid cluster with BasicSecurity enabled. The Druid [Quick Start](https://druid.apache.org/docs/latest/tutorials/) ships with `MyBasicMetadataAuthenticator` and `MyBasicMetadataAuthorizer` already configured; for production, configure your own authenticator and authorizer through the Druid runtime properties and pass their names via `authenticator` / `authorizer` on the `SecretEngine` below.
- Pre-create the Druid roles you want to bind credentials to (e.g. `role1`, `role2`). The plugin only binds — it does not create roles.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Apache Druid

Create an `AppBinding` pointing at the Druid coordinator's REST endpoint. Unlike most database engines, the URL here is **not** a JDBC/connection URI — it is the HTTP(S) base URL of the Druid coordinator (the BasicSecurity REST API lives under `/druid-ext/basic-security/...`). The referenced Secret carries the username and password of a Druid BasicSecurity principal that has permission to manage authenticator users.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: druid
  namespace: demo
spec:
  clientConfig:
    url: http://druid-coordinator.demo.svc:8081
  secret:
    name: druid-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: druid-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: admin
  password: change-me
```

> If you front the Druid coordinator with a self-signed TLS cert (e.g. the Quick Start with TLS enabled), set `SecretEngine.spec.druid.insecure: true` below. Drop the knob once you front the coordinator with a real CA-issued certificate.

## Enable and Configure Druid Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on a specified path and configure the secret engine with the given configuration.

A sample `SecretEngine` for Druid:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: druid-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  druid:
    databaseRef:
      name: druid
      namespace: demo
    pluginName: druid-database-plugin            # optional; this is the default
    allowedRoles:
      - "*"
    authenticator: MyBasicMetadataAuthenticator  # optional; this is the Quick Start default
    authorizer: MyBasicMetadataAuthorizer        # optional; this is the Quick Start default
    insecure: false                              # set true only for self-signed dev clusters
```

Apply it and wait for `STATUS=Success`:

```bash
$ kubectl apply -f druid-engine.yaml
secretengine.engine.kubevault.com/druid-engine created

$ kubectl get secretengines -n demo
NAME           STATUS    AGE
druid-engine   Success   10s
```

Use `kubectl describe secretengine -n demo druid-engine` to inspect error events, if any.

## Create a DruidRole

A [`DruidRole`](/docs/concepts/secret-engine-crds/database-secret-engine/druidrole.md) describes how the plugin should mint a dynamic credential. `creationStatements` is a single-element string slice holding a JSON role document of the form `{"roles":["role1","role2"]}`. The listed roles **must already exist** on the authorizer named in `SecretEngine.spec.druid.authorizer` — the plugin only binds, it does not create roles.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: DruidRole
metadata:
  name: druid-readonly
  namespace: demo
spec:
  secretEngineRef:
    name: druid-engine
  creationStatements:
    - '{"roles":["role1","role2"]}'
  defaultTTL: 1h
  maxTTL: 24h
```

Apply and verify:

```bash
$ kubectl apply -f druid-role.yaml
druidrole.engine.kubevault.com/druid-readonly created

$ kubectl get druidrole -n demo
NAME             STATUS    AGE
druid-readonly   Success   12s
```

The role name in Vault follows the format `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`, so you can verify directly with the Vault CLI:

```bash
$ vault read your-database-path/roles/k8s.-.demo.druid-readonly
Key                      Value
---                      -----
creation_statements      [{"roles":["role1","role2"]}]
db_name                  k8s.-.demo.druid
default_ttl              1h
max_ttl                  24h
```

Deleting the `DruidRole` removes the role from Vault.

## Issue Druid credentials

Request a dynamic credential by creating a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: druid-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: DruidRole
    name: druid-readonly
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Approve it through the KubeVault CLI:

```bash
$ kubectl vault approve secretaccessrequest druid-cred-rqst -n demo
approved
```

Once approved, the operator issues the credential, stores it in a `Secret`, and binds the listed subjects via a `Role`/`RoleBinding`. The plugin creates a new authenticator user on Druid and binds it to `role1` and `role2` on the configured authorizer. The credential lives on the lease until you delete the `SecretAccessRequest` or it expires; on lease revocation the plugin removes the authenticator user.

```bash
$ kubectl get secretaccessrequest druid-cred-rqst -n demo -o json | jq '.status'
{
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.druid-readonly/abc...",
    "renewable": true
  },
  "secret": {
    "name": "druid-cred-rqst-xxxxxx"
  }
}

$ kubectl get secret -n demo druid-cred-rqst-xxxxxx -o jsonpath='{.data.username}' | base64 -d
v-kubernetes-demo-XXXXXXXX

$ kubectl get secret -n demo druid-cred-rqst-xxxxxx -o jsonpath='{.data.password}' | base64 -d
xxxxxxxxxxxxxxxxxx
```

Use the issued `username` / `password` as HTTP Basic-Auth credentials when querying Druid (Brokers, Router, etc.); the credential is revoked when the `SecretAccessRequest` is deleted.

## Further reading

- Apache Druid security overview: https://druid.apache.org/docs/latest/operations/security-overview/
- OpenBao Druid plugin: https://github.com/sigilr/openbao/pull/12
