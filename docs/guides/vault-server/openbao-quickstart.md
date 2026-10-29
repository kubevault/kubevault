---
title: OpenBao Quickstart
menu:
  docs_{{ .version }}:
    identifier: openbao-quickstart
    name: OpenBao Quickstart
    parent: vault-server-guides
    weight: 11
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# OpenBao Quickstart

[OpenBao](https://openbao.org/) is the Linux Foundation's open-source fork of HashiCorp Vault, and it's the distribution KubeVault builds and tests against by default. This guide deploys a single-node OpenBao server with the KubeVault operator and walks through enabling a secret engine and reading/writing a secret — the fastest path to a working OpenBao server on Kubernetes.

For a Raft- or PostgreSQL-backed production setup, see the [Raft Backend](/docs/guides/storage-backends/raft.md) and [PostgreSQL Backend](/docs/guides/storage-backends/postgresql.md) guides. For the full `VaultServer` field reference, see the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) concept doc.

## Before you begin

- Install the KubeVault operator in your cluster: [setup guide](/docs/setup/README.md). This also installs the `kubevault-catalog` chart, which ships a ready-to-use `VaultServerVersion` catalog for OpenBao.
- (Optional but recommended) Install the [KubeVault kubectl plugin](/docs/setup/install/kubectl_plugin.md) — it's used below to fetch the generated root token.

To keep things isolated, use a separate namespace called `demo` throughout this guide:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Pick an OpenBao version

The `kubevault-catalog` chart installs a `VaultServerVersion` for every supported OpenBao release, plus the `sigilr-*` versions of Sigilr, an OpenBao derivative. List them:

```bash
$ kubectl get vaultserverversions | grep -E 'NAME|openbao|sigilr'
NAME             VERSION   VAULT_IMAGE                             DEPRECATED   AGE
openbao-2.4.3    2.4.3     docker.io/openbao/openbao:2.4.3                      45h
openbao-2.6.1    2.6.1     docker.io/openbao/openbao:2.6.1                      45h
openbao-2.7.1    2.7.1     docker.io/openbao/openbao:2.7.1                      45h
sigilr-2.6.1.1   2.6.1     ghcr.io/sigilr/openbao:2.6.1-sigilr.1                45h
sigilr-2.6.1.2   2.6.1     ghcr.io/sigilr/openbao:2.6.1-sigilr.2                45h
```

This guide uses `sigilr-2.6.1.2`, the latest Sigilr version at the time of writing. To use upstream OpenBao instead, set one of the `openbao-*` versions as `spec.version`.

## Deploy a VaultServer

You need to be familiar with one CRD for this guide: [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md).

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: sigilr-2.6.1.2
  replicas: 1
  backend:
    inmem: {}
  unsealer:
    secretShares: 4
    secretThreshold: 2
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```

The `inmem` backend keeps everything in memory, so data is lost whenever the pod restarts. That is fine for this quickstart, but switch to the [Raft](/docs/guides/storage-backends/raft.md) or [PostgreSQL](/docs/guides/storage-backends/postgresql.md) backend for anything you want to keep. A restarted pod comes back empty and uninitialized, and the operator will not initialize it again because `vault-keys` still holds the previous unseal keys, so delete and re-create the `VaultServer` if that happens.

Deploy it:

```bash
$ kubectl apply -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/vault-server/vaultserver-openbao.yaml
vaultserver.kubevault.com/vault created
```

Wait for the `STATUS` to become `Ready`:

```bash
$ kubectl get vaultserver -n demo -w
NAME    REPLICAS   VERSION          STATUS         AGE
vault   1          sigilr-2.6.1.2   Initializing   8s
vault   1          sigilr-2.6.1.2   Ready          41s
```

Behind the scenes, the KubeVault operator:

- creates a `StatefulSet` running the OpenBao server,
- creates a `Service` for the OpenBao API (port `8200`) and cluster port (`8201`),
- creates an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) carrying the connection info other KubeVault CRDs (`SecretEngine`, `VaultPolicy`, …) use to reach this server,
- initializes and unseals OpenBao, storing the unseal keys and root token in the `vault-keys` Kubernetes `Secret`,
- enables the Kubernetes auth method and provisions the operator's own auth role.

```bash
$ kubectl get appbinding -n demo
NAME    AGE
vault   54s

