---
title: Manage SAP HANA credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-hanadb
    name: Overview
    parent: hanadb-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage SAP HANA credentials using the KubeVault operator

OpenBao's [`hana-database-plugin`](https://github.com/sigilr/openbao/pull/3) is a **dynamic-credentials** database plugin for [SAP HANA](https://help.sap.com/docs/HANA_PLATFORM/). The plugin was ported from pre-BUSL HashiCorp Vault and is SQL-statement based (the same shape PostgreSQL uses): `HanaDBRole.spec.creationStatements` is a list of HANA SQL statements with `{{name}}`, `{{password}}`, and `{{expiration}}` placeholders, and credentials are issued dynamically through a `SecretAccessRequest`.

The same CRD shape is used both for the in-process `hana-database-plugin` and for the hub-spoke `remote-hana-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-hana-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [HanaDBRole](/docs/concepts/secret-engine-crds/database-secret-engine/hanadbrole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Run a SAP HANA database the operator can reach over the network. Refer to the SAP HANA platform docs at https://help.sap.com/docs/HANA_PLATFORM/ for deployment options.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for SAP HANA

Create an `AppBinding` pointing at the HANA database. The URL is a HANA DSN (`hdb://<user>:<pass>@<host>:<port>`); the referenced Secret carries the username and password the plugin uses to log in as the rotation principal.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: hanadb
  namespace: demo
spec:
  clientConfig:
    url: hdb://hana.demo.svc:39041
  secret:
    name: hanadb-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: hanadb-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: SYSTEM
  password: change-me
```

## Enable and Configure SAP HANA Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on a specified path and configure the secret engine with the given configuration.

A sample `SecretEngine` for SAP HANA:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: hanadb-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  hanadb:
    databaseRef:
      name: hanadb
      namespace: demo
    pluginName: hana-database-plugin   # optional; this is the default
    allowedRoles:
      - "*"
    maxOpenConnections: 4
    maxIdleConnections: 0
    maxConnectionLifetime: 0s
```

Apply it and wait for `STATUS=Success`:

```bash
$ kubectl apply -f hanadb-engine.yaml
secretengine.engine.kubevault.com/hanadb-engine created

$ kubectl get secretengines -n demo
NAME            STATUS    AGE
hanadb-engine   Success   10s
```

Use `kubectl describe secretengine -n demo hanadb-engine` to inspect error events, if any.

## Create a HanaDBRole

A [`HanaDBRole`](/docs/concepts/secret-engine-crds/database-secret-engine/hanadbrole.md) describes how the plugin should mint a dynamic credential. Each entry in `creationStatements` is a HANA SQL statement, executed in sequence with the `{{name}}`, `{{password}}`, and `{{expiration}}` placeholders substituted at credential-issue time.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: HanaDBRole
metadata:
  name: hanadb-readonly
  namespace: demo
spec:
  secretEngineRef:
    name: hanadb-engine
  creationStatements:
    - |
      CREATE USER "{{name}}" PASSWORD "{{password}}" NO FORCE_FIRST_PASSWORD_CHANGE VALID UNTIL '{{expiration}}';
      GRANT SELECT ON SCHEMA APP TO "{{name}}";
  defaultTTL: 1h
  maxTTL: 24h
```

Apply and verify:

```bash
$ kubectl apply -f hanadb-role.yaml
hanadbrole.engine.kubevault.com/hanadb-readonly created

$ kubectl get hanadbrole -n demo
NAME              STATUS    AGE
hanadb-readonly   Success   12s
```

The role name in Vault follows the format `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`, so you can verify directly with the Vault CLI:

```bash
$ vault read your-database-path/roles/k8s.-.demo.hanadb-readonly
Key                      Value
---                      -----
creation_statements      [CREATE USER "{{name}}" PASSWORD "{{password}}" NO FORCE_FIRST_PASSWORD_CHANGE VALID UNTIL '{{expiration}}'; GRANT SELECT ON SCHEMA APP TO "{{name}}";]
db_name                  k8s.-.demo.hanadb
default_ttl              1h
max_ttl                  24h
```

Deleting the `HanaDBRole` removes the role from Vault.

## Issue SAP HANA credentials

Request a dynamic credential by creating a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: hanadb-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: HanaDBRole
    name: hanadb-readonly
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Approve it through the KubeVault CLI:

```bash
$ kubectl vault approve secretaccessrequest hanadb-cred-rqst -n demo
approved
```

Once approved, the operator issues the credential, stores it in a `Secret`, and binds the listed subjects via a `Role`/`RoleBinding`. The credential lives on the lease until you delete the `SecretAccessRequest` or it expires.

```bash
$ kubectl get secretaccessrequest hanadb-cred-rqst -n demo -o json | jq '.status'
{
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.hanadb-readonly/abc...",
    "renewable": true
  },
  "secret": {
    "name": "hanadb-cred-rqst-xxxxxx"
  }
}

$ kubectl get secret -n demo hanadb-cred-rqst-xxxxxx -o jsonpath='{.data.username}' | base64 -d
v-kubernetes-demo-XXXXXXXX

$ kubectl get secret -n demo hanadb-cred-rqst-xxxxxx -o jsonpath='{.data.password}' | base64 -d
xxxxxxxxxxxxxxxxxx
```

Use the issued `username` / `password` to log in to the HANA database; the credential is revoked when the `SecretAccessRequest` is deleted (the plugin runs `revocationStatements`, or falls back to a sensible `DROP USER` if you didn't set any).

## Further reading

- SAP HANA platform docs: https://help.sap.com/docs/HANA_PLATFORM/
- OpenBao SAP HANA plugin: https://github.com/sigilr/openbao/pull/3
