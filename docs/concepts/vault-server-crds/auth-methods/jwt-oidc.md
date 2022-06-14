---
title: Connect to Vault using JWT/OIDC Auth Method
menu:
  docs_{{ .version }}:
    identifier: jwt-oidc-auth-methods
    name: JWT/OIDC
    parent: auth-methods-vault-server-crds
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

The `JWT` auth method can be used to authenticate with Vault using `OIDC` or by providing a `JWT`.

The `OIDC` method allows authentication via a configured `OIDC Provider` using the user's web browser. This method may be initiated from the `Vault UI` or the command line. Alternatively, a `JWT` can be provided directly. 

## Enable & Configure JWT/OIDC Auth method

While deploying the `VaultServer` it's possible to define the list of auth methods users want to enable with it. 

A `VaultServer` `.spec.authMethods` section may look like this:

```yaml
spec:
  authMethods:
    - type: jwt
      path: jwt
      jwtConfig:
        ...
    - type: oidc
      path: oidc
      oidcConfig:
        ...

```

* `.spec.authMethods.type` is a required field, the type of authentication method we want to enable.
* `.spec.authMethods.path` is a required field, the path where we want to enable this authentication method.
* `.spec.authMethods.jwtConfig / .spec.authMethods.oidcConfig` contains various configuration for this authentication method. Details about configuration `parameters` can be found here: [JWT/OIDC Configuration](https://www.vaultproject.io/api-docs/auth/jwt#configure).

After an authentication method is successfully enabled, `KubeVault` operator will configure it with the provided configuration.

After successfully enabling & configuring authentication methods, a VaultServer `.status.authMethodStatus` may look like this:
```yaml
status:
  authMethodStatus:
  - path: jwt
    status: EnableSucceeded
    type: jwt
  - path: kubernetes
    status: EnableSucceeded
    type: kubernetes

```

So, this is how `JWT/OIDC` authentication method could be enabled & configured with `KubeVault`. For a step-by-step guide see [this](/docs/guides/vault-server/auth-method.md).