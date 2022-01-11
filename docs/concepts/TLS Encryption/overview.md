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

```yaml
spec:
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: vault-issuer
    certificates:
    - alias: server
      secretName: vautl-server-certs
      subject:
        organizations:
        - kubevault
    - alias: client
      secretName: vault-client-certs
      subject:
        organizations:
        - kubevault

```

The `spec.tls` contains the following fields:

- `tls.issuerRef` - is an `optional` field that references to the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for Elasticsearch. If the `issuerRef` is not specified, the operator creates a self-signed CA and also creates necessary certificate (valid: 365 days) secrets using that CA.
  - `apiGroup` - is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
  - `kind` - is the type of resource that is being referenced. The supported values are `Issuer` and `ClusterIssuer`.
  - `name` - is the name of the resource ( `Issuer` or `ClusterIssuer` ) that is being referenced.

- `tls.certificates` - is an `optional` field that specifies a list of certificate configurations used to configure the  certificates. It has the following fields:
  - `alias` - represents the identifier of the certificate. It has the following possible value:
    - `server` - is used for the server certificate configuration.
    - `client` - is used for the client certificate configuration.
    - `storage` - is used for the storage certificate configuration.
    - `ca` - is used for the ca certificate configuration.

  - `secretName` - ( `string` | `"<vault-name>-alias-certs"` ) - specifies the k8s secret name that holds the certificates.

  - `subject` - specifies an `X.509` distinguished name (DN). It has the following configurable fields:
    - `organizations` ( `[]string` | `nil` ) - is a list of organization names.
    - `organizationalUnits` ( `[]string` | `nil` ) - is a list of organization unit names.
    - `countries` ( `[]string` | `nil` ) -  is a list of country names (ie. Country Codes).
    - `localities` ( `[]string` | `nil` ) - is a list of locality names.
    - `provinces` ( `[]string` | `nil` ) - is a list of province names.
    - `streetAddresses` ( `[]string` | `nil` ) - is a list of street addresses.
    - `postalCodes` ( `[]string` | `nil` ) - is a list of postal codes.
    - `serialNumber` ( `string` | `""` ) is a serial number.

    For more details, visit [here](https://golang.org/pkg/crypto/x509/pkix/#Name).

  - `duration` ( `string` | `""` ) - is the period during which the certificate is valid. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as `"300m"`, `"1.5h"` or `"20h45m"`. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
  - `renewBefore` ( `string` | `""` ) - is a specifiable time before expiration duration.
  - `dnsNames` ( `[]string` | `nil` ) - is a list of subject alt names.
  - `ipAddresses` ( `[]string` | `nil` ) - is a list of IP addresses.
  - `uris` ( `[]string` | `nil` ) - is a list of URI Subject Alternative Names.
  - `emailAddresses` ( `[]string` | `nil` ) - is a list of email Subject Alternative Names.

  
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