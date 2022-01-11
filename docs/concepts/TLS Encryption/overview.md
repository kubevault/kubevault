---
title: Overview | KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: vault-tls-overview
    name: Overview
    parent: vault-server-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Overview

**Prerequisite :** To configure TLS/SSL in `VaultServer`, `KubeVault` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

To issue a certificate, the following cr of `cert-manager` is used:

- `Issuer/ClusterIssuer`: Issuers and ClusterIssuers represent certificate authorities (CAs) that are able to generate signed certificates by honoring certificate signing requests. All cert-manager certificates require a referenced issuer that is in a ready condition to attempt to honor the request. You can learn more details [here](https://cert-manager.io/docs/concepts/issuer/).

- `Certificate`: `cert-manager` has the concept of Certificates that define the desired x509 certificate which will be renewed and kept up to date. You can learn more details [here](https://cert-manager.io/docs/concepts/certificate/).

**VaultServer CRD Specification:**

KubeValt uses the following cr fields to enable SSL/TLS encryption in `VaultServer`.

- `spec:`
  - `tls:`
    - `issuerRef`

  
## How TLS/SSL configures in VaultServer

The following figure shows how `KubeVault` is used to configure TLS/SSL in Postgres. Open the image in a new tab to see the enlarged version.

Deploying VaultServer with TLS/SSL configuration process consists of the following steps:

1. At first, a user creates an `Issuer/ClusterIssuer` cr.

2. Then the user creates a `VaultServer` cr.

3. `KubeVault` community operator watches for the `VaultServer` cr.

4. When it finds one, it creates `Secret`, `Service`, etc. for the `VaultServer`.

5. `KubeVault` operator watches for `VaultServer`(5c), `Issuer/ClusterIssuer`(5b), `Secret` and `Service`(5a).

6. When it finds all the resources(`VaultServer`, `Issuer/ClusterIssuer`, `Secret`, `Service`), it creates `Certificates` by using `tls.issuerRef` and `tls.certificates` field specification from `VaultServer` cr.

7. `cert-manager` watches for certificates.

8. When it finds one, it creates certificate secrets `cert-secrets`(server, client, exporter secrets, etc.) that hold the actual self-signed certificate.

9. `KubeVault` community operator watches for the Certificate secrets `tls-secrets`.

10. When it finds all the tls-secret, it creates a `StatefulSet` so that Postgres server is configured with TLS/SSL.

In the next doc, we are going to show a step by step guide on how to configure a `VaultServer` with TLS/SSL.