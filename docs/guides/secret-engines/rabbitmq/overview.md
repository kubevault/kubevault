---
title: Manage RabbitMQ credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-rabbitmq
    name: Overview
    parent: rabbitmq-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage RabbitMQ credentials using the KubeVault operator

OpenBao's [`rabbitmq-database-plugin`](https://github.com/sigilr/openbao/pull/8) is a **dynamic-credentials** database plugin for [RabbitMQ](https://www.rabbitmq.com/). The plugin provisions credentials via the [RabbitMQ Management HTTP API](https://www.rabbitmq.com/management.html) (using [`rabbit-hole/v3`](https://github.com/michaelklishin/rabbit-hole)). Each issued credential becomes a native RabbitMQ user whose `tags`, per-vhost permissions, and (optionally) per-topic permissions are taken from the `creationStatements` JSON role document. On lease revocation the plugin deletes the user with `DELETE /api/users/<name>`, which is naturally idempotent — so there is no `revocation_statements` field on the `RabbitMQRole`.

The same CRD shape is used both for the in-process `rabbitmq-database-plugin` and for the hub-spoke `remote-rabbitmq-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-rabbitmq-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [RabbitMQRole](/docs/concepts/secret-engine-crds/database-secret-engine/rabbitmqrole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Run a RabbitMQ cluster with the [management plugin](https://www.rabbitmq.com/management.html) enabled. The official `rabbitmq:3-management` Docker image exposes the HTTP API at `15672`.
- Have an administrator user on RabbitMQ (the default `guest`/`guest` works for `localhost`-only setups; create a real admin for production). The plugin authenticates against the management API as this user when issuing credential operations.
- Decide which RabbitMQ [tags and per-vhost permissions](https://www.rabbitmq.com/access-control.html) the dynamic users should receive (e.g. `administrator`, or per-vhost `configure`/`write`/`read` regex permissions).

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for RabbitMQ

Create an `AppBinding` pointing at the RabbitMQ **management HTTP API** base URL (e.g. `http://rabbitmq.demo.svc:15672`). Unlike most database engines, the URL here is **not** an AMQP URI — it is the management plugin's HTTP base URL, which `rabbit-hole/v3` uses to call the REST endpoints. The referenced Secret carries HTTP Basic Auth credentials (`username` + `password`) used by the plugin to authenticate against the management API.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: rabbitmq
  namespace: demo
spec:
  clientConfig:
    url: http://rabbitmq.demo.svc:15672
  secret:
    name: rabbitmq-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: rabbitmq-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: admin
  password: admin-password
```

> If your RabbitMQ management endpoint uses a self-signed TLS certificate (`https://...:15671`), set `SecretEngine.spec.rabbitmq.insecure: true` below. Drop the knob once you front the management API with a real CA-issued certificate.

## Enable and Configure RabbitMQ Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on a specified path and configure the secret engine with the given configuration.

A sample `SecretEngine` for RabbitMQ:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: rabbitmq-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  rabbitmq:
    databaseRef:
      name: rabbitmq
      namespace: demo
    pluginName: rabbitmq-database-plugin   # optional; this is the default
    allowedRoles:
      - "*"
    # passwordPolicy: my-policy            # optional; name of a Vault password policy
    insecure: false                        # set true only for self-signed dev clusters
```

Apply it and wait for `STATUS=Success`:

```bash
$ kubectl apply -f rabbitmq-engine.yaml
secretengine.engine.kubevault.com/rabbitmq-engine created

$ kubectl get secretengines -n demo
NAME              STATUS    AGE
rabbitmq-engine   Success   10s
```

Use `kubectl describe secretengine -n demo rabbitmq-engine` to inspect error events, if any.

## Create a RabbitMQRole

A [`RabbitMQRole`](/docs/concepts/secret-engine-crds/database-secret-engine/rabbitmqrole.md) describes how the plugin should mint a dynamic credential. `creationStatements` is a single-element string slice holding a JSON role document with any combination of `tags`, `vhosts`, and `vhost_topics`. **At least one of `tags` or `vhosts` must be set.** The `vhosts` map keys are RabbitMQ vhost names (`/` is the default vhost) and each value is the standard RabbitMQ permission triple (`configure` / `write` / `read` regular expressions). See [Access Control](https://www.rabbitmq.com/access-control.html) for the full RabbitMQ authorization model.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: RabbitMQRole
metadata:
  name: rabbitmq-admin
  namespace: demo
spec:
  secretEngineRef:
    name: rabbitmq-engine
  creationStatements:
    - '{"tags":"administrator","vhosts":{"/":{"configure":".*","write":".*","read":".*"}}}'
  defaultTTL: 1h
  maxTTL: 24h
```

A more scoped example that grants only publish/consume on a single vhost (no broker administration):

```yaml
spec:
  creationStatements:
    - '{"vhosts":{"/app":{"configure":"^$","write":"^events\\.","read":"^events\\."}}}'
```

Apply and verify:

```bash
$ kubectl apply -f rabbitmq-role.yaml
rabbitmqrole.engine.kubevault.com/rabbitmq-admin created

$ kubectl get rabbitmqrole -n demo
NAME             STATUS    AGE
rabbitmq-admin   Success   12s
```

The role name in Vault follows the format `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`, so you can verify directly with the Vault CLI:

```bash
$ vault read your-database-path/roles/k8s.-.demo.rabbitmq-admin
Key                      Value
---                      -----
creation_statements      [{"tags":"administrator","vhosts":{"/":{"configure":".*","write":".*","read":".*"}}}]
db_name                  k8s.-.demo.rabbitmq
default_ttl              1h
max_ttl                  24h
```

Deleting the `RabbitMQRole` removes the role from Vault.

## Issue RabbitMQ credentials

Request a dynamic credential by creating a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: rabbitmq-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: RabbitMQRole
    name: rabbitmq-admin
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Approve it through the KubeVault CLI:

```bash
$ kubectl vault approve secretaccessrequest rabbitmq-cred-rqst -n demo
approved
```

Once approved, the operator issues the credential, stores it in a `Secret`, and binds the listed subjects via a `Role`/`RoleBinding`. Internally the plugin calls `PUT /api/users/<name>` to create the user with the generated password, then `PUT /api/users/<name>/...` to apply `tags` and per-vhost permissions exactly as listed in the `creationStatements` JSON. The credential lives on the lease until you delete the `SecretAccessRequest` or it expires; on lease revocation the plugin calls `DELETE /api/users/<name>` (idempotent — so revocation works even if the user has already been removed out-of-band, and no `revocation_statements` are required on the `RabbitMQRole`).

```bash
$ kubectl get secretaccessrequest rabbitmq-cred-rqst -n demo -o json | jq '.status'
{
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.rabbitmq-admin/abc...",
    "renewable": true
  },
  "secret": {
    "name": "rabbitmq-cred-rqst-xxxxxx"
  }
}

$ kubectl get secret -n demo rabbitmq-cred-rqst-xxxxxx -o jsonpath='{.data.username}' | base64 -d
v-kubernetes-demo-XXXXXXXX

$ kubectl get secret -n demo rabbitmq-cred-rqst-xxxxxx -o jsonpath='{.data.password}' | base64 -d
xxxxxxxxxxxxxxxxxx
```

Use the issued `username` / `password` to open an AMQP connection against your RabbitMQ cluster (or hit the management API). The credential is revoked when the `SecretAccessRequest` is deleted.

## Further reading

- RabbitMQ access control: https://www.rabbitmq.com/access-control.html
- OpenBao RabbitMQ plugin: https://github.com/sigilr/openbao/pull/8
