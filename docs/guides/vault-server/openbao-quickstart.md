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

For a Raft- or PostgreSQL-backed production setup, see the [Raft Backend](/docs/guides/vault-server/vault-server-raft-backend.md) and [PostgreSQL Backend](/docs/guides/vault-server/vault-server-postgres-backend.md) guides. For the full `VaultServer` field reference, see the [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) concept doc.

## Before you begin

- Install the KubeVault operator in your cluster: [setup guide](/docs/setup/README.md). This also installs the `kubevault-catalog` chart, which ships a ready-to-use `VaultServerVersion` catalog for OpenBao.
- (Optional but recommended) Install the [KubeVault kubectl plugin](/docs/setup/install/kubectl_plugin.md) — it's used below to fetch the generated root token.

To keep things isolated, use a separate namespace called `demo` throughout this guide:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Pick an OpenBao version

The `kubevault-catalog` chart installs a `VaultServerVersion` for every supported OpenBao release. List them:

```bash
$ kubectl get vaultserverversions | grep openbao
NAME            VERSION   VAULT_IMAGE                  DEPRECATED   AGE
openbao-2.4.3   2.4.3     openbao/openbao:2.4.3                     45h
openbao-2.6.1   2.6.1     openbao/openbao:2.6.1                     45h
openbao-2.7.1   2.7.1     openbao/openbao:2.7.1                     45h
```

This guide uses `openbao-2.7.1`, the latest at the time of writing.

## Deploy a VaultServer

You need to be familiar with one CRD for this guide: [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md).

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: openbao-2.7.1
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

The `inmem` backend keeps everything in memory, so data is lost whenever the pod restarts — fine for this quickstart, but switch to the [Raft](/docs/guides/vault-server/vault-server-raft-backend.md) or [PostgreSQL](/docs/guides/vault-server/vault-server-postgres-backend.md) backend for anything you want to keep.

Deploy it:

```bash
$ kubectl apply -f https://github.com/kubevault/kubevault/raw/{{< param "info.version" >}}/docs/examples/guides/vault-server/vaultserver-openbao.yaml
vaultserver.kubevault.com/vault created
```

Wait for the `STATUS` to become `Running`:

```bash
$ kubectl get vaultserver -n demo -w
NAME    NODES   VERSION         STATUS       AGE
vault   1       openbao-2.7.1   Processing   8s
vault   1       openbao-2.7.1   Sealed       32s
vault   1       openbao-2.7.1   Running      41s
```

Behind the scenes, the KubeVault operator:

- creates a `Deployment` running the OpenBao server,
- creates a `Service` for the OpenBao API (port `8200`) and cluster port (`8201`),
- creates an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) carrying the connection info other KubeVault CRDs (`SecretEngine`, `VaultPolicy`, …) use to reach this server,
- initializes and unseals OpenBao, storing the unseal keys and root token in the `vault-keys` Kubernetes `Secret`,
- enables the Kubernetes auth method and provisions the operator's own auth role.

```bash
$ kubectl get appbinding -n demo
NAME    AGE
vault   54s

$ kubectl get secret -n demo
NAME          TYPE     DATA   AGE
vault-keys    Opaque   6      54s
```

## Talk to OpenBao

Fetch the pod name and the generated root token:

```bash
$ kubectl get pods -n demo
NAME      READY   STATUS    RESTARTS   AGE
vault-0   2/2     Running   0          2m

$ export VAULT_TOKEN=$(kubectl vault root-token get vaultserver vault -n demo --value-only)
```

> No `kubectl vault` plugin? Read the token straight out of the secret instead:
> `export VAULT_TOKEN=$(kubectl get secret -n demo vault-keys -o jsonpath='{.data.vault-root-token}' | base64 -d)`

The OpenBao container ships the `bao` CLI and is already configured to talk to its own API, so `kubectl exec` is the shortest path in — no port-forwarding or TLS certificate wrangling required:

```bash
$ kubectl exec -n demo vault-0 -c vault -- bao status
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    4
Threshold       2
Version         2.7.1
Storage Type    inmem
HA Enabled      false
```

Enable a `KV` secret engine and write a secret, forwarding the root token into the container for this one call:

```bash
$ kubectl exec -n demo vault-0 -c vault -- env VAULT_TOKEN=$VAULT_TOKEN bao secrets enable -path=kv kv
Success! Enabled the kv secrets engine at: kv/

$ kubectl exec -n demo vault-0 -c vault -- env VAULT_TOKEN=$VAULT_TOKEN bao kv put kv/hello message=world
Success! Data written to: kv/hello
```

Read it back:

```bash
$ kubectl exec -n demo vault-0 -c vault -- env VAULT_TOKEN=$VAULT_TOKEN bao kv get kv/hello
====== Data ======
Key        Value
---        -----
message    world
```

If you'd rather use the `bao` binary from your own machine, port-forward instead and point it at `https://127.0.0.1:8200` with `VAULT_SKIP_VERIFY=true` (the server's TLS certificate is self-signed by default):

```bash
$ kubectl port-forward -n demo svc/vault 8200
```

```bash
$ export VAULT_ADDR=https://127.0.0.1:8200
$ export VAULT_SKIP_VERIFY=true
$ bao status
```

## Clean up

```bash
$ kubectl delete ns demo
namespace "demo" deleted
```

Because `terminationPolicy` is `WipeOut`, deleting the `VaultServer` (and this namespace) also removes its underlying storage and secrets.

## Next steps

- Move to a durable backend: [Raft Backend](/docs/guides/vault-server/vault-server-raft-backend.md) or [PostgreSQL Backend](/docs/guides/vault-server/vault-server-postgres-backend.md).
- Enable a database secret engine and issue dynamic credentials: [Secret Engine guides](/docs/guides/secret-engines/postgres/overview.md).
- Run OpenBao across a hub cluster and multiple spokes: [Deploy Vault in a Hub-Spoke Model](/docs/guides/hub-spoke/deploy-hub-spoke.md).
- Configure additional auth methods: [Auth Method](/docs/guides/vault-server/auth-method.md).
