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
* `.spec.authMethods.jwtConfig / .spec.authMethods.oidcConfig` contains various configuration for this authentication method. Some of the `paramerters` are listed here: 
  * `defaultLeaseTTL` - The default lease duration, specified as a string duration like "5s" or "30m".
  * `maxLeaseTTL` - The maximum lease duration, specified as a string duration like "5s" or "30m".
  * `pluginName` - The name of the plugin in the plugin catalog to use.
  * `auditNonHMACRequestKeys` - List of keys that will not be HMAC'd by audit devices in the request data object.
  * `auditNonHMACResponseKeys` - List of keys that will not be HMAC'd by audit devices in the response data object.
  * `listingVisibility` - Specifies whether to show this mount in the UI-specific listing endpoint.
  * `passthroughRequestHeaders` - List of headers to whitelist and pass from the request to the backend.
  * `credentialSecretRef` - K8s Secret reference containing credential related secrets.
  * `tlsSecretRef` - K8s Secret reference containing tls related secrets.
  * `oidcDiscoveryURL` - The OIDC Discovery URL, without any .well-known component (base path). Cannot be used with "jwks_url" or "jwt_validation_pubkeys".
  * `oidcClientID` - The OAuth Client ID from the provider for OIDC roles.
  * `oidcResponseMode` - The response mode to be used in the OAuth2 request. Allowed values are "query" and "form_post". Defaults to "query". If using Vault namespaces, and oidc_response_mode is "form_post", then "namespace_in_state" should be set to false.
  * `oidcResponseTypes` - (comma-separated string, or array of strings: <optional>) - The response types to request. Allowed values are "code" and "id_token". Defaults to "code". Note: "id_token" may only be used if "oidc_response_mode" is set to "form_post".
  * `defaultRole` - The default role to use if none is provided during login.
  * `providerConfig` - Configuration options for provider-specific handling. Providers with specific handling include: Azure, Google. The options are described in each provider's section in OIDC Provider Setup.
  * `jwksURL` - JWKS URL to use to authenticate signatures. Cannot be used with "oidc_discovery_url" or "jwt_validation_pubkeys".
  * `jwtValidationPubkeys` - (comma-separated string, or array of strings: <optional>). A list of PEM-encoded public keys to use to authenticate signatures locally. Cannot be used with "jwks_url" or "oidc_discovery_url".
  * `jwtSupportedAlgs` - (comma-separated string, or array of strings: <optional>) A list of supported signing algorithms. Defaults to [RS256] for OIDC roles. Defaults to all available algorithms for JWT roles.
  * `boundIssuer` - The value against which to match the iss claim in a JWT.

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

We can verify it using the `Vault CLI`:

```bash
$ vault auth list

Path           Type          Accessor                    Description
----           ----          --------                    -----------
jwt/           jwt           auth_jwt_ba23cc30           n/a
kubernetes/    kubernetes    auth_kubernetes_40fd86fd    n/a
token/         token         auth_token_950c8b80         token based credentials
```

So, this is how `JWT/OIDC` authentication method could be enabled & configured with `KubeVault`. 

> For a step-by-step guide on JWT/OIDC authentication method, see [this](/docs/guides/vault-server/auth-method.md).
