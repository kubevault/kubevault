---
title: Changelog | KubeVault
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubevault-v2026.10.12
    name: Changelog-v2026.10.12
    parent: welcome
    weight: 20261012
product_name: kubevault
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2026.10.12/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2026.10.12/
---

# KubeVault v2026.10.12 (2026-10-11)


## [kubevault/apimachinery](https://github.com/kubevault/apimachinery)

### [v0.26.0](https://github.com/kubevault/apimachinery/releases/tag/v0.26.0)

- [02e3e47e](https://github.com/kubevault/apimachinery/commit/02e3e47e) Use Go 1.27, k8s 1.37 & debian:13 (#183)
- [a293630d](https://github.com/kubevault/apimachinery/commit/a293630d) Cleanup cves (#182)
- [43b8761b](https://github.com/kubevault/apimachinery/commit/43b8761b) Remove Unsupported Database Secret Engines
- [aa85ded2](https://github.com/kubevault/apimachinery/commit/aa85ded2) engine/v1alpha1: add EtcdRole + EtcdConfiguration CRDs
- [a38185a6](https://github.com/kubevault/apimachinery/commit/a38185a6) Update deps
- [55dfece2](https://github.com/kubevault/apimachinery/commit/55dfece2) pkg/features: add feature-gate registry for database secret engines (#173)
- [343e4250](https://github.com/kubevault/apimachinery/commit/343e4250) Add Weaviate and Qdrant Creation Statement  (#176)
- [61150d2b](https://github.com/kubevault/apimachinery/commit/61150d2b) Migrate code generation to kube_codegen.sh (#174)
- [05fc38eb](https://github.com/kubevault/apimachinery/commit/05fc38eb) Align RabbitMQ Secret Engine with Other Secret Engines (#172)
- [96eef3ae](https://github.com/kubevault/apimachinery/commit/96eef3ae) engine/v1alpha1: fix RabbitMQConfiguration connection_uri field name + add mutual-TLS support (#170)
- [ed8f4361](https://github.com/kubevault/apimachinery/commit/ed8f4361) engine/v1alpha1: sort database-engine switch cases; revert DocumentDB CRDs (#169)
- [71f28bc6](https://github.com/kubevault/apimachinery/commit/71f28bc6) engine/v1alpha1: add ZooKeeperRole + ZooKeeperConfiguration CRDs (#154)
- [3b3b49c2](https://github.com/kubevault/apimachinery/commit/3b3b49c2) engine/v1alpha1: add WeaviateRole + WeaviateConfiguration CRDs (#153)
- [d938c06c](https://github.com/kubevault/apimachinery/commit/d938c06c) engine/v1alpha1: add SolrRole + SolrConfiguration CRDs (#152)
- [4390e2bd](https://github.com/kubevault/apimachinery/commit/4390e2bd) engine/v1alpha1: add RabbitMQRole + RabbitMQConfiguration CRDs (#151)
- [a98fca2e](https://github.com/kubevault/apimachinery/commit/a98fca2e) engine/v1alpha1: add QdrantRole + QdrantConfiguration CRDs (#150)
- [3199e5a4](https://github.com/kubevault/apimachinery/commit/3199e5a4) engine/v1alpha1: add OracleRole + OracleConfiguration CRDs (#149)
- [c49e7a39](https://github.com/kubevault/apimachinery/commit/c49e7a39) engine/v1alpha1: add Neo4jRole + Neo4jConfiguration CRDs (#148)
- [6215c9ff](https://github.com/kubevault/apimachinery/commit/6215c9ff) engine/v1alpha1: add MSSQLServerRole + MSSQLServerConfiguration CRDs (#147)
- [c43b2db8](https://github.com/kubevault/apimachinery/commit/c43b2db8) engine/v1alpha1: add MilvusRole + MilvusConfiguration CRDs (#146)
- [3812dbed](https://github.com/kubevault/apimachinery/commit/3812dbed) engine/v1alpha1: add MemcachedRole + MemcachedConfiguration CRDs (#145)
- [efc9f8ab](https://github.com/kubevault/apimachinery/commit/efc9f8ab) engine/v1alpha1: add KafkaRole + KafkaConfiguration CRDs (#144)
- [31674255](https://github.com/kubevault/apimachinery/commit/31674255) engine/v1alpha1: add IgniteRole + IgniteConfiguration CRDs (#143)
- [6cdd790b](https://github.com/kubevault/apimachinery/commit/6cdd790b) engine/v1alpha1: add HazelcastRole + HazelcastConfiguration CRDs (#142)
- [24ba2f0b](https://github.com/kubevault/apimachinery/commit/24ba2f0b) engine/v1alpha1: add HanaDBRole + HanaDBConfiguration CRDs (#141)
- [3d8376d1](https://github.com/kubevault/apimachinery/commit/3d8376d1) engine/v1alpha1: add DruidRole + DruidConfiguration CRDs (#140)
- [d9686e5e](https://github.com/kubevault/apimachinery/commit/d9686e5e) engine/v1alpha1: add DocumentDBRole + DocumentDBConfiguration CRDs (#139)
- [716791b9](https://github.com/kubevault/apimachinery/commit/716791b9) engine/v1alpha1: add DB2Role + DB2Configuration CRDs (#138)



## [kubevault/cli](https://github.com/kubevault/cli)

### [v0.26.0](https://github.com/kubevault/cli/releases/tag/v0.26.0)

- [2ab37914](https://github.com/kubevault/cli/commit/2ab37914) Prepare for release v0.26.0 (#234)
- [d4cd42cc](https://github.com/kubevault/cli/commit/d4cd42cc) Use Go 1.27, k8s 1.37 & debian:13 (#233)



## [kubevault/crd-manager](https://github.com/kubevault/crd-manager)

### [v0.2.0](https://github.com/kubevault/crd-manager/releases/tag/v0.2.0)

- [4aa48a1](https://github.com/kubevault/crd-manager/commit/4aa48a1) Prepare for release v0.2.0 (#4)
- [4c52981](https://github.com/kubevault/crd-manager/commit/4c52981) Use Go 1.27, k8s 1.37 & debian:13 (#3)
- [aee9142](https://github.com/kubevault/crd-manager/commit/aee9142) Remove Unsupported Database Secret Engines
- [b8373de](https://github.com/kubevault/crd-manager/commit/b8373de) vendor: bump kubevault.dev/apimachinery, add EnsureEtcdRoleCRDs
- [0edf993](https://github.com/kubevault/crd-manager/commit/0edf993) vendor: bump kubevault.dev/apimachinery for default-enabled Postgres/MySQL
- [0257b2d](https://github.com/kubevault/crd-manager/commit/0257b2d) Add CI/release workflows, hack scripts, and Makefile targets
- [0bcbf54](https://github.com/kubevault/crd-manager/commit/0bcbf54) Add README



## [kubevault/installer](https://github.com/kubevault/installer)

### [v2026.10.12](https://github.com/kubevault/installer/releases/tag/v2026.10.12)

- [7899f13a](https://github.com/kubevault/installer/commit/7899f13a) Prepare for release v2026.10.12 (#477)
- [cb142eec](https://github.com/kubevault/installer/commit/cb142eec) Use Go 1.27, k8s 1.37 & debian:13 (#476)
- [94a5ab9c](https://github.com/kubevault/installer/commit/94a5ab9c) Remove Unsupported Database Secret Engines
- [07e03749](https://github.com/kubevault/installer/commit/07e03749) charts: add Etcd to featureGates defaults
- [6aef3965](https://github.com/kubevault/installer/commit/6aef3965) Add kubevault-crd-manager chart, remove Role CRDs from kubevault-operator (#471)
- [7cef6388](https://github.com/kubevault/installer/commit/7cef6388) Makefile: use update-codegen.sh bundled in the gengo-builder image, bump image to release-1.34 (#472)
- [d0e844f6](https://github.com/kubevault/installer/commit/d0e844f6) crds: sync bundled CRDs from kubevault/apimachinery@master (#470)
- [ccaf55ab](https://github.com/kubevault/installer/commit/ccaf55ab) Support overriding kubevault-catalog dir via flag (#469)



## [kubevault/operator](https://github.com/kubevault/operator)

### [v0.26.0](https://github.com/kubevault/operator/releases/tag/v0.26.0)




## [kubevault/unsealer](https://github.com/kubevault/unsealer)

### [v0.26.0](https://github.com/kubevault/unsealer/releases/tag/v0.26.0)

- [6aa0c59a](https://github.com/kubevault/unsealer/commit/6aa0c59a) Use Go 1.27, k8s 1.37 & debian:13 (#165)




