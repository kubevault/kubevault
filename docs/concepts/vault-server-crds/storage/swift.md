---
title: OpenStack Swift | Vault Server Storage
menu:
  docs_{{ .version }}:
    identifier: swift-storage
    name: OpenStack Swift
    parent: storage-vault-server-crds
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Swift

In Swift storage backend, Vault data will be stored in [OpenStack Swift Container](http://docs.openstack.org/developer/swift/). Vault documentation for Swift storage can be found in [here](https://www.vaultproject.io/docs/configuration/storage/swift.html).

```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault-with-swift
  namespace: demo
spec:
  replicas: 1
  version: "1.2.0"
  backend:
    swift:
      authURL: "https://auth.cloud.ovh.net/v2.0/"
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
      authURL: <auth_url>
      container: <container_name>
      credentialSecret: <secret_name>
      region: <region_name>
      tenant: <tenant_name>
      tenantID: <tenant_id>
      domain: <domain>
      projectDomain: <project_domain>
      storageURL: <storage_url>
      authTokenSecret: <secret_name>
```

Here, we are going to describe the various attributes of the `spec.backend.swift` field.

### swift.authURL

`swift.authURL` is a required field that specifies the OpenStack authentication endpoint.

```yaml
spec:
  backend:
    swift:
      authURL: "https://auth.cloud.ovh.net/v2.0/"
```

### swift.container

`swift.container` is a required field that specifies the name of the Swift container.

```yaml
spec:
  backend:
    swift:
      container: "my-vault-container"
```

### swift.credentialSecret

`swift.credentialSecret` is a required field that specifies the name of the secret containing the OpenStack account/username and password. The secret contains the following keys:

- `username`
- `password`

```yaml
spec:
  backend:
    swift:
      credentialSecret: "os-credential"
```

### swift.tenant

`swift.tenant` is an optional field that specifies the name of the tenant. If it is not specified, then Vault will set the value to the default tenant of the username.

```yaml
spec:
  backend:
    swift:
      tenant: "123456789"
```

### swift.region

`swift.region` is an optional field that specifies the name of the region.

```yaml
spec:
  backend:
    swift:
      region: "BHS1"
```

### swift.tenantID

`swift.tenantID` is an optional field that specifies the id of the tenant.

```yaml
spec:
  backend:
    swift:
      tenantID: "11111111"
```

### swift.domain

`swift.domain` is an optional field that specifies the name of the user domain.

```yaml
spec:
  backend:
    swift:
      domain: "my-domain"
```

### swift.projectDomain

`swift.domain` is an optional field that specifies the name of the project's domain.

```yaml
spec:
  backend:
    swift:
      projectDomain: "my-project-domain"
```

### swift.trustID

`swift.trustID` is an optional field that specifies the id of the trust.

```yaml
spec:
  backend:
    swift:
      trustID: "trust-id"
```

### swift.storageURL

`swift.storageURL` is an optional field that specifies the storage URL from alternate authentication.

```yaml
spec:
  backend:
    swift:
      storageURL: "storage.com"
```

### swift.authTokenSecret

`swift.authTokenSecret` is an optional field that specifies the name of the secret containing auth token from alternate authentication.

```yaml
spec:
  backend:
    swift:
      authTokenSecret: "auth-token-secret"
```

### swift.maxParallel

`maxParallel` is an optional field that specifies the maximum number of parallel operations to take place. This field accepts integer value. If this field is not specified, then Vault will set value to `128`.

```yaml
spec:
  backend:
    swift:
      maxParallel: 124
```
