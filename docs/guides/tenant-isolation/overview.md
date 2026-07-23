---
title: Tenant Isolation with OpenBao Namespaces | KubeVault
menu:
  docs_{{ .version }}:
    identifier: tenant-isolation-overview
    name: Overview
    parent: tenant-isolation-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Tenant Isolation with OpenBao Namespaces

KubeVault can automatically place a database's Vault secret engine — its mount, connection
config, roles, and issued credentials — inside a per-tenant **OpenBao namespace**, so each
tenant organization's secrets live together and are isolated from every other org. The
feature is **opt-in and off by default**, and is safe to enable on a running server.

> **Backend requirement:** namespaces are an **OpenBao / Vault Enterprise** capability. On
> Vault OSS the feature stays inert (see [Capability](#capability)).

## Concepts

- **Tenant = organization, keyed by `org-id`.** A tenant is an *organization*, not a single
  Kubernetes namespace. A namespace that belongs to an org is marked with a label and
  annotation:

  ```yaml
  metadata:
    labels:
      ace.appscode.com/client-org: "true"     # this namespace belongs to an org
    annotations:
      ace.appscode.com/org-id: "acme-7f3a"     # the org identity
  ```

  One org can own **several** namespaces, all carrying the same `org-id`. They all map to a
  single OpenBao namespace named `<org-id>`.

- **Two-level opt-in.**
  1. `VaultServer.spec.isolateTenants: true` is the **master gate** for a server.
  2. With the gate on, each `SecretEngine` adopts a namespace based on the org that owns its
     database — no manual `spec.namespace` needed.

  With the gate **off** (the default) everything stays in the root namespace, and an
  explicit `SecretEngine.spec.namespace` is rejected.

- **The `{db}Role` → database → org chain.** A `{db}Role` (`MySQLRole`, `PostgresRole`,
  `MongoDBRole`, …) references a `SecretEngine`; the `SecretEngine` references a database
  AppBinding. The **database's Kubernetes namespace** defines the tenant boundary: if that
  namespace is a client-org, the engine and all its credentials are placed in the org's
  OpenBao namespace. Every `{db}Role` and `SecretAccessRequest` for that engine inherits the
  same namespace.

Supported engines: the database engines that have the org→DB linkage — **MySQL, MariaDB,
Postgres, MongoDB, Redis, Elasticsearch**. AWS/GCP/Azure/PKI/KV are not namespaced by this
feature.

## Enable it

Turn on the master gate on the VaultServer:

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: "1.20.0-openbao"   # a namespace-capable (OpenBao) distribution
  isolateTenants: true
  # …
```

Then label the tenant's database namespaces (usually done by the platform / a cluster
admin, not by tenants):

```bash
kubectl label   ns acme-prod ace.appscode.com/client-org=true
kubectl annotate ns acme-prod ace.appscode.com/org-id=acme-7f3a
```

Now a `SecretEngine` whose database lives in `acme-prod` is provisioned in the OpenBao
namespace `acme-7f3a`, at `acme-7f3a/database/…`, and its dynamic credentials are issued and
revoked there. No `SecretEngine.spec.namespace` is required.

### Explicit namespace (advanced)

With the gate **on**, you may still pin an engine to a specific (possibly hierarchical)
OpenBao namespace:

```yaml
kind: SecretEngine
spec:
  namespace: "acme-7f3a/project-x"   # must already exist in OpenBao; only honored when isolateTenants is on
```

An explicit `spec.namespace` overrides org-derivation and must pre-exist in OpenBao.

## Behavior

| DB namespace | `spec.namespace` | engine state | resolves to | note |
|---|---|---|---|---|
| not a client-org | — | new | **root** | not a tenant |
| client-org `X` | — | new | **X** | auto-derived |
| client-org `X` | — | already in root | **root** | sticky; needs migration to move (below) |
| `X`, later labelled | — | already in root | **root** | a label change never moves a live mount |
| client-org `X` | `acme/proj` | new | **acme/proj** | explicit override wins |
| client-org, **no `org-id`** | — | new | *(blocked)* | `TenantNamespaceUnresolved`, requeues |
| two namespaces, same org `X` | — | new | **X (shared)** | one OpenBao namespace per org |
| gate off | `acme` | new | **root** | explicit namespace rejected (master gate) |
| backend not namespace-capable | any | new | **root** | `TenantIsolationUnsupported` |

Two invariants make this safe:

- **Existing mounts never move on their own.** Turning `isolateTenants` on, adding the
  client-org label later, or turning the gate back off changes what an engine *would*
  resolve to, but never re-mounts a **live** engine (which would drop its leases). Migration
  is explicit — see below.
- **A client-org namespace missing its `org-id` blocks, never falls back to root.** The
  engine is not mounted (condition `TenantNamespaceUnresolved`) so org data never lands in
  the shared root tree.

### Capability

Isolation needs a namespace-capable backend. If the backend is **not** capable:

- **Admission** rejects `isolateTenants: true` on a distribution known to lack namespaces
  (Vault OSS).
- **At runtime**, the VaultServer surfaces `TenantIsolationUnsupported` and SecretEngines
  resolve to root. When capable, it surfaces `TenantIsolationReady`.

## Migrating an existing engine

An engine already mounted in root (for example, provisioned before you enabled isolation)
is **sticky** and reports `TenantMigrationPending`. Moving it is a **one-time, destructive
re-mount that drops the engine's existing leases**, so it is admin-authorized via one of two
annotations.

**Per engine** — move just one SecretEngine:

```bash
kubectl annotate secretengine my-db kubevault.com/migrate-namespace=true
```

**In bulk, per client-org namespace** — move every SecretEngine whose database lives in a
namespace and whose `spec.vaultRef` matches one of the listed Vault servers. The value is a
JSON array of Vault-server AppBinding refs in `namespace/name` form:

```bash
kubectl annotate ns acme-prod \
  'kubevault.com/migrate-vault-secrets=["kv/vault","kv/vault-2"]'
```

On either trigger the operator, for each authorized engine, unmounts the current namespace,
re-mounts under `<org-id>`, emits a **Warning** event that leases were dropped, updates
`status.effectiveNamespace`, and then clears the authorization (removes the per-engine
annotation, prunes processed refs from the namespace array). Both triggers are idempotent —
an engine already at its target is a no-op.

## Conditions & status

`SecretEngine.status.effectiveNamespace` is the source of truth: the OpenBao namespace the
engine is actually provisioned in (empty = root). Every `{db}Role` inherits it.

| Condition | On | Meaning |
|---|---|---|
| `TenantIsolationReady` | VaultServer | gate on and the backend supports namespaces |
| `TenantIsolationUnsupported` | VaultServer | gate on but the backend has no namespace support; isolation is inert |
| `TenantNamespaceUnresolved` | SecretEngine | the DB's namespace is a client-org but has no valid `org-id`; the engine is **not** mounted, requeued |
| `TenantMigrationPending` | SecretEngine | the engine's desired namespace differs from where it is mounted; waiting for a migration annotation |
| `TenantNamespacePendingHub` | SecretEngine | (hub-spoke) the derived org namespace does not yet exist on the hub; requeuing |

## Hub-spoke (OCM) deployments

When a spoke's database is served through the OpenBao spoke relay (see
[Hub-Spoke Deployment](/docs/guides/hub-spoke/)), tenant isolation extends to the spoke: the
engine mounts at `<org-id>/k8s.<spoke>.<type>.<ns>.<name>` on the hub while staying confined
to the spoke's own prefix.

- Enable it by turning on `isolateTenants` on the **hub** `VaultServer` (placement-driven
  spokes inherit the gate automatically); a standalone `VaultRelay` sets
  `spec.isolateTenants` explicitly.
- The spoke derives the org from its **own** local namespace labels; the **hub** creates the
  OpenBao namespace (spokes never create hub namespaces). Until the hub creates it, the
  engine reports `TenantNamespacePendingHub` and requeues.

### How the hub learns which namespaces to create

The hub cannot see the spoke's client-org namespaces directly, so each spoke reports the
OpenBao namespaces it needs through a **`NamespaceSlice`** — a namespaced resource modeled on
Kubernetes `EndpointSlice`. The spoke operator maintains a slice whose `spec.namespaces[]`
lists one entry per required namespace (`name` — the effective namespace, i.e. the org-id;
`externalID` — the org identity; `conditions.ready`), and the hub reads it back over the same
OCM ManifestWork status-feedback channel it already uses for relay health, then idempotently
creates each namespace with `sys/namespaces/<org-id>`. As with `EndpointSlice`, a large set
can shard across several slices, all grouped to their `VaultServer` by the
`kubevault.com/vaultserver-name` and `kubevault.com/vaultserver-namespace` labels.

`NamespaceSlice` is internal plumbing the operator manages automatically — you never create or
edit one; it is described here only to explain the hub-spoke flow.

> **Security requirement:** on hub-spoke, per-spoke isolation is enforced by a Vault policy
> that embeds the cluster name as the literal prefix `k8s.<cluster>.`. This holds **only if
> the OCM managed-cluster name is a strict RFC-1123 DNS label** (lowercase alphanumeric and
> `-`, no dots). KubeVault validates this and fails closed on a non-conforming name.

## Non-goals & limitations

- Only the **database** engines (with an org→DB linkage) are namespaced.
- Auto-derived namespaces are a **single `<org-id>` segment**; hierarchical namespaces
  (`org/project`) are available only via an explicit `spec.namespace`.
- The operator **never garbage-collects** an org's OpenBao namespace when its last engine is
  removed (an empty namespace is cheap and the namespace belongs to the org, which may still
  own other engines).
- Turning `isolateTenants` off is **non-destructive**: it never evacuates an
  already-namespaced engine; it surfaces `TenantMigrationPending` and waits for the annotation.

## Reference

**VaultServer**

- `spec.isolateTenants` (bool, default `false`) — master opt-in.

**SecretEngine**

- `spec.namespace` (string) — explicit OpenBao namespace; only honored when the server's
  `isolateTenants` is on.
- `status.effectiveNamespace` (string) — the namespace the engine is provisioned in.

**Namespace keys**

- label `ace.appscode.com/client-org: "true"` — the namespace belongs to an org.
- annotation `ace.appscode.com/org-id: <slug>` — the org identity (the OpenBao namespace name).

**Migration annotations**

- `kubevault.com/migrate-namespace: "true"` — on a SecretEngine; migrate that engine.
- `kubevault.com/migrate-vault-secrets: '["<ns>/<name>", …]'` — on a client-org Namespace;
  migrate matching engines in bulk.

**NamespaceSlice** (hub-spoke; operator-managed, read-only to users)

- `spec.namespaces[]` — required OpenBao namespaces, each `{name, externalID, conditions.ready}`.
- `status.namespaceCount` — number of entries (shown as the `NamespaceCount` print column).
- labels `kubevault.com/vaultserver-name` + `kubevault.com/vaultserver-namespace` — group
  the slice(s) to their owning `VaultServer`.

## Next steps

- [Deploy Vault in a Hub-Spoke Model](/docs/guides/hub-spoke/)
- [Secret Engine Guides](/docs/guides/secret-engines/)
