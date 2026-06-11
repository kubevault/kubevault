---
title: Manage Apache Kafka credentials using the KubeVault operator
menu:
  docs_{{ .version }}:
    identifier: overview-kafka
    name: Overview
    parent: kafka-secret-engines
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Manage Apache Kafka credentials using the KubeVault operator

OpenBao's [`kafka-database-plugin`](https://github.com/sigilr/openbao/pull/15) is a **dynamic-credentials** database plugin for [Apache Kafka](https://kafka.apache.org/). The plugin provisions credentials by writing [SASL/SCRAM](https://kafka.apache.org/documentation/#security_sasl_scram) user records via the franz-go AdminClient: each issued credential becomes a SCRAM-SHA-256 (default) or SCRAM-SHA-512 user on the cluster. ACLs are **not yet implemented** by the plugin — the `acls` field on the role JSON is reserved and must currently be empty; provision ACLs out of band via `kafka-acls.sh`.

The same CRD shape is used both for the in-process `kafka-database-plugin` and for the hub-spoke `remote-kafka-plugin`; the difference is whether the [Vault AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) referenced by `SecretEngine.spec.vaultRef` is marked `deploymentMode: RemoteAgent` (then the SecretEngine controller rewrites `plugin_name` to `remote-kafka-plugin` and attaches `spoke_name`).

You need to be familiar with the following CRDs:

- [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md)
- [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md)
- [KafkaRole](/docs/concepts/secret-engine-crds/database-secret-engine/kafkarole.md)

## Before you begin

- Install KubeVault operator in your cluster from [here](/docs/setup/README.md).
- Run an Apache Kafka cluster with SASL/SCRAM enabled. The plugin authenticates to the brokers using a SASL principal that has permission to write SCRAM credential records (typically the broker's super-user). The broker listener must speak `SASL_PLAINTEXT` or `SASL_SSL`.
- Pre-create any ACLs you want the issued credentials to inherit via `kafka-acls.sh`. The plugin only manages SCRAM user records; ACL management is not yet implemented.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Vault Server

Deploy a Vault Server using the KubeVault operator: [Deploy Vault Server](/docs/guides/vault-server/vault-server.md). The KubeVault operator will create an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) wiring up Kubernetes auth.

```bash
$ kubectl get appbinding -n demo vault -o yaml
```

## AppBinding for Apache Kafka

Create an `AppBinding` pointing at the Kafka cluster. Unlike SQL-style engines, the URL here is **not** a JDBC connection string — it is the **broker CSV** that the franz-go client uses directly (e.g. `broker1:9092,broker2:9092,broker3:9092`). The referenced Secret carries the SASL username and password of the SCRAM principal that has permission to manage user records.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: kafka
  namespace: demo
spec:
  clientConfig:
    url: kafka-0.kafka-broker.demo.svc:9092,kafka-1.kafka-broker.demo.svc:9092,kafka-2.kafka-broker.demo.svc:9092
  secret:
    name: kafka-cred
---
apiVersion: v1
kind: Secret
metadata:
  name: kafka-cred
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: admin
  password: change-me
```

> If the brokers terminate SASL on a TLS listener (`SASL_SSL`), set `SecretEngine.spec.kafka.useTLS: true` below. If the listener cert chains to a private/self-signed CA, also set `insecure: true` to disable verification — drop it once you front the brokers with a real CA-issued certificate.

## Enable and Configure Kafka Secret Engine

When a [SecretEngine](/docs/concepts/secret-engine-crds/secretengine.md) crd object is created, the KubeVault operator will enable a secret engine on a specified path and configure the secret engine with the given configuration.

A sample `SecretEngine` for Kafka:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretEngine
metadata:
  name: kafka-engine
  namespace: demo
spec:
  vaultRef:
    name: vault
  kafka:
    databaseRef:
      name: kafka
      namespace: demo
    pluginName: kafka-database-plugin     # optional; this is the default
    allowedRoles:
      - "*"
    mechanism: SCRAM-SHA-256              # optional; default. SCRAM-SHA-512 is also valid. PLAIN is rejected.
    useTLS: false                         # set true for SASL_SSL listeners
    insecure: false                       # set true only for self-signed dev clusters
```

