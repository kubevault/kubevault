---
title: Manage Hazelcast credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-hazelcast
    name: Overview
    parent: hazelcast-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Hazelcast credentials using the KubeVault operator

OpenBao's [`hazelcast-database-plugin`](https://github.com/sigilr/openbao/pull/20) is a **static-credentials-only** database plugin. [Hazelcast](https://docs.hazelcast.com/) OSS has no runtime user-management API — authentication is configured in each member's XML configuration at startup — so the plugin cannot create or delete users on demand. Instead, it pings `/hazelcast/health/ready` with Basic Auth to verify reachability and returns "dynamic credentials are not supported" for `bao read database/creds/<role>`. KubeVault treats Hazelcast like a static-credentials engine: you provision the Hazelcast principal out of band (in `hazelcast.xml`), then use [`HazelcastRole`](/docs/concepts/secret-engine-crds/database-secret-engine/hazelcastrole.md) to attach rotation metadata to that pre-existing principal.

The same CRD shape is used both for the in-process `hazelcast-database-plugin` and for the hub-spoke `remote-hazelcast-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-hazelcast-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [HazelcastRole](/docs/concepts/secret-engine-crds/database-secret-engine/hazelcastrole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Provision a Hazelcast deployment with the [REST health endpoint enabled](https://docs.hazelcast.com/hazelcast/latest/maintain-cluster/monitoring#health-check) (`/hazelcast/health/ready`). The plugin only uses this endpoint for an optional reachability check.
- Configure each Hazelcast principal in `hazelcast.xml` (`<security><realms>...`) for every role you want to rotate. KubeVault will not create the principal.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Hazelcast

Create an `AppBinding` pointing at the Hazelcast member's health endpoint. The username/password Secret (if present) feeds the plugin's Basic Auth healthcheck.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: hazelcast
  namespace: demo
spec:
  clientConfig:
    url: http://hazelcast.demo.svc:5701
  secret:
    name: hazelcast-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: hazelcast-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: bao
  password: <strong-password>
```

## Enable and configure the Hazelcast secret engine

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: hazelcast-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  hazelcast:
    databaseRef:
      name: hazelcast
      namespace: demo
    pluginName: hazelcast-database-plugin
    allowedRoles:
    - "*"
```

Apply and wait for the engine to land:

```bash
$ kubectl apply -f hazelcast-secret-engine.yaml
secretengine.engine.kubevault.com/hazelcast-engine created
$ kubectl get secretengines -n demo
NAME               STATUS    AGE
hazelcast-engine   Success   10s
```

Behind the scenes the KubeVault operator writes:

```
bao write database/config/k8s.<cluster>.demo.hazelcast \
    plugin_name=hazelcast-database-plugin \
    url=http://hazelcast.demo.svc:5701 \
    username=bao \
    password=<strong-password> \
    allowed_roles="*"
```

When the referenced AppBinding is `deploymentMode: RemoteAgent`, the operator substitutes `plugin_name=remote-hazelcast-plugin` and adds `spoke_name=<spoke>` so the hub forwards the call to the matching `bao agent run` daemon. See the [remote-db-plugin DESIGN](https://github.com/sigilr/openbao/blob/db-plugin-hazelcast/plugins/database/remote-db-plugin/DESIGN.md) for the hub-spoke flow.

## Create a HazelcastRole

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: HazelcastRole
metadata:
  name: app
  namespace: demo
spec:
  secretEngineRef:
    name: hazelcast-engine
  defaultTTL: 24h
  maxTTL: 168h
```

```bash
$ kubectl apply -f hazelcast-role.yaml
hazelcastrole.engine.kubevault.com/app created
$ kubectl get hazelcastrole -n demo
NAME   STATUS    AGE
app    Success   8s
```

The KubeVault operator writes the role metadata as `database/roles/k8s.<cluster>.demo.app`. Hazelcast's dynamic creds endpoint will still return the documented "dynamic credentials are not supported" error — that's the plugin contract; pair the `HazelcastRole` with `bao write database/static-roles/<name>` and a pre-existing Hazelcast principal.

## Rotate static credentials

Configure the static-roles binding directly against OpenBao (the static-role API is not currently exposed via a KubeVault CRD):

```bash
$ bao write database/static-roles/k8s.<cluster>.demo.app \
    db_name=k8s.<cluster>.demo.hazelcast \
    username=APP \
    rotation_period=24h
```

`bao read database/static-creds/k8s.<cluster>.demo.app` returns the latest rotated password. KubeVault revokes the role on `HazelcastRole` deletion via the standard finalizer path.

## Cleanup

```bash
$ kubectl delete hazelcastrole -n demo app
$ kubectl delete secretengine -n demo hazelcast-engine
```

## Caveats

- **No dynamic credentials.** `bao read database/creds/<role>` always returns "dynamic credentials are not supported" — Hazelcast OSS has no runtime user-management API.
- **Out-of-band principal management.** Hazelcast principals live in `hazelcast.xml` at member startup; the plugin rotates passwords via `bao` audit but cannot apply them to the Hazelcast member without a config reload.
- **Insecure flag.** `spec.hazelcast.insecure=true` disables TLS verification against the Hazelcast member health endpoint. Use only in dev.
