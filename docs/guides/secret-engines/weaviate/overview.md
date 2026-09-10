---
title: Manage Weaviate credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-weaviate
    name: Overview
    parent: weaviate-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Weaviate credentials using the KubeVault operator

OpenBao's `weaviate-database-plugin` manages dynamic credentials for [Weaviate](https://weaviate.io/developers/weaviate/configuration/authorization) self-hosted using Weaviate's User Management and RBAC REST APIs. When a client requests credentials, OpenBao creates a new database user in Weaviate, assigns the roles specified in `creationStatements`, and returns the dynamic username and generated API key. When the credential lease expires or is revoked, OpenBao deletes the user from Weaviate, immediately revoking access.

The same CRD shape is used both for the in-process `weaviate-database-plugin` and for the hub-spoke `remote-weaviate-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-weaviate-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [WeaviateRole](/docs/concepts/secret-engine-crds/database-secret-engine/weaviate.md)
- [SecretAccessRequest](/docs/concepts/request-crds/secretaccessrequest.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Provision a Weaviate deployment with API-key authentication and RBAC enabled (`AUTHENTICATION_DB_USERS_ENABLED=true` and `AUTHORIZATION_ENABLE_RBAC=true`).
- Create a namespace for testing:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Weaviate

Create an `AppBinding` pointing at the Weaviate HTTP endpoint. The secret's `AUTHENTICATION_APIKEY_ALLOWED_KEYS` field carries the Weaviate admin API key (root user) — KubeVault forwards it to the plugin as `api_key=`, which the plugin uses to authenticate against `/v1/.well-known/ready` and call Weaviate's User Management and RBAC APIs.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: weaviate
  namespace: demo
spec:
  clientConfig:
    url: https://weaviate.demo.svc:8080
  secret:
    kind: Secret
    name: weaviate-auth
---
apiVersion: v1
kind: Secret
metadata:
  name: weaviate-auth
  namespace: demo
type: Opaque
stringData:
  AUTHENTICATION_APIKEY_ALLOWED_KEYS: <weaviate-admin-api-key>
```

> Note: If Weaviate runs with mTLS, specify client certificates in the AppBinding or secret.

## Enable and configure the Weaviate secret engine

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: weaviate-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  weaviate:
    databaseRef:
      name: weaviate
      namespace: demo
    pluginName: weaviate-database-plugin
    allowedRoles:
    - "*"
```

Apply and wait for the engine to land:

```bash
$ kubectl apply -f weaviate-secret-engine.yaml
secretengine.engine.kubevault.com/weaviate-engine created

$ kubectl get secretengines -n demo
NAME              STATUS    AGE
weaviate-engine   Success   10s
```

Behind the scenes the KubeVault operator writes:

```
bao write database/config/k8s.<cluster>.demo.weaviate \
    plugin_name=weaviate-database-plugin \
    url=https://weaviate.demo.svc:8080 \
    api_key=<weaviate-api-key> \
    allowed_roles="*"
```

When the referenced AppBinding is `deploymentMode: RemoteAgent`, the operator substitutes `plugin_name=remote-weaviate-plugin` and adds `spoke_name=<spoke>` so the hub forwards the call to the matching `bao relay run` daemon.

## Create a WeaviateRole

Define dynamic permissions for the role using `creationStatements`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: WeaviateRole
metadata:
  name: app
  namespace: demo
spec:
  secretEngineRef:
    name: weaviate-engine
  defaultTTL: 24h
  maxTTL: 168h
  creationStatements:
    - |
      {
        "roles": [
          "viewer"
        ]
      }
```

Custom roles can also be defined inline under `custom_roles`:

```yaml
  creationStatements:
    - |
      {
        "roles": [
          "viewer"
        ],
        "custom_roles": [
          {
            "name": "customrole",
            "permissions": [
              {
                "action": "read_data",
                "collections": {
                  "collection": "Products"
                }
              }
            ]
          }
        ]
      }
```

```bash
$ kubectl apply -f weaviate-role.yaml
weaviaterole.engine.kubevault.com/app created

$ kubectl get weaviaterole -n demo
NAME   STATUS    AGE
app    Success   8s
```

## Generate dynamic credentials via SecretAccessRequest

To generate dynamic credentials, create a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: weaviate-credentials
  namespace: demo
spec:
  roleRef:
    kind: WeaviateRole
    name: app
  ttl: 1h
```

```bash
$ kubectl apply -f secret-access-request.yaml
secretaccessrequest.engine.kubevault.com/weaviate-credentials created
```

Wait for the request to be approved:

```bash
$ kubectl get secretaccessrequest -n demo weaviate-credentials
NAME                   STATUS     SECRET                         AGE
weaviate-credentials   Approved   weaviate-credentials-xyz123   10s
```

### Inspect the issued credentials

Extract the generated credentials from the Kubernetes Secret:

```bash
$ SECRET_NAME=$(kubectl get secretaccessrequest weaviate-credentials -n demo -o jsonpath='{.status.secret.name}')

$ kubectl get secret $SECRET_NAME -n demo -o yaml
apiVersion: v1
data:
  password: <base64-encoded-api-key>
  username: <base64-encoded-username>
kind: Secret
...
```

The Secret contains:
- `username`: Unique dynamic username (e.g. `v-kubernet-k8s...`).
- `password`: Dynamic API key generated by Weaviate for this user.

Clients present this key to Weaviate in the `Authorization: Bearer <token>` header.

## Revocation

When the `SecretAccessRequest` is deleted or the lease TTL expires, KubeVault revokes the lease in OpenBao. OpenBao deletes the user from Weaviate, immediately revoking access.

```bash
$ kubectl delete secretaccessrequest -n demo weaviate-credentials
```

## Cleanup

```bash
$ kubectl delete weaviaterole -n demo app
$ kubectl delete secretengine -n demo weaviate-engine
```
