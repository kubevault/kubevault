---
title: Manage Qdrant credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-qdrant
    name: Overview
    parent: qdrant-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Qdrant credentials using the KubeVault operator

OpenBao's `qdrant-database-plugin` manages dynamic credentials for [Qdrant](https://qdrant.tech/documentation/guides/security/) using Granular Access API Keys (HS256-signed JSON Web Tokens). When a client requests credentials, OpenBao creates a validation point in Qdrant, signs a JWT carrying collection-level RBAC permissions and a `value_exists` validation claim, and returns the token and dynamic username. When the credential lease expires or is revoked, OpenBao removes the validation point from Qdrant, immediately invalidating the token.

The same CRD shape is used both for the in-process `qdrant-database-plugin` and for the hub-spoke `remote-qdrant-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-qdrant-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [QdrantRole](/docs/concepts/secret-engine-crds/database-secret-engine/qdrant.md)
- [SecretAccessRequest](/docs/concepts/secret-engine-crds/secret-access-request.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Provision a Qdrant deployment with API-key authentication and JWT RBAC enabled (`jwtRbac: true` in KubeDB, or `QDRANT__SERVICE__API_KEY` and `QDRANT__SERVICE__JWT_RBAC=true` in container env).
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

## AppBinding for Qdrant

Create an `AppBinding` pointing at the Qdrant HTTP endpoint. The secret's `password` or `api-key` field carries the Qdrant master API key — KubeVault forwards it to the plugin as `api_key=`, which the plugin uses for reachability checks against `/readyz` and for signing dynamic JWTs.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: qdrant
  namespace: demo
spec:
  clientConfig:
    url: https://qdrant.demo.svc:6333
  secret:
    kind: Secret
    name: qdrant-auth
---
apiVersion: v1
kind: Secret
metadata:
  name: qdrant-auth
  namespace: demo
type: Opaque
stringData:
  api-key: <qdrant-admin-api-key>
```

> Note: If Qdrant runs with mTLS, specify client certificates in the AppBinding or secret.

## Enable and configure the Qdrant secret engine

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: qdrant-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  qdrant:
    databaseRef:
      name: qdrant
      namespace: demo
    pluginName: qdrant-database-plugin
    allowedRoles:
    - "*"
```

Apply and wait for the engine to land:

```bash
$ kubectl apply -f qdrant-secret-engine.yaml
secretengine.engine.kubevault.com/qdrant-engine created

$ kubectl get secretengines -n demo
NAME            STATUS    AGE
qdrant-engine   Success   10s
```

Behind the scenes the KubeVault operator writes:

```
bao write database/config/k8s.<cluster>.demo.qdrant \
    plugin_name=qdrant-database-plugin \
    url=https://qdrant.demo.svc:6333 \
    api_key=<qdrant-api-key> \
    allowed_roles="*"
```

When the referenced AppBinding is `deploymentMode: RemoteAgent`, the operator substitutes `plugin_name=remote-qdrant-plugin` and adds `spoke_name=<spoke>` so the hub forwards the call to the matching `bao relay run` daemon.

## Create a QdrantRole

Define dynamic permissions for the role using `creationStatements`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: QdrantRole
metadata:
  name: app
  namespace: demo
spec:
  secretEngineRef:
    name: qdrant-engine
  defaultTTL: 24h
  maxTTL: 168h
  creationStatements:
    - |
      {
        "access": [
          {
            "collection": "my_collection",
            "access": "rw"
          }
        ]
      }
```

```bash
$ kubectl apply -f qdrant-role.yaml
qdrantrole.engine.kubevault.com/app created

$ kubectl get qdrantrole -n demo
NAME   STATUS    AGE
app    Success   8s
```

## Generate dynamic credentials via SecretAccessRequest

To generate dynamic credentials, create a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: qdrant-credentials
  namespace: demo
spec:
  roleRef:
    kind: QdrantRole
    name: app
  ttl: 1h
```

```bash
$ kubectl apply -f secret-access-request.yaml
secretaccessrequest.engine.kubevault.com/qdrant-credentials created
```

Wait for the request to be approved:

```bash
$ kubectl get secretaccessrequest -n demo qdrant-credentials
NAME                 STATUS     SECRET                       AGE
qdrant-credentials   Approved   qdrant-credentials-xyz123   10s
```

### Inspect the issued credentials

Extract the generated credentials from the Kubernetes Secret:

```bash
$ SECRET_NAME=$(kubectl get secretaccessrequest qdrant-credentials -n demo -o jsonpath='{.status.secret.name}')

$ kubectl get secret $SECRET_NAME -n demo -o yaml
apiVersion: v1
data:
  password: <base64-encoded-jwt>
  username: <base64-encoded-username>
kind: Secret
...
```

The Secret contains:
- `username`: Unique dynamic username (e.g. `v-kubernet-k8s...`).
- `password`: Signed HS256 JWT encoding the collection permissions, lease expiry, and `value_exists` revocation claim.

Clients pass this token to Qdrant via the `api-key` header or `Authorization: Bearer <token>`.

## Revocation

When the `SecretAccessRequest` is deleted or the lease TTL expires, KubeVault revokes the lease in OpenBao. OpenBao deletes the validation point from Qdrant, immediately revoking the token across all Qdrant nodes.

```bash
$ kubectl delete secretaccessrequest -n demo qdrant-credentials
```

## Cleanup

```bash
$ kubectl delete qdrantrole -n demo app
$ kubectl delete secretengine -n demo qdrant-engine
```
