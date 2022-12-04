---
title: Vault Ops Request Overview | KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: overview-ops-request-concepts
    name: Overview
    parent: ops-request-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

# VaultOpsRequest

## What is VaultOpsRequest

`VaultOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for `Vault` administrative operations like restart, reconfigure TLS etc. in a Kubernetes native way.

## VaultOpsRequest CRD Specifications

Like any official Kubernetes resource, a `VaultOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and Status sections.

Here, some sample `VaultOpsRequest` CRs for different administrative operations is given below:

Sample `VaultOpsRequest` for restart `VaultServer`: