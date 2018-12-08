---
title: Changelog | Vault operator
description: Changelog
menu:
  product_vault-operator_0.1.0:
    identifier: changelog-vault
    name: Changelog
    parent: welcome
    weight: 10
product_name: vault-operator
menu_name: product_vault-operator_0.1.0
section_menu_id: welcome
url: /products/vault-operator/0.1.0/welcome/changelog/
aliases:
  - /products/vault-operator/0.1.0/CHANGELOG/
---

# Change Log

## [Unreleased](https://github.com/kubevault/operator/tree/HEAD)

**Merged pull requests:**

- Update chart [\#106](https://github.com/kubevault/operator/pull/106) ([tamalsaha](https://github.com/tamalsaha))
- Add validation webhook [\#105](https://github.com/kubevault/operator/pull/105) ([nightfury1204](https://github.com/nightfury1204))
- Add vault server version crd [\#104](https://github.com/kubevault/operator/pull/104) ([nightfury1204](https://github.com/nightfury1204))
-  Add init container for vault configuration [\#103](https://github.com/kubevault/operator/pull/103) ([nightfury1204](https://github.com/nightfury1204))
- Use kubernetes-1.11.3 [\#102](https://github.com/kubevault/operator/pull/102) ([tamalsaha](https://github.com/tamalsaha))
- Add ensure operation and cleaup unused code [\#101](https://github.com/kubevault/operator/pull/101) ([nightfury1204](https://github.com/nightfury1204))
- Generate crd yamls with fixed IntHash schema [\#100](https://github.com/kubevault/operator/pull/100) ([tamalsaha](https://github.com/tamalsaha))
- Use IntHash as status.observedGeneration [\#99](https://github.com/kubevault/operator/pull/99) ([tamalsaha](https://github.com/tamalsaha))
- Add support for swift storage [\#98](https://github.com/kubevault/operator/pull/98) ([nightfury1204](https://github.com/nightfury1204))
- fix github status [\#97](https://github.com/kubevault/operator/pull/97) ([tahsinrahman](https://github.com/tahsinrahman))
- update pipeline [\#96](https://github.com/kubevault/operator/pull/96) ([tahsinrahman](https://github.com/tahsinrahman))
- Add concourse configs [\#95](https://github.com/kubevault/operator/pull/95) ([tahsinrahman](https://github.com/tahsinrahman))
- Improve Helm chart options [\#94](https://github.com/kubevault/operator/pull/94) ([tamalsaha](https://github.com/tamalsaha))
- Enable status sub resource for crd yamls [\#92](https://github.com/kubevault/operator/pull/92) ([tamalsaha](https://github.com/tamalsaha))
- Retry UpdateStatus calls [\#91](https://github.com/kubevault/operator/pull/91) ([tamalsaha](https://github.com/tamalsaha))
- Move crds to api folder [\#90](https://github.com/kubevault/operator/pull/90) ([tamalsaha](https://github.com/tamalsaha))
- Various improvements [\#89](https://github.com/kubevault/operator/pull/89) ([tamalsaha](https://github.com/tamalsaha))
- Correctly handle ignored openapi prefixes [\#87](https://github.com/kubevault/operator/pull/87) ([tamalsaha](https://github.com/tamalsaha))
- Add option for user provided TLS [\#86](https://github.com/kubevault/operator/pull/86) ([nightfury1204](https://github.com/nightfury1204))
- Set generated binary name to vault-operator [\#85](https://github.com/kubevault/operator/pull/85) ([tamalsaha](https://github.com/tamalsaha))
- Don't add admission/v1beta1 group as a prioritized version [\#84](https://github.com/kubevault/operator/pull/84) ([tamalsaha](https://github.com/tamalsaha))
- Use version and additional columns for crds [\#83](https://github.com/kubevault/operator/pull/83) ([tamalsaha](https://github.com/tamalsaha))
- Add support for dynamoDB [\#82](https://github.com/kubevault/operator/pull/82) ([nightfury1204](https://github.com/nightfury1204))
- Enable status subresource for crds [\#81](https://github.com/kubevault/operator/pull/81) ([tamalsaha](https://github.com/tamalsaha))
- Format shell script [\#80](https://github.com/kubevault/operator/pull/80) ([tamalsaha](https://github.com/tamalsaha))
- Update client-go to v8.0.0 [\#79](https://github.com/kubevault/operator/pull/79) ([tamalsaha](https://github.com/tamalsaha))
- Add Filesystem support [\#78](https://github.com/kubevault/operator/pull/78) ([nightfury1204](https://github.com/nightfury1204))
- Add support for MySQL [\#77](https://github.com/kubevault/operator/pull/77) ([nightfury1204](https://github.com/nightfury1204))
- Add support for postgresSQL [\#76](https://github.com/kubevault/operator/pull/76) ([nightfury1204](https://github.com/nightfury1204))
- Support for azure storage backend and azure key vault unsealer [\#71](https://github.com/kubevault/operator/pull/71) ([nightfury1204](https://github.com/nightfury1204))
- Support for aws s3 storage and awsKmsSsm unsealer [\#70](https://github.com/kubevault/operator/pull/70) ([nightfury1204](https://github.com/nightfury1204))
- Fix installer [\#69](https://github.com/kubevault/operator/pull/69) ([tamalsaha](https://github.com/tamalsaha))
- Rename org to kubevault from kube-vault [\#68](https://github.com/kubevault/operator/pull/68) ([tamalsaha](https://github.com/tamalsaha))
- Apply validation rules to vault server names [\#67](https://github.com/kubevault/operator/pull/67) ([tamalsaha](https://github.com/tamalsaha))
- Move openapi-spec to api package [\#66](https://github.com/kubevault/operator/pull/66) ([tamalsaha](https://github.com/tamalsaha))
- Add support for gcs backend, google kms gcs unsealer [\#65](https://github.com/kubevault/operator/pull/65) ([nightfury1204](https://github.com/nightfury1204))
- Add unsealer Spec [\#64](https://github.com/kubevault/operator/pull/64) ([nightfury1204](https://github.com/nightfury1204))
- Fix build [\#63](https://github.com/kubevault/operator/pull/63) ([tamalsaha](https://github.com/tamalsaha))
- Fix badges [\#62](https://github.com/kubevault/operator/pull/62) ([tamalsaha](https://github.com/tamalsaha))
- Update package path to github.com/kube-vault/operator [\#61](https://github.com/kubevault/operator/pull/61) ([tamalsaha](https://github.com/tamalsaha))
- Don't panic if admission options is nil [\#59](https://github.com/kubevault/operator/pull/59) ([tamalsaha](https://github.com/tamalsaha))
- Update rbac permissions [\#56](https://github.com/kubevault/operator/pull/56) ([tamalsaha](https://github.com/tamalsaha))
- Add Update\*\*\*Status helpers [\#55](https://github.com/kubevault/operator/pull/55) ([tamalsaha](https://github.com/tamalsaha))
- Update client-go to v7.0.0 [\#54](https://github.com/kubevault/operator/pull/54) ([tamalsaha](https://github.com/tamalsaha))
- Add vault controller [\#53](https://github.com/kubevault/operator/pull/53) ([nightfury1204](https://github.com/nightfury1204))
- Update workload library [\#51](https://github.com/kubevault/operator/pull/51) ([tamalsaha](https://github.com/tamalsaha))
- Improve installer [\#50](https://github.com/kubevault/operator/pull/50) ([tamalsaha](https://github.com/tamalsaha))
- Fix installer [\#49](https://github.com/kubevault/operator/pull/49) ([tamalsaha](https://github.com/tamalsaha))
- Update workload client [\#48](https://github.com/kubevault/operator/pull/48) ([tamalsaha](https://github.com/tamalsaha))
- Update workload client [\#47](https://github.com/kubevault/operator/pull/47) ([tamalsaha](https://github.com/tamalsaha))
- Update workload client [\#46](https://github.com/kubevault/operator/pull/46) ([tamalsaha](https://github.com/tamalsaha))
- Add vault api [\#45](https://github.com/kubevault/operator/pull/45) ([nightfury1204](https://github.com/nightfury1204))
- Update workload api [\#44](https://github.com/kubevault/operator/pull/44) ([tamalsaha](https://github.com/tamalsaha))
- Switch to mutating webhook [\#43](https://github.com/kubevault/operator/pull/43) ([tamalsaha](https://github.com/tamalsaha))
- Update vault deployment [\#42](https://github.com/kubevault/operator/pull/42) ([tamalsaha](https://github.com/tamalsaha))
- Rename Secret to VaultSecret [\#41](https://github.com/kubevault/operator/pull/41) ([tamalsaha](https://github.com/tamalsaha))
- Add installer scripts [\#40](https://github.com/kubevault/operator/pull/40) ([tamalsaha](https://github.com/tamalsaha))
- Compress vault operator binary [\#39](https://github.com/kubevault/operator/pull/39) ([tamalsaha](https://github.com/tamalsaha))
- Rename --analytics flag [\#38](https://github.com/kubevault/operator/pull/38) ([tamalsaha](https://github.com/tamalsaha))
- Update docker build script [\#37](https://github.com/kubevault/operator/pull/37) ([tamalsaha](https://github.com/tamalsaha))
- Update .gitignore [\#36](https://github.com/kubevault/operator/pull/36) ([tamalsaha](https://github.com/tamalsaha))
- Rename api types [\#35](https://github.com/kubevault/operator/pull/35) ([tamalsaha](https://github.com/tamalsaha))
- Clone apis [\#34](https://github.com/kubevault/operator/pull/34) ([tamalsaha](https://github.com/tamalsaha))
- Update links [\#33](https://github.com/kubevault/operator/pull/33) ([tamalsaha](https://github.com/tamalsaha))
- Update chart [\#32](https://github.com/kubevault/operator/pull/32) ([tamalsaha](https://github.com/tamalsaha))
- Update package path [\#31](https://github.com/kubevault/operator/pull/31) ([tamalsaha](https://github.com/tamalsaha))
- Add travis yaml [\#29](https://github.com/kubevault/operator/pull/29) ([tahsinrahman](https://github.com/tahsinrahman))
- Use shared informer factory [\#28](https://github.com/kubevault/operator/pull/28) ([tamalsaha](https://github.com/tamalsaha))
- Update client-go to v6.0.0 [\#27](https://github.com/kubevault/operator/pull/27) ([tamalsaha](https://github.com/tamalsaha))
- Add front matter for steward cli [\#25](https://github.com/kubevault/operator/pull/25) ([tamalsaha](https://github.com/tamalsaha))
- Use client-go 5.x [\#23](https://github.com/kubevault/operator/pull/23) ([tamalsaha](https://github.com/tamalsaha))
- Add chart [\#21](https://github.com/kubevault/operator/pull/21) ([tamalsaha](https://github.com/tamalsaha))
- Renew token periodically [\#19](https://github.com/kubevault/operator/pull/19) ([tamalsaha](https://github.com/tamalsaha))
- Add initializer for Controllers [\#14](https://github.com/kubevault/operator/pull/14) ([tamalsaha](https://github.com/tamalsaha))
- Revise initializers [\#12](https://github.com/kubevault/operator/pull/12) ([tamalsaha](https://github.com/tamalsaha))
- Fix initializers [\#10](https://github.com/kubevault/operator/pull/10) ([tamalsaha](https://github.com/tamalsaha))
- Implement pod initializer and finalizer [\#9](https://github.com/kubevault/operator/pull/9) ([tamalsaha](https://github.com/tamalsaha))
- Create vault secret for service account. [\#8](https://github.com/kubevault/operator/pull/8) ([tamalsaha](https://github.com/tamalsaha))
- Sync secrets to vault [\#7](https://github.com/kubevault/operator/pull/7) ([tamalsaha](https://github.com/tamalsaha))
- Init vault [\#6](https://github.com/kubevault/operator/pull/6) ([tamalsaha](https://github.com/tamalsaha))
- Initial skeleton [\#1](https://github.com/kubevault/operator/pull/1) ([tamalsaha](https://github.com/tamalsaha))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*