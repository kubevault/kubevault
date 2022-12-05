---
title: Vault Ops Request Overview
menu:
  docs_{{ .version }}:
    identifier: overview-ops-request-guides
    name: Overview
    parent: ops-request-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/README.md).

# Reconfiguring TLS of VaultServer

This guide will give an overview on how KubeVault Enterprise operator reconfigures TLS configuration i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates of a `VaultServer`.

## Before You Begin

- You should be familiar with the following `KubeVault` concepts:
  - [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md)
  - [VaultOpsRequest](/docs/concepts/vault-ops-request/overview.md)

## How Reconfiguring VaultServer TLS Configuration Process Works

The Reconfiguring VaultServer TLS process consists of the following steps:

1. At first, a user creates a `VaultServer` Custom Resource Object (CRO).

2. `KubeVault` operator watches the `VaultServer` CRO.

3. When the operator finds a `VaultServer` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the TLS configuration of the `VaultServer` the user creates a `VaultOpsRequest` CR with desired information.

5. `KubeVault` operator watches the `VaultOpsRequest` CR.

6. When it finds a `VaultOpsRequest` CR, it pauses the `VaultServer` object which is referred from the `VaultOpsRequest`. So, the `KubeVault` operator doesn't perform any operations on the `VaultServer` object during the reconfiguring TLS process.

7. Then the `KubeVault` operator will add, remove, update or rotate TLS configuration based on the Ops Request yaml.

8. Then the `KubeVault` operator will restart all the Pods of the database so that they restart with the new TLS configuration defined in the `VaultOpsRequest` CR.

9. After the successful reconfiguring of the `VaultServer` TLS, the `KubeVault` operator resumes the `VaultServer` object so that the `KubeVault` Community operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on reconfiguring TLS configuration of a VaultServer using `VaultOpsRequest` CRD.