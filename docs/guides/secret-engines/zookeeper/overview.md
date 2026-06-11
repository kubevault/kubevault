---
title: Manage Apache ZooKeeper credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-zookeeper
    name: Overview
    parent: zookeeper-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Apache ZooKeeper credentials using the KubeVault operator

OpenBao's [`zookeeper-database-plugin`](https://github.com/sigilr/openbao/pull/21) is a **static-credentials-only** database plugin. [Apache ZooKeeper](https://zookeeper.apache.org/doc/current/zookeeperAdmin.html) has no runtime user-management API for SASL/digest principals — they are loaded from server-side `jaas.conf` at startup — so the plugin cannot create or delete users on demand. Instead, it opens a TCP connection to the ZooKeeper endpoint and sends the 4-letter word `ruok`; a healthy node replies `imok`. Note that ZooKeeper 3.5+ requires the `ruok` command to be whitelisted explicitly via `4lw.commands.whitelist=ruok` in `zoo.cfg` (or the env var `ZOO_4LW_COMMANDS_WHITELIST=ruok,stat`); otherwise the healthcheck connect will succeed but the server will close the socket without replying. The plugin returns "dynamic credentials are not supported" for `bao read database/creds/<role>`. KubeVault treats ZooKeeper like a static-credentials engine: you provision the ZooKeeper principal out of band (in the server's `jaas.conf`), then use [`ZooKeeperRole`](/docs/concepts/secret-engine-crds/database-secret-engine/zookeeperrole.md) to attach rotation metadata to that pre-existing principal.

The same CRD shape is used both for the in-process `zookeeper-database-plugin` and for the hub-spoke `remote-zookeeper-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-zookeeper-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [ZooKeeperRole](/docs/concepts/secret-engine-crds/database-secret-engine/zookeeperrole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Provision an Apache ZooKeeper deployment (ensemble or standalone) configured with [SASL/digest authentication](https://zookeeper.apache.org/doc/current/zookeeperProgrammers.html#sc_BuiltinACLSchemes) loaded from `jaas.conf` at server startup. The plugin only uses the TCP endpoint and the `ruok` four-letter-word for a reachability check.
- Whitelist `ruok` (and any other 4lw commands you rely on) in `zoo.cfg`: `4lw.commands.whitelist=ruok` — or set `ZOO_4LW_COMMANDS_WHITELIST=ruok,stat` in the server environment. Without this the healthcheck connect succeeds but receives no `imok` reply.
- Configure each ZooKeeper principal in the server-side `jaas.conf` for every role you want to rotate. KubeVault will not create the principal.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for ZooKeeper

Create an `AppBinding` pointing at the ZooKeeper TCP endpoint. The username/password Secret is forwarded to the plugin for symmetry; the `ruok` healthcheck itself does not authenticate.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: zookeeper
  namespace: demo
spec:
  clientConfig:
    url: tcp://zk.demo.svc:2181
  secret:
    name: zookeeper-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: zookeeper-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: bao
  password: <strong-password>
```

## Enable and configure the ZooKeeper secret engine

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: zookeeper-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  zookeeper:
    databaseRef:
      name: zookeeper
      namespace: demo
    pluginName: zookeeper-database-plugin
    allowedRoles:
    - "*"
```

Apply and wait for the engine to land:

```bash
$ kubectl apply -f zookeeper-secret-engine.yaml
secretengine.engine.kubevault.com/zookeeper-engine created
$ kubectl get secretengines -n demo
NAME               STATUS    AGE
zookeeper-engine   Success   10s
```

Behind the scenes the KubeVault operator writes:

```
bao write database/config/k8s.<cluster>.demo.zookeeper \
    plugin_name=zookeeper-database-plugin \
    url=tcp://zk.demo.svc:2181 \
    username=bao \
    password=<strong-password> \
    allowed_roles="*"
```

When the referenced AppBinding is `deploymentMode: RemoteAgent`, the operator substitutes `plugin_name=remote-zookeeper-plugin` and adds `spoke_name=<spoke>` so the hub forwards the call to the matching `bao agent run` daemon. See the [remote-db-plugin DESIGN](https://github.com/sigilr/openbao/blob/db-plugin-zookeeper/plugins/database/remote-db-plugin/DESIGN.md) for the hub-spoke flow.

## Create a ZooKeeperRole

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: ZooKeeperRole
metadata:
  name: app
  namespace: demo
spec:
  secretEngineRef:
    name: zookeeper-engine
  defaultTTL: 24h
  maxTTL: 168h
```

```bash
$ kubectl apply -f zookeeper-role.yaml
zookeeperrole.engine.kubevault.com/app created
$ kubectl get zookeeperrole -n demo
NAME   STATUS    AGE
app    Success   8s
```

The KubeVault operator writes the role metadata as `database/roles/k8s.<cluster>.demo.app`. ZooKeeper's dynamic creds endpoint will still return the documented "dynamic credentials are not supported" error — that's the plugin contract; pair the `ZooKeeperRole` with `bao write database/static-roles/<name>` and a pre-existing ZooKeeper principal from `jaas.conf`.

## Rotate static credentials

Configure the static-roles binding directly against OpenBao (the static-role API is not currently exposed via a KubeVault CRD):

```bash
$ bao write database/static-roles/k8s.<cluster>.demo.app \
    db_name=k8s.<cluster>.demo.zookeeper \
    username=APP \
    rotation_period=24h
```

`bao read database/static-creds/k8s.<cluster>.demo.app` returns the latest rotated password. KubeVault revokes the role on `ZooKeeperRole` deletion via the standard finalizer path.

## Cleanup

```bash
$ kubectl delete zookeeperrole -n demo app
$ kubectl delete secretengine -n demo zookeeper-engine
```

## Caveats

- **No dynamic credentials.** `bao read database/creds/<role>` always returns "dynamic credentials are not supported" — ZooKeeper has no runtime user-management API for SASL/digest principals.
- **Out-of-band principal management.** ZooKeeper principals live in the server-side `jaas.conf` at startup; the plugin rotates passwords via `bao` audit but cannot apply them to the ZooKeeper server without a config reload.
- **4lw whitelist required.** ZooKeeper 3.5+ ignores `ruok` unless explicitly whitelisted via `4lw.commands.whitelist=ruok` in `zoo.cfg` or `ZOO_4LW_COMMANDS_WHITELIST=ruok,stat` in the environment.
- **Insecure flag.** `spec.zookeeper.insecure=true` disables TLS verification when probing the ZooKeeper TCP endpoint. Use only in dev.
