---
title: NamespaceSlice | KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: namespaceslice-vault-server-crds
    name: NamespaceSlice
    parent: vault-server-crds-concepts
    weight: 14
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# NamespaceSlice

## What is NamespaceSlice

A `NamespaceSlice` is a Kubernetes `CustomResourceDefinition` (CRD) the KubeVault operator uses internally to report the OpenBao namespaces a [hub-spoke](/docs/guides/hub-spoke/deploy-hub-spoke.md) spoke needs created on the hub, as part of [tenant isolation](/docs/guides/tenant-isolation/overview.md). It is modeled on the Kubernetes `EndpointSlice`: a large set of required namespaces is split across multiple `NamespaceSlice` objects (shards), each grouped back to the owning `VaultServer` by the `kubevault.com/vaultserver-name` and `kubevault.com/vaultserver-namespace` labels — the same way an `EndpointSlice` groups to a `Service` via `kubernetes.io/service-name`.

`NamespaceSlice` is **internal plumbing that the operator manages automatically — you never create or edit one**. It exists to solve a specific problem in the hub-spoke topology: the hub cannot reach a spoke's Kubernetes API server directly, so it has no way to see which of the spoke's database namespaces carry the `ace.appscode.com/client-org` label and therefore need an OpenBao namespace. `NamespaceSlice` is the channel that carries that information from spoke to hub over OCM.

## How it's used

1. The **spoke-side operator** resolves the org-id for each local client-org database (see [tenant isolation](/docs/guides/tenant-isolation/overview.md)) and records the deduplicated, validated set it needs on the hub in one or more `NamespaceSlice` shards, under `spec.namespaces[]`.
2. The **hub** pushes an empty `NamespaceSlice` per shard it currently tracks in the spoke's `ManifestWork`, setting only `spec.hubVaultRef` — it deliberately never sets `spec.namespaces`, so server-side-apply field ownership leaves the spoke's reported set intact (the hub owns `hubVaultRef`; the spoke owns `namespaces`). The hub reads `spec.namespaces` back through OCM ManifestWork **status feedback**, re-validates every entry, and idempotently creates each namespace with `sys/namespaces/<name>`.
3. Until the hub creates it, a spoke `SecretEngine` that needs the namespace reports the `TenantNamespacePendingHub` condition and requeues. This is eventually consistent and self-healing.

### Sharding

A single `NamespaceSlice` shard holds up to `maxSpokeNamespaces` entries (operator flag `--max-spoke-namespaces`, default `256`). Once a shard fills up, the spoke rolls the overflow into the next shard (`<hub-appbinding>-1`, `-2`, …), and the hub grows how many shards it tracks — and reads back via ManifestWork status feedback — by at most one shard per reconcile, up to `maxNamespaceSliceShards` (operator flag `--max-namespace-slice-shards`, default `32`). A spoke reporting more namespaces than `maxSpokeNamespaces × maxNamespaceSliceShards` has the overflow dropped, with a warning logged by the operator.

## NamespaceSlice CRD Specification

Like any official Kubernetes resource, a `NamespaceSlice` object has `TypeMeta`, `ObjectMeta`, `Spec`, and `Status` sections.

A sample `NamespaceSlice` object is shown below:

```yaml
apiVersion: kubevault.com/v1alpha2
kind: NamespaceSlice
metadata:
  name: vault-agent-hub-vault-0
  namespace: demo
  labels:
    kubevault.com/vaultserver-name: vault
    kubevault.com/vaultserver-namespace: demo
spec:
  hubVaultRef:
    name: vault
    namespace: demo
  namespaces:
  - name: acme-7f3a
    externalID: acme-7f3a
    conditions:
      ready: true
status:
  observedGeneration: 3
  namespaceCount: 1
```

### NamespaceSlice Spec

#### spec.hubVaultRef

`spec.hubVaultRef` identifies the hub `VaultServer` this shard's namespaces should be created against — the same `VaultServer` the `kubevault.com/vaultserver-name` + `kubevault.com/vaultserver-namespace` labels group the shard to, so the ref and the labels always agree. It is stamped by the hub; the spoke never sets it.

#### spec.namespaces

`spec.namespaces` is this shard's list of required OpenBao namespaces — the analogue of a single `Endpoint` in an `EndpointSlice`. Each entry has:

- `name`: the effective OpenBao namespace to provision. In the current tenant-isolation model this is the org-id.
- `externalID`: the external identity this namespace maps to (the KubeDB Platform org-id, in string form), letting the effective namespace name diverge from the org-id in the future without changing the association.
- `conditions.ready`: whether this namespace entry is validated and required, as reported by the spoke. A namespace explicitly marked not-ready is skipped by the hub.

Set only by the spoke; the hub never writes it.

### NamespaceSlice Status

- `observedGeneration`: the most recent generation of the object the spoke has reconciled.
- `namespaceCount`: mirrors `len(spec.namespaces)` — a cheap print column (`kubectl get nss`) and a bounded status-feedback scalar.

## Next Steps

- Learn about [Tenant Isolation with OpenBao Namespaces](/docs/guides/tenant-isolation/overview.md), including the hub-spoke section describing how the hub learns which namespaces to create.
- Deploy the full hub-spoke model with OCM: [guide](/docs/guides/hub-spoke/deploy-hub-spoke.md).
- Learn about the [VaultRelay](/docs/concepts/vault-server-crds/vaultrelay.md) CRD.
