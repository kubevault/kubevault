---
title: OpenStack Swift | Vault Server Storage
menu:
  docs_0.1.0:
    identifier: swift-storage
    name: OpenStack Swift
    parent: storage-vault-server-crds
    weight: 50
menu_name: docs_0.1.0
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Swift

In Swift storage backend, data will be stored in [OpenStack Swift Container](http://docs.openstack.org/developer/swift/). Vault documentation for Swift storage can be found in [here](https://www.vaultproject.io/docs/configuration/storage/swift.html).

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault-with-swift
  namespace: demo
spec:
  nodes: 1
  version: "0.11.1"
  backend:
    swift:
      authUrl: "https://auth.cloud.ovh.net/v2.0/"
      container: "my-vault-container"
      credentialSecret: "os-credential"
      region: "BHS1"
      tenant: "123456789999"
```

## spec.backend.swift

To use Swift as backend storage in Vault specify `spec.backend.swift` in [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    swift:
      authUrl: <auth_url>
      container: <container_name>
      credentialSecret: <secret_name>
      region: <region_name>
      tenant: <tenant_name>
      tenantID: <tenant_id>
      domain: <domain>
      projectDomain: <project_domain>
      storageUrl: <storage_url>
      authTokenSecret: <secret_name>
```

`spec.backend.swift` has following fields:

#### swift.authUrl

`swift.authUrl` is a required field that specifies the OpenStack authentication endpoint.

```yaml
spec:
  backend:
    swift:
      authUrl: "https://auth.cloud.ovh.net/v2.0/"
```

#### swift.container

`swift.container` is a required field that specifies the name of the Swift container.

```yaml
spec:
  backend:
    swift:
      container: "my-vault-container"
```

#### swift.credentialSecret

`swift.credentialSecret` is a required field that specifies the name of the secret containing the OpenStack account/username and password. The secret contains the following keys:

- `username`
- `password`

```yaml
spec:
  backend:
    swift:
      credentialSecret: "os-credential"
```

#### swift.tenant

`swift.tenant` is an optional field that specifies the name of the tenant. If it is not specifies, then Vault will set value to the default tenant of the username.

```yaml
spec:
  backend:
    swift:
      tenant: "123456789"
```

#### swift.region

`swift.region` is an optional field that specifies the name of the region.

```yaml
spec:
  backend:
    swift:
      region: "BHS1"
```

#### swift.tenantID

`swift.tenantID` is an optional field that specifies the id of the tenant.

```yaml
spec:
  backend:
    swift:
      tenantID: "11111111"
```

#### swift.domain

`swift.domain` is an optional field that specifies the name of the user domain.

```yaml
spec:
  backend:
    swift:
      domain: "my-domain"
```


#### swift.projectDomain

`swift.domain` is an optional field that specifies the name of the project's domain.

```yaml
spec:
  backend:
    swift:
      projectDomain: "my-project-domain"
```

#### swift.trustID

`swift.trustID` is an optional field that specifies the id of the trust.

```yaml
spec:
  backend:
    swift:
      trustID: "trust-id"
```

#### swift.storageUrl

`swift.storageUrl` is an optional field that specifies the storage URL from alternate authentication.

```yaml
spec:
  backend:
    swift:
      storageUrl: "storage.com"
```

#### swift.authTokenSecret

`swift.authTokenSecret` is an optional field that specifies the name of the secret containing auth token from alternate authentication.

```yaml
spec:
  backend:
    swift:
      authTokenSecret: "auth-token-secret"
```

#### swift.maxParallel

`maxParallel` is an optional field that specifies the maximum number of parallel operations to take place. This field accepts integer value. If this field is not specified, then Vault will set value `128`.

```yaml
spec:
  backend:
    swift:
      maxParallel: 124
```
