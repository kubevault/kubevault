---
title: Changelog | KubeVault
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubevault-v2026.8.7
    name: Changelog-v2026.8.7
    parent: welcome
    weight: 20260807
product_name: kubevault
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2026.8.7/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2026.8.7/
---

# KubeVault v2026.8.7 (2026-08-07)


## [kubevault/apimachinery](https://github.com/kubevault/apimachinery)

### [v0.25.0](https://github.com/kubevault/apimachinery/releases/tag/v0.25.0)

- [599a2bed](https://github.com/kubevault/apimachinery/commit/599a2bed) chore(deps): go mod tidy && go mod vendor (#168)
- [1a0e4147](https://github.com/kubevault/apimachinery/commit/1a0e4147) Add Sigilr as an OpenBao-derivative Vault distro (#167)
- [4efa732f](https://github.com/kubevault/apimachinery/commit/4efa732f) Remove VaultRelay.spec.image; resolve it from the hub AppBinding's spec.version (#166)
- [a4f10f8d](https://github.com/kubevault/apimachinery/commit/a4f10f8d) Add Namespace Slice Helpers (#165)
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
- [d03a6c7c](https://github.com/kubevault/apimachinery/commit/d03a6c7c) Add AGENTS.md (#133)
- [075c9c56](https://github.com/kubevault/apimachinery/commit/075c9c56) Add 1gtm-app[bot] to kodiak auto_approve_usernames (#132)
- [3eac0dd1](https://github.com/kubevault/apimachinery/commit/3eac0dd1) Cleanup cves (#131)
- [a1a8b7b2](https://github.com/kubevault/apimachinery/commit/a1a8b7b2) Harden CI workflows (#130)
- [7e04d17f](https://github.com/kubevault/apimachinery/commit/7e04d17f) Configure dependabot refresh schedule (#128)



## [kubevault/cli](https://github.com/kubevault/cli)

### [v0.25.0](https://github.com/kubevault/cli/releases/tag/v0.25.0)

- [a64b70d8](https://github.com/kubevault/cli/commit/a64b70d8) Prepare for release v0.25.0 (#231)
- [f6bcff38](https://github.com/kubevault/cli/commit/f6bcff38) Clean up deps
- [5461dc81](https://github.com/kubevault/cli/commit/5461dc81) Prepare for release v0.25.0-rc.2 (#230)
- [7866f7d3](https://github.com/kubevault/cli/commit/7866f7d3) Modernize golangci-lint config (#229)
- [0f49286d](https://github.com/kubevault/cli/commit/0f49286d) Drop kubevault.dev/apimachinery/client dependency (#228)
- [3832c2fc](https://github.com/kubevault/cli/commit/3832c2fc) Add CLAUDE.md pointing to AGENTS.md
- [4839a579](https://github.com/kubevault/cli/commit/4839a579) Fix release tracker workflow
- [61d31fd8](https://github.com/kubevault/cli/commit/61d31fd8) Prepare for release v0.25.0-rc.1 (#227)
- [ca45ac56](https://github.com/kubevault/cli/commit/ca45ac56) Add AGENTS.md (#226)
- [4ed7b9d5](https://github.com/kubevault/cli/commit/4ed7b9d5) Harden CI workflows (#225)
- [b42d2b9b](https://github.com/kubevault/cli/commit/b42d2b9b) Prepare for release v0.25.0-rc.0 (#224)
- [c0234062](https://github.com/kubevault/cli/commit/c0234062) Cleanup cves (#223)
- [00b5938d](https://github.com/kubevault/cli/commit/00b5938d) Harden CI workflows (#222)
- [289a185b](https://github.com/kubevault/cli/commit/289a185b) Configure dependabot refresh schedule (#220)



## [kubevault/installer](https://github.com/kubevault/installer)

### [v2026.8.7](https://github.com/kubevault/installer/releases/tag/v2026.8.7)

- [31143fe9](https://github.com/kubevault/installer/commit/31143fe9) Prepare for release v2026.8.7 (#467)
- [ef9f89db](https://github.com/kubevault/installer/commit/ef9f89db) Add OpenBao Version 2.6.1 (#466)
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
- [ad0ee4fa](https://github.com/kubevault/installer/commit/ad0ee4fa) Prepare for release v2026.5.18-rc.1 (#443)
- [14ccba08](https://github.com/kubevault/installer/commit/14ccba08) Fix release tracker workflow
- [4e8a272d](https://github.com/kubevault/installer/commit/4e8a272d) Add AGENTS.md (#442)
- [80830cdf](https://github.com/kubevault/installer/commit/80830cdf) publish-oci.yml: replace GHCRX app token with LGTM_GITHUB_TOKEN (#441)
- [a7636bb6](https://github.com/kubevault/installer/commit/a7636bb6) Remove bzr install from workflows (#440)
- [2bb52809](https://github.com/kubevault/installer/commit/2bb52809) Prepare for release v2026.5.14-rc.0 (#437)
- [81043cc4](https://github.com/kubevault/installer/commit/81043cc4) Pin docker/login-action to v4.1.0 (#439)
- [7420547c](https://github.com/kubevault/installer/commit/7420547c) Harden CI workflows (#438)
- [cbe0e016](https://github.com/kubevault/installer/commit/cbe0e016) Cleanup cves (#435)
- [4bf42d14](https://github.com/kubevault/installer/commit/4bf42d14) Harden CI workflows (#434)
- [a1e1f7bf](https://github.com/kubevault/installer/commit/a1e1f7bf) Update cve report (#431)
- [1471293b](https://github.com/kubevault/installer/commit/1471293b) Update cve report (#430)
- [b1076760](https://github.com/kubevault/installer/commit/b1076760) Update cve report (#429)
- [3a063b0d](https://github.com/kubevault/installer/commit/3a063b0d) Update cve report (#428)
- [bad352e3](https://github.com/kubevault/installer/commit/bad352e3) Update cve report (#427)
- [258774cc](https://github.com/kubevault/installer/commit/258774cc) Configure dependabot refresh schedule (#426)
- [3739ecf2](https://github.com/kubevault/installer/commit/3739ecf2) Update cve report (#425)
- [e2ead60b](https://github.com/kubevault/installer/commit/e2ead60b) Update cve report (#424)
- [7e7140a7](https://github.com/kubevault/installer/commit/7e7140a7) Update cve report (#423)
- [e58673ae](https://github.com/kubevault/installer/commit/e58673ae) Update cve report (#422)
- [45567a96](https://github.com/kubevault/installer/commit/45567a96) Update cve report (#421)
- [da65b37e](https://github.com/kubevault/installer/commit/da65b37e) Update cve report (#420)
- [9eee7121](https://github.com/kubevault/installer/commit/9eee7121) Update cve report (#419)
- [001cb34c](https://github.com/kubevault/installer/commit/001cb34c) Update cve report (#418)
- [0e330ac2](https://github.com/kubevault/installer/commit/0e330ac2) Update cve report (#417)
- [4022b1f1](https://github.com/kubevault/installer/commit/4022b1f1) Update cve report (#416)
- [072f7834](https://github.com/kubevault/installer/commit/072f7834) Update cve report (#415)
- [41bc426a](https://github.com/kubevault/installer/commit/41bc426a) Update cve report (#414)
- [3e830d6a](https://github.com/kubevault/installer/commit/3e830d6a) Update cve report (#413)
- [46f7442d](https://github.com/kubevault/installer/commit/46f7442d) Update cve report (#412)
- [8c32d070](https://github.com/kubevault/installer/commit/8c32d070) Update cve report (#411)
- [26480eea](https://github.com/kubevault/installer/commit/26480eea) Update cve report (#410)



## [kubevault/operator](https://github.com/kubevault/operator)

### [v0.25.0](https://github.com/kubevault/operator/releases/tag/v0.25.0)

- [fc6afdce](https://github.com/kubevault/operator/commit/fc6afdce9) Merge commit 'f63c2452e679487c11e6addd77f07e973c0aca5a' into release-0.25
- [f63c2452](https://github.com/kubevault/operator/commit/f63c2452e) Prepare for release v0.25.0 (#228)
- [d1338690](https://github.com/kubevault/operator/commit/d1338690b) Refactor Vault namespace/relay code and fix relay-image & migration correctness bugs (#224)
- [4402f8d6](https://github.com/kubevault/operator/commit/4402f8d6e) Support Sigilr as an OpenBao-derivative distribution (#227)
- [2fc4269c](https://github.com/kubevault/operator/commit/2fc4269ce) Resolve VaultRelay's container image from the hub AppBinding's spec.version (#225)
- [0e659db1](https://github.com/kubevault/operator/commit/0e659db1a) Ensure Namespace Slice CRD, Add ClusterID to Webhook (#223)
- [5b9c0ed9](https://github.com/kubevault/operator/commit/5b9c0ed90) Merge commit '2fcebcc9b329cba16769d030cd5e5bc666b35b75' into release-0.25
- [2fcebcc9](https://github.com/kubevault/operator/commit/2fcebcc9b) Prepare for release v0.25.0-rc.2 (#222)
- [6a93d458](https://github.com/kubevault/operator/commit/6a93d458d) Automatic OpenBao namespaces for KubeDB Platform tenants (#213)
- [1706fa88](https://github.com/kubevault/operator/commit/1706fa88e) design: tenant isolation on the OCM hub-spoke model (#214)
- [fe678460](https://github.com/kubevault/operator/commit/fe6784605) Pin client traffic to the Vault primary (spec.exposePrimary) (#220)
- [f76a2164](https://github.com/kubevault/operator/commit/f76a21646) Modernize golangci-lint config (#218)
- [400d7789](https://github.com/kubevault/operator/commit/400d7789f) Rename VaultAgent to VaultRelay; agentbackend -> relaybackend (agent->relay) (#217)
- [e3abfcbf](https://github.com/kubevault/operator/commit/e3abfcbf3) Remove nul file (#212)
- [11af4877](https://github.com/kubevault/operator/commit/11af48777) design: automatic OpenBao namespaces for KubeDB Platform tenants (#210)
- [b937d1d6](https://github.com/kubevault/operator/commit/b937d1d66) Fix errcheck lint: check fmt.Fprintf return in vault_test (#211)
- [8d8afcde](https://github.com/kubevault/operator/commit/8d8afcdeb) VaultAgent: hub-spoke deployment across OCM-managed clusters (#191)
- [8e259c10](https://github.com/kubevault/operator/commit/8e259c108) Fix e2e workflows (#190)
- [a09d5f2e](https://github.com/kubevault/operator/commit/a09d5f2ea) feat: Migrate operator to kubebuilder/controller-runtime style (#175)
- [83b7774f](https://github.com/kubevault/operator/commit/83b7774f0) Add CLAUDE.md pointing to AGENTS.md
- [3d90749d](https://github.com/kubevault/operator/commit/3d90749d1) Fix release tracker workflow
- [9418ad4b](https://github.com/kubevault/operator/commit/9418ad4b7) Fix release tracker workflow
- [593dd1ee](https://github.com/kubevault/operator/commit/593dd1ee6) Prepare for release v0.25.0-rc.1 (#174)
- [ecd00441](https://github.com/kubevault/operator/commit/ecd004413) Add AGENTS.md (#173)
- [eb6ba157](https://github.com/kubevault/operator/commit/eb6ba1574) Pin git user to 1gtm in update-crds/update-docs workflows (#172)
- [29418235](https://github.com/kubevault/operator/commit/294182359) Pin docker/login-action to v4.1.0 (#171)
- [545de7ef](https://github.com/kubevault/operator/commit/545de7ef7) Use docker/login-action instead of docker login command (#170)
- [402f21bc](https://github.com/kubevault/operator/commit/402f21bc4) Harden CI workflows (#169)
- [5a1522f2](https://github.com/kubevault/operator/commit/5a1522f26) Prepare for release v0.25.0-rc.0 (#168)
- [ae69e2d1](https://github.com/kubevault/operator/commit/ae69e2d1c) Cleanup cves (#167)
- [7c4b7a9a](https://github.com/kubevault/operator/commit/7c4b7a9a0) Harden CI workflows (#166)
- [eea0fa0c](https://github.com/kubevault/operator/commit/eea0fa0c7) Configure dependabot refresh schedule (#164)



## [kubevault/unsealer](https://github.com/kubevault/unsealer)

### [v0.25.0](https://github.com/kubevault/unsealer/releases/tag/v0.25.0)

- [bcce8ad3](https://github.com/kubevault/unsealer/commit/bcce8ad3) Cleanup deps
- [0f6a92ca](https://github.com/kubevault/unsealer/commit/0f6a92ca) Modernize golangci-lint config (#162)
- [d2717085](https://github.com/kubevault/unsealer/commit/d2717085) add namespace policy (#161)
- [a34e6891](https://github.com/kubevault/unsealer/commit/a34e6891) Grant policy-controller the relay/* backend paths (hub-spoke placement) (#160)
- [a7fe869a](https://github.com/kubevault/unsealer/commit/a7fe869a) Add CLAUDE.md pointing to AGENTS.md
- [3b5bf9c4](https://github.com/kubevault/unsealer/commit/3b5bf9c4) Fix release tracker workflow
- [aa6de2ae](https://github.com/kubevault/unsealer/commit/aa6de2ae) Add AGENTS.md (#158)
- [a7096a5a](https://github.com/kubevault/unsealer/commit/a7096a5a) Pin git user to 1gtm in update-crds/update-docs workflows (#157)
- [c92ce049](https://github.com/kubevault/unsealer/commit/c92ce049) Pin docker/login-action to v4.1.0 (#156)
- [97e4d369](https://github.com/kubevault/unsealer/commit/97e4d369) Add 1gtm-app[bot] to kodiak auto_approve_usernames (#155)
- [8dda08bf](https://github.com/kubevault/unsealer/commit/8dda08bf) Use docker/login-action instead of docker login command (#154)
- [50ce5684](https://github.com/kubevault/unsealer/commit/50ce5684) Normalize Prepare git user, fetch-depth, drop permission-issues (#153)
- [87ceec47](https://github.com/kubevault/unsealer/commit/87ceec47) Cleanup cves (#152)
- [5989ad61](https://github.com/kubevault/unsealer/commit/5989ad61) Use GitHub App token for release tracker comments (#151)
- [4f672af9](https://github.com/kubevault/unsealer/commit/4f672af9) Harden CI workflows (#150)
- [e283b0fb](https://github.com/kubevault/unsealer/commit/e283b0fb) Configure dependabot refresh schedule (#148)




