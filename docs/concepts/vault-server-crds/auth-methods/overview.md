---
title: Vault Server Authentication Methods
menu:
  docs_0.2.0:
    identifier: overview-auth-methods
    name: AppBinding
    parent: auth-methods-vault-server-crds
    weight: 5
menu_name: docs_0.2.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Vault Server Authentication Methods

In Vault operator, usually Vault connection information are handled by [AppBinding](/docs/concepts/vault-server-crds/auth-methods/appbinding.md). Following authentication methods are currently supported by Vault operator using AppBinding:

- [AWS IAM Auth Method](/docs/concepts/vault-server-crds/auth-methods/aws-iam.md)
- [Kubernetes Auth Method](/docs/concepts/vault-server-crds/auth-methods/kubernetes.md)
- [TLS Certificates Auth Method](/docs/concepts/vault-server-crds/auth-methods/tls.md)
- [Token Auth Method](/docs/concepts/vault-server-crds/auth-methods/token.md)
- [Userpass Auth Method](/docs/concepts/vault-server-crds/auth-methods/userpass.md)
- [GCP IAM Auth Method](/docs/concepts/vault-server-crds/auth-methods/gcp-iam.md)
- [Azure Auth Method](/docs/concepts/vault-server-crds/auth-methods/azure.md)