Apply it and wait for `STATUS=Success`:

```bash
$ kubectl apply -f kafka-engine.yaml
secretengine.engine.kubevault.com/kafka-engine created

$ kubectl get secretengines -n demo
NAME           STATUS    AGE
kafka-engine   Success   10s
```

Use `kubectl describe secretengine -n demo kafka-engine` to inspect error events, if any.

## Create a KafkaRole

A [`KafkaRole`](/docs/concepts/secret-engine-crds/database-secret-engine/kafkarole.md) describes how the plugin should mint a dynamic credential. `creationStatements` is a single-element string slice holding a JSON role document of the form `{"mechanism":"SCRAM-SHA-256","acls":[]}`. The `mechanism` field overrides the SecretEngine-level default on a per-role basis; the `acls` field is reserved and **must be empty** today — the plugin rejects non-empty `acls`. Provision ACLs separately via `kafka-acls.sh`.

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: KafkaRole
metadata:
  name: kafka-producer
  namespace: demo
spec:
  secretEngineRef:
    name: kafka-engine
  creationStatements:
    - '{"mechanism":"SCRAM-SHA-256","acls":[]}'
  defaultTTL: 1h
  maxTTL: 24h
```

Apply and verify:

```bash
$ kubectl apply -f kafka-role.yaml
kafkarole.engine.kubevault.com/kafka-producer created

$ kubectl get kafkarole -n demo
NAME             STATUS    AGE
kafka-producer   Success   12s
```

The role name in Vault follows the format `k8s.{clusterName}.{metadata.namespace}.{metadata.name}`, so you can verify directly with the Vault CLI:

```bash
$ vault read your-database-path/roles/k8s.-.demo.kafka-producer
Key                      Value
---                      -----
creation_statements      [{"mechanism":"SCRAM-SHA-256","acls":[]}]
db_name                  k8s.-.demo.kafka
default_ttl              1h
max_ttl                  24h
```

Deleting the `KafkaRole` removes the role from Vault.

## Issue Kafka credentials

Request a dynamic credential by creating a `SecretAccessRequest`:

```yaml
apiVersion: engine.kubevault.com/v1alpha1
kind: SecretAccessRequest
metadata:
  name: kafka-cred-rqst
  namespace: demo
spec:
  roleRef:
    kind: KafkaRole
    name: kafka-producer
  subjects:
    - kind: ServiceAccount
      name: demo-sa
      namespace: demo
```

Approve it through the KubeVault CLI:

```bash
$ kubectl vault approve secretaccessrequest kafka-cred-rqst -n demo
approved
```

Once approved, the operator issues the credential, stores it in a `Secret`, and binds the listed subjects via a `Role`/`RoleBinding`. The plugin creates a new SCRAM user on the Kafka cluster using the configured `mechanism`. The credential lives on the lease until you delete the `SecretAccessRequest` or it expires; on lease revocation the plugin removes the SCRAM user record.

```bash
$ kubectl get secretaccessrequest kafka-cred-rqst -n demo -o json | jq '.status'
{
  "lease": {
    "duration": "1h0m0s",
    "id": "your-database-path/creds/k8s.-.demo.kafka-producer/abc...",
    "renewable": true
  },
  "secret": {
    "name": "kafka-cred-rqst-xxxxxx"
  }
}

$ kubectl get secret -n demo kafka-cred-rqst-xxxxxx -o jsonpath='{.data.username}' | base64 -d
v-kubernetes-demo-XXXXXXXX

$ kubectl get secret -n demo kafka-cred-rqst-xxxxxx -o jsonpath='{.data.password}' | base64 -d
xxxxxxxxxxxxxxxxxx
```

Use the issued `username` / `password` as your franz-go / sarama / librdkafka client's SASL credentials (`sasl.mechanism=SCRAM-SHA-256`); the credential is revoked when the `SecretAccessRequest` is deleted. ACL bindings for the issued user must already exist on the cluster (see the [Before you begin](#before-you-begin) note about `kafka-acls.sh`).

## Further reading

- Apache Kafka SASL/SCRAM: https://kafka.apache.org/documentation/#security_sasl_scram
- OpenBao Kafka plugin: https://github.com/sigilr/openbao/pull/15
