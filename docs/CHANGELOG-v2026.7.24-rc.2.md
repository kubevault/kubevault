---
title: Changelog | KubeVault
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubevault-v2026.7.24-rc.2
    name: Changelog-v2026.7.24-rc.2
    parent: welcome
    weight: 20260724
product_name: kubevault
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2026.7.24-rc.2/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2026.7.24-rc.2/
---

# KubeVault v2026.7.24-rc.2 (2026-07-25)


## [kubevault/apimachinery](https://github.com/kubevault/apimachinery)

### [v0.25.0-rc.2](https://github.com/kubevault/apimachinery/releases/tag/v0.25.0-rc.2)

- [9506b557](https://github.com/kubevault/apimachinery/commit/9506b557) Add tenant-namespace API for automatic OpenBao namespaces (#156)
- [ba64406c](https://github.com/kubevault/apimachinery/commit/ba64406c) Add spec.exposePrimary and the ClientTrafficPinned condition (#164)
- [aecac0c3](https://github.com/kubevault/apimachinery/commit/aecac0c3) Modernize golangci-lint config (#162)
- [0f322b76](https://github.com/kubevault/apimachinery/commit/0f322b76) Rename Vault Agent Leftovers (#161)
- [db14309e](https://github.com/kubevault/apimachinery/commit/db14309e) Use vr as shortname for VaultRelay
- [19e335e6](https://github.com/kubevault/apimachinery/commit/19e335e6) Rename VaultAgent CRD to VaultRelay (agent->relay) (#158)
- [c2df818e](https://github.com/kubevault/apimachinery/commit/c2df818e) Name the spoke kubernetes-auth role k8s-<cluster>-<vs> (#155)
- [e584cf38](https://github.com/kubevault/apimachinery/commit/e584cf38) Add VaultAgent and OCM agent placement API to VaultServer (#137)
- [b177c3ef](https://github.com/kubevault/apimachinery/commit/b177c3ef) Bump go.bytebuilders.dev/audit to v0.0.52 (#135)
- [50440b1b](https://github.com/kubevault/apimachinery/commit/50440b1b) Update go.bytebuilders.dev/audit to v0.0.51 (#134)
- [3b2e0d05](https://github.com/kubevault/apimachinery/commit/3b2e0d05) Add CLAUDE.md pointing to AGENTS.md
- [258f03b3](https://github.com/kubevault/apimachinery/commit/258f03b3) Fix release tracker workflow



## [kubevault/cli](https://github.com/kubevault/cli)

### [v0.25.0-rc.2](https://github.com/kubevault/cli/releases/tag/v0.25.0-rc.2)

- [5461dc81](https://github.com/kubevault/cli/commit/5461dc81) Prepare for release v0.25.0-rc.2 (#230)
- [7866f7d3](https://github.com/kubevault/cli/commit/7866f7d3) Modernize golangci-lint config (#229)
- [0f49286d](https://github.com/kubevault/cli/commit/0f49286d) Drop kubevault.dev/apimachinery/client dependency (#228)
- [3832c2fc](https://github.com/kubevault/cli/commit/3832c2fc) Add CLAUDE.md pointing to AGENTS.md
- [4839a579](https://github.com/kubevault/cli/commit/4839a579) Fix release tracker workflow



## [kubevault/installer](https://github.com/kubevault/installer)

### [v2026.7.24-rc.2](https://github.com/kubevault/installer/releases/tag/v2026.7.24-rc.2)

- [49c7e5fe](https://github.com/kubevault/installer/commit/49c7e5fe) Prepare for release v2026.7.24-rc.2 (#464)
- [f3fb56eb](https://github.com/kubevault/installer/commit/f3fb56eb) Remove update command for kubevault-certified chart (#463)
- [9e4e586d](https://github.com/kubevault/installer/commit/9e4e586d) Fetch kubevault-certified chart dependencies before regenerating it; add OpenShift chart-verify script (#462)
- [1564d541](https://github.com/kubevault/installer/commit/1564d541) Sync CRDs for tenant isolation (spec.isolateTenants, NamespaceSlice) (#461)
- [87362390](https://github.com/kubevault/installer/commit/87362390) Sync vaultserver CRDs for spec.exposePrimary and fix the pods-patch comment (#460)
- [f35219c6](https://github.com/kubevault/installer/commit/f35219c6) Add make refresh target and require it before opening a PR (#459)
- [b8a96caa](https://github.com/kubevault/installer/commit/b8a96caa) Grant the operator pods patch for role labeling (#458)
- [761ddb47](https://github.com/kubevault/installer/commit/761ddb47) Clean up cves
- [62c8c74b](https://github.com/kubevault/installer/commit/62c8c74b) Build kubevault-certified dependencies in update-chart-dependencies.sh (#457)
- [f2fff71b](https://github.com/kubevault/installer/commit/f2fff71b) Add reproducible make targets for catalog and certified charts (#456)
- [33f8eca6](https://github.com/kubevault/installer/commit/33f8eca6) Modernize golangci-lint config (#455)
- [64741473](https://github.com/kubevault/installer/commit/64741473) Add Additional Rules in Cluster Role  (#454)
- [42ae5979](https://github.com/kubevault/installer/commit/42ae5979) Rename VaultAgent to VaultRelay in charts/CRDs (agent->relay) (#452)
- [4e174433](https://github.com/kubevault/installer/commit/4e174433) Support OCM hub-driven VaultAgent placement in operator charts (#449)
- [3ed9e4bb](https://github.com/kubevault/installer/commit/3ed9e4bb) catalog: mark pre-1.10 Vault versions as deprecated (#451)
- [39b17213](https://github.com/kubevault/installer/commit/39b17213) Add audit-token-requester ClusterRoleBinding to charts (#450)
- [7a88fdf7](https://github.com/kubevault/installer/commit/7a88fdf7) Fix charts (#448)
- [b35a72b6](https://github.com/kubevault/installer/commit/b35a72b6) feat(charts): Phase 7 webhook host + --register-crds + VaultServer conversion webhook (#444)
- [5d791450](https://github.com/kubevault/installer/commit/5d791450) Use ace-user-roles v2026.6.12 with audit cluster role (#446)
- [ef9df4ab](https://github.com/kubevault/installer/commit/ef9df4ab) Add docker.io prefix to Docker Hub images (#447)
- [82acef9e](https://github.com/kubevault/installer/commit/82acef9e) Add CLAUDE.md pointing to AGENTS.md



## [kubevault/operator](https://github.com/kubevault/operator)

### [v0.25.0-rc.2](https://github.com/kubevault/operator/releases/tag/v0.25.0-rc.2)

- [5b9c0ed9](https://github.com/kubevault/operator/commit/5b9c0ed90) Merge commit '2fcebcc9b329cba16769d030cd5e5bc666b35b75' into release-0.25



## [kubevault/unsealer](https://github.com/kubevault/unsealer)

### [v0.25.0-rc.2](https://github.com/kubevault/unsealer/releases/tag/v0.25.0-rc.2)

- [0f6a92ca](https://github.com/kubevault/unsealer/commit/0f6a92ca) Modernize golangci-lint config (#162)
- [d2717085](https://github.com/kubevault/unsealer/commit/d2717085) add namespace policy (#161)
- [a34e6891](https://github.com/kubevault/unsealer/commit/a34e6891) Grant policy-controller the relay/* backend paths (hub-spoke placement) (#160)
- [a7fe869a](https://github.com/kubevault/unsealer/commit/a7fe869a) Add CLAUDE.md pointing to AGENTS.md
- [3b5bf9c4](https://github.com/kubevault/unsealer/commit/3b5bf9c4) Fix release tracker workflow




