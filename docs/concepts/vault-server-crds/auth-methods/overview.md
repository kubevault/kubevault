---
title: Managing Externally Provisioned Vault Servers
menu:
  docs_{{ .version }}:
    identifier: overview-auth-methods
    name: External Vault
    parent: auth-methods-vault-server-crds
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Managing Externally Provisioned Vault Servers

The KubeVault operator can manage policies and secret engines of Vault servers which are not provisioned by the KubeVault operator. These Vault servers can be running outside a Kubernetes cluster or running inside a Kubernetes cluster but provisioned using a Helm chart.

The KubeVault operator uses an [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md) to connect to an externally provisioned Vault server. Following authentication methods are currently supported by the KubeVault operator:

- [AWS IAM Auth Method](/docs/concepts/vault-server-crds/auth-methods/aws-iam.md)
- [Kubernetes Auth Method](/docs/concepts/vault-server-crds/auth-methods/kubernetes.md)
- [TLS Certificates Auth Method](/docs/concepts/vault-server-crds/auth-methods/tls.md)
- [Token Auth Method](/docs/concepts/vault-server-crds/auth-methods/token.md)
- [Userpass Auth Method](/docs/concepts/vault-server-crds/auth-methods/userpass.md)
- [GCP IAM Auth Method](/docs/concepts/vault-server-crds/auth-methods/gcp-iam.md)
- [Azure Auth Method](/docs/concepts/vault-server-crds/auth-methods/azure.md)
