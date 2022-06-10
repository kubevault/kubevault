---
title: Changelog | KubeVault
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubevault-v2022.06.16
    name: Changelog-v2022.06.16
    parent: welcome
    weight: 20220616
product_name: kubevault
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2022.06.16/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2022.06.16/
---

# KubeVault v2022.06.16 (2022-06-10)


## [kubevault/apimachinery](https://github.com/kubevault/apimachinery)

### [v0.8.0](https://github.com/kubevault/apimachinery/releases/tag/v0.8.0)

- [6cbcd6ed](https://github.com/kubevault/apimachinery/commit/6cbcd6ed) Update to k8s 1.24 toolchain (#60)
- [2970666d](https://github.com/kubevault/apimachinery/commit/2970666d) Add support for JWT/OIDC auth method (#58)
- [885da664](https://github.com/kubevault/apimachinery/commit/885da664) Test against Kubernetes 1.24.0 (#59)
- [da0750a7](https://github.com/kubevault/apimachinery/commit/da0750a7) v1alpha2 api conversion (#52)
- [716f6fa7](https://github.com/kubevault/apimachinery/commit/716f6fa7) Use Go 1.18 (#55)
- [831710c9](https://github.com/kubevault/apimachinery/commit/831710c9) Use Go 1.18 (#53)
- [78081ee6](https://github.com/kubevault/apimachinery/commit/78081ee6) Add kubevault.com/v1alpha2 api (#50)
- [33b160d7](https://github.com/kubevault/apimachinery/commit/33b160d7) make fmt (#48)
- [6c189049](https://github.com/kubevault/apimachinery/commit/6c189049) Add Kind() methods (#47)



## [kubevault/cli](https://github.com/kubevault/cli)

### [v0.8.0](https://github.com/kubevault/cli/releases/tag/v0.8.0)

- [e41268d7](https://github.com/kubevault/cli/commit/e41268d7) Prepare for release v0.8.0 (#155)
- [2b54317e](https://github.com/kubevault/cli/commit/2b54317e) Update to k8s 1.24 toolchain (#154)
- [b684e2b9](https://github.com/kubevault/cli/commit/b684e2b9) Add v1alpha2 changes (#153)
- [c338ec82](https://github.com/kubevault/cli/commit/c338ec82) Use Go 1.18 (#152)
- [60b63920](https://github.com/kubevault/cli/commit/60b63920) Use Go 1.18 (#150)
- [6b7ec5ef](https://github.com/kubevault/cli/commit/6b7ec5ef) make fmt (#149)
- [4e5ce4c9](https://github.com/kubevault/cli/commit/4e5ce4c9) Update UID generation for GenericResource (#148)



## [kubevault/installer](https://github.com/kubevault/installer)

### [v2022.06.16](https://github.com/kubevault/installer/releases/tag/v2022.06.16)

- [43feb59](https://github.com/kubevault/installer/commit/43feb59) Prepare for release v2022.06.16 (#171)
- [b2ac521](https://github.com/kubevault/installer/commit/b2ac521) Update registry templates to support custom default registry (ghcr.io) (#170)
- [0b0c419](https://github.com/kubevault/installer/commit/0b0c419) Don't set tag in values files
- [cd242d6](https://github.com/kubevault/installer/commit/cd242d6) Add support for vault:1.10.3 (#167)
- [530db29](https://github.com/kubevault/installer/commit/530db29) Add secrets-store-reader chart (#169)
- [98e19e5](https://github.com/kubevault/installer/commit/98e19e5) Test against Kubernetes 1.24.0 (#168)
- [f212b1b](https://github.com/kubevault/installer/commit/f212b1b) Get operator tag from .Chart.AppVersion (#166)
- [9db93b4](https://github.com/kubevault/installer/commit/9db93b4) Test against Kubernetes 1.24.0 (#165)
- [f07c0de](https://github.com/kubevault/installer/commit/f07c0de) Test operator monitoring (#164)
- [9f7011e](https://github.com/kubevault/installer/commit/9f7011e) Add apiservice get permission for crd conversion webhook config (#163)
- [1a01910](https://github.com/kubevault/installer/commit/1a01910) Change MY_POD_ env fix to POD_
- [ce24272](https://github.com/kubevault/installer/commit/ce24272) Clean up monitoring values from chart (#161)
- [3cf0088](https://github.com/kubevault/installer/commit/3cf0088) Remove unimplemented webhook config
- [5937bbc](https://github.com/kubevault/installer/commit/5937bbc) Remove validators.policy.kubevault.com apiservice
- [d3084e5](https://github.com/kubevault/installer/commit/d3084e5) Add webhook & kubevault-operator permission (#160)
- [19e0a9b](https://github.com/kubevault/installer/commit/19e0a9b) Split into operator and webhook chart (#158)
- [2b8b424](https://github.com/kubevault/installer/commit/2b8b424) Use Go 1.18 (#157)
- [0bc3634](https://github.com/kubevault/installer/commit/0bc3634) Use Go 1.18 (#156)
- [e18357e](https://github.com/kubevault/installer/commit/e18357e) make fmt (#154)
- [427e146](https://github.com/kubevault/installer/commit/427e146) Use webhooks suffix for webhook resources (#152)



## [kubevault/operator](https://github.com/kubevault/operator)

### [v0.8.0](https://github.com/kubevault/operator/releases/tag/v0.8.0)

- [06ad2e3a](https://github.com/kubevault/operator/commit/06ad2e3a) Prepare for release v0.8.0 (#63)
- [26197fe0](https://github.com/kubevault/operator/commit/26197fe0) Update to k8s 1.24 toolchain (#62)
- [d6f03739](https://github.com/kubevault/operator/commit/d6f03739) Add support for JWT/OIDC auth method, Fix Vault resources sync (#56)
- [58e1cf81](https://github.com/kubevault/operator/commit/58e1cf81) Update ci.yml
- [6cad6340](https://github.com/kubevault/operator/commit/6cad6340) Disable trivy scanner
- [8932662a](https://github.com/kubevault/operator/commit/8932662a) Use CI hosts with label ubuntu-latest
- [28c4829d](https://github.com/kubevault/operator/commit/28c4829d) Test against Kubernetes 1.24.0 (#61)
- [f195aaad](https://github.com/kubevault/operator/commit/f195aaad) Enable CI checks
- [2c14f4d8](https://github.com/kubevault/operator/commit/2c14f4d8) Run e2e tests (#60)
- [0361cecd](https://github.com/kubevault/operator/commit/0361cecd) Fix CI (#59)
- [f225c385](https://github.com/kubevault/operator/commit/f225c385) Introduce separate commands for operator and webhook (#58)
- [935b7ce5](https://github.com/kubevault/operator/commit/935b7ce5) Use sefl-hosted runner (#57)
- [c677d9ae](https://github.com/kubevault/operator/commit/c677d9ae) Use Go 1.18 (#54)
- [90f0ab47](https://github.com/kubevault/operator/commit/90f0ab47) make fmt (#52)
- [17593e4c](https://github.com/kubevault/operator/commit/17593e4c) Use webhooks suffix for webhook resources (#51)
- [41bbd827](https://github.com/kubevault/operator/commit/41bbd827) Cancel concurrent CI runs for same pr/commit (#50)



## [kubevault/unsealer](https://github.com/kubevault/unsealer)

### [v0.8.0](https://github.com/kubevault/unsealer/releases/tag/v0.8.0)

- [2b05c29b](https://github.com/kubevault/unsealer/commit/2b05c29b) Update to k8s 1.24 toolchain (#120)
- [085df87a](https://github.com/kubevault/unsealer/commit/085df87a) Use Go 1.18 (#119)
- [f70eb944](https://github.com/kubevault/unsealer/commit/f70eb944) Use Go 1.18 (#118)
- [41d67ef4](https://github.com/kubevault/unsealer/commit/41d67ef4) make fmt (#117)
- [7ff0f673](https://github.com/kubevault/unsealer/commit/7ff0f673) Cancel concurrent CI runs for same pr/commit (#116)