$ kubectl get secret -n demo vault-keys
NAME         TYPE     DATA   AGE
vault-keys   Opaque   5      54s
```

`vault-keys` holds the four unseal keys and the root token, stored under names like `k8s.<cluster-id>.demo.vault-root-token` and `k8s.<cluster-id>.demo.vault-unseal-key-0`.

## Talk to OpenBao

Fetch the pod name and the generated root token:

```bash
$ kubectl get pods -n demo
NAME      READY   STATUS    RESTARTS   AGE
vault-0   2/2     Running   0          2m

$ export VAULT_TOKEN=$(kubectl vault root-token get vaultserver vault -n demo --value-only)
```

> No `kubectl vault` plugin? Read the token straight out of the secret instead (the key name carries a cluster-specific prefix, so select it by its `-root-token` suffix):
> `export VAULT_TOKEN=$(kubectl get secret -n demo vault-keys -o json | jq -r '.data | to_entries[] | select(.key | endswith("-root-token")) | .value' | base64 -d)`

The OpenBao container ships the `bao` CLI, so `kubectl exec` is the shortest path in, with no port-forwarding required. This server has no TLS configured, so its API listens on plain HTTP, while `bao` defaults to `https://127.0.0.1:8200`. Pass `VAULT_ADDR` into the container with each call:

```bash
$ kubectl exec -n demo vault-0 -c vault -- env VAULT_ADDR=http://127.0.0.1:8200 bao status
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    4
Threshold       2
Version         2.6.1-sigilr.2
Storage Type    inmem
Cluster Name    vault-cluster-4ec2d230
Cluster ID      6d1c3b8b-59f7-cebe-6374-8047a9ba8238
HA Enabled      false
```

Enable a `KV` secret engine and write a secret, forwarding the root token into the container for this one call:

```bash
$ kubectl exec -n demo vault-0 -c vault -- env VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=$VAULT_TOKEN bao secrets enable -path=kv kv
Success! Enabled the kv secrets engine at: kv/

$ kubectl exec -n demo vault-0 -c vault -- env VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=$VAULT_TOKEN bao kv put kv/hello message=world
Success! Data written to: kv/hello
```

Read it back:

```bash
$ kubectl exec -n demo vault-0 -c vault -- env VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=$VAULT_TOKEN bao kv get kv/hello
====== Data ======
Key        Value
---        -----
message    world
```

If you'd rather use the `bao` binary from your own machine, port-forward instead and point it at the plain-HTTP listener on `http://127.0.0.1:8200`:

```bash
$ kubectl port-forward -n demo svc/vault 8200
```

```bash
$ export VAULT_ADDR=http://127.0.0.1:8200
$ bao status
```

## Clean up

```bash
$ kubectl delete ns demo
namespace "demo" deleted
```

Deleting the namespace removes the `VaultServer`, its pod and the `vault-keys` secret. Since this server uses the in-memory backend, there is no external storage to clean up.

## Next steps

- Move to a durable backend: [Raft Backend](/docs/guides/storage-backends/raft.md) or [PostgreSQL Backend](/docs/guides/storage-backends/postgresql.md).
- Enable a database secret engine and issue dynamic credentials: [Secret Engine guides](/docs/guides/secret-engines/database/postgres/overview.md).
- Run OpenBao across a hub cluster and multiple spokes: [Deploy Vault in a Hub-Spoke Model](/docs/guides/hub-spoke/deploy-hub-spoke.md).
- Configure additional auth methods: [Auth Method](/docs/guides/vault-server/auth-method.md).
