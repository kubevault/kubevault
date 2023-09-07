---
title: Changelog | KubeVault
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubevault-v2023.9.7
    name: Changelog-v2023.9.7
    parent: welcome
    weight: 20230907
product_name: kubevault
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2023.9.7/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2023.9.7/
---

# KubeVault v2023.9.7 (2023-09-07)


## [kubevault/apimachinery](https://github.com/kubevault/apimachinery)

### [v0.16.0](https://github.com/kubevault/apimachinery/releases/tag/v0.16.0)

- [349e26c3](https://github.com/kubevault/apimachinery/commit/349e26c3) Update deps (#88)
- [70fe1196](https://github.com/kubevault/apimachinery/commit/70fe1196) Update Object Ref of Role in SAR Sepc (#86)
- [61c7dec1](https://github.com/kubevault/apimachinery/commit/61c7dec1) Show secret name in SecretAccessRequest
- [fb4e4b83](https://github.com/kubevault/apimachinery/commit/fb4e4b83) Update conditions api (#85)
- [d45cf33d](https://github.com/kubevault/apimachinery/commit/d45cf33d) Update license verifier (#84)



## [kubevault/cli](https://github.com/kubevault/cli)

### [v0.16.0](https://github.com/kubevault/cli/releases/tag/v0.16.0)

- [d330e716](https://github.com/kubevault/cli/commit/d330e716) Prepare for release v0.16.0 (#180)
- [3a4b07e5](https://github.com/kubevault/cli/commit/3a4b07e5) Update deps (#179)
- [a507b3ee](https://github.com/kubevault/cli/commit/a507b3ee) Update deps
- [0ea76b7e](https://github.com/kubevault/cli/commit/0ea76b7e) Update license verifier (#177)



## [kubevault/installer](https://github.com/kubevault/installer)

### [v2023.9.7](https://github.com/kubevault/installer/releases/tag/v2023.9.7)

- [2794038](https://github.com/kubevault/installer/commit/2794038) Prepare for release v2023.9.7 (#213)
- [3bbace5](https://github.com/kubevault/installer/commit/3bbace5) Update deps (#212)
- [25f1bb8](https://github.com/kubevault/installer/commit/25f1bb8) Enable seccompProfile RuntimeDefault for CI (#211)
- [b3d9ee0](https://github.com/kubevault/installer/commit/b3d9ee0) Update clusterrole to list namespace (#210)
- [e63f017](https://github.com/kubevault/installer/commit/e63f017) Add Vault Version 1.13.3 (#208)
- [22e3460](https://github.com/kubevault/installer/commit/22e3460) Remove seccomp profile for charts (#209)
- [b0cb8d3](https://github.com/kubevault/installer/commit/b0cb8d3) Use helm repo version
- [57aad9b](https://github.com/kubevault/installer/commit/57aad9b) Don't mount license vol when both license and licenseSecretName is empty (#207)
- [9c73417](https://github.com/kubevault/installer/commit/9c73417) Add licenseSecretName values (#206)
- [5cccbed](https://github.com/kubevault/installer/commit/5cccbed) Switch to failurePolicy: Ignore by default for webhooks (#205)



## [kubevault/operator](https://github.com/kubevault/operator)

### [v0.16.0](https://github.com/kubevault/operator/releases/tag/v0.16.0)

- [221a32a0](https://github.com/kubevault/operator/commit/221a32a0) Prepare for release v0.16.0 (#108)
- [2809f196](https://github.com/kubevault/operator/commit/2809f196) Update deps (#107)
- [0ca991a4](https://github.com/kubevault/operator/commit/0ca991a4) Fix Vault CR Deletion when Backend DB deleted (#102)
- [addb895b](https://github.com/kubevault/operator/commit/addb895b) Configure seccomp (#106)
- [9e4bb1f6](https://github.com/kubevault/operator/commit/9e4bb1f6) Add cross namespace secret access request (#105)
- [5c549f84](https://github.com/kubevault/operator/commit/5c549f84) Update Conditions API (#104)
- [c3be22cd](https://github.com/kubevault/operator/commit/c3be22cd) Use updated conditions api
- [dc8ac8f8](https://github.com/kubevault/operator/commit/dc8ac8f8) Use restricted pod security label (#103)
- [2de8e148](https://github.com/kubevault/operator/commit/2de8e148) Update license verifier (#101)



## [kubevault/unsealer](https://github.com/kubevault/unsealer)

### [v0.16.0](https://github.com/kubevault/unsealer/releases/tag/v0.16.0)

- [04800df4](https://github.com/kubevault/unsealer/commit/04800df4) Update deps (#131)
- [f3e89c6f](https://github.com/kubevault/unsealer/commit/f3e89c6f) Update license verifier (#130)




