---
title: AppBinding | KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: appbinding-vault-server-crds
    name: AppBinding
    parent: vault-server-crds-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# AppBinding

## What is AppBinding

AppBinding CRD provides a way to specify connection information, credential, and parameters that are necessary for communicating with an app/service. In KubeVault operator, AppBinding used to communicate with externally provisioned Vault, database, etc.

An `AppBinding` is a Kubernetes `CustomResourceDefinition`(CRD) which points to an application using either its URL (usually for a non-Kubernetes resident service instance) or a Kubernetes service object (if self-hosted in a Kubernetes cluster), some optional parameters and a credential secret. To learn more about AppBinding and the problems it solves, please read this blog post: [The case for AppBinding](https://blog.byte.builders/post/the-case-for-appbinding).

If you deploy a Vault using [KubeVault](https://kubevault.com/docs/latest/welcome/), the `AppBinding` object will be created automatically for it. Otherwise, you have to create an `AppBinding` object manually pointing to your desired Vault.

KubeVault uses [Stash](https://stash.run/) to perform backup/recovery of Vault cluster. Stash needs to know how to connect with a target Vault and the credentials necessary to access it. This is done via an `AppBinding`.

## AppBinding CRD Specification

Like any official Kubernetes resource, an `AppBinding` has `TypeMeta`, `ObjectMeta` and `Spec` sections. However, unlike other Kubernetes resources, it does not have a `Status` section.

An `AppBinding` object created by `KubeVault` for a Vault is shown below,

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault
  namespace: demo
spec:
  appRef:
    apiGroup: kubevault.com
    kind: VaultServer
    name: vault
    namespace: demo
  clientConfig:
    service:
      name: vault
      port: 8200
      scheme: http
  parameters:
    apiVersion: config.kubevault.com/v1alpha1
    backend: raft
    backupTokenSecretRef:
      name: vault-backup-token
    kind: VaultServerConfiguration
    kubernetes:
      serviceAccountName: vault
      tokenReviewerServiceAccountName: vault-k8s-token-reviewer
      usePodServiceAccountForCSIDriver: true
    path: kubernetes
    stash:
      addon:
        backupTask:
          name: vault-backup-1.10.3
          params:
            - name: keyPrefix
              value: k8s.75cf1449-0675-4cce-9fbe-965441450355.demo.vault
        restoreTask:
          name: vault-restore-1.10.3
          params:
            - name: keyPrefix
              value: k8s.75cf1449-0675-4cce-9fbe-965441450355.demo.vault
    unsealer:
      mode:
        googleKmsGcs:
          bucket: vault-testing-keys
          credentialSecretRef:
            name: gcp-cred
          kmsCryptoKey: vault-testing-key
          kmsKeyRing: vault-testing
          kmsLocation: global
          kmsProject: appscode-testing
      secretShares: 5
      secretThreshold: 3
    vaultRole: vault-policy-controller
```

Here, we are going to describe the sections of an `AppBinding` crd.

Here, we are going to describe the sections of an `AppBinding` crd that are relevant to KubeVault.

### AppBinding `Spec`

An `AppBinding` object has the following fields in the `spec` section:

### spec.appRef

`spec.appRef` refers to the application for which this AppBinding was created. 

```yaml
appRef:
  apiGroup: kubevault.com
  kind: VaultServer
  name: vault
  namespace: demo
```

#### spec.parameters.backupTokenSecretRef

`spec.parameters.backupTokenSecretRef` specifies the name of the secret which contains the token with the permission to take backup snapshot & restore. This secret must be in the same namespace as the `AppBinding`.

This secret must contain the following keys:

|    Key     | Usage                                                   |
|:----------:|---------------------------------------------------------|
|  `token`   | token with permission to take backup snapshot & restore |

### spec.parameters.backend

`spec.parameters.backend` indicates the Storage Backend type Vault is using.

### spec.parameters.stash

`spec.parameters.stash` section contains the stash parameters for Backup & Restore tasks. `spec.parameters.stash.addon` contains the 
information about the `Task` for backup & restore. It also contains the `params` which indicates the `keyPrefix`
that is prepended with the name of vault `unseal-keys` & `root-token`, 
e.g. `k8s.kubevault.com.demo.vault-root-token`, `k8s.kubevault.com.demo.vault-root-token-unseal-key-0`, `k8s.kubevault.com.demo.vault-root-token-unseal-key-1`, etc.

```yaml
stash:
  addon:
    backupTask:
      name: vault-backup-1.10.3
      params:
        - name: keyPrefix
          value: k8s.kubevault.com.demo.vault
    restoreTask:
      name: vault-restore-1.10.3
      params:
        - name: keyPrefix
          value: k8s.kubevault.com.demo.vault
```

### spec.parameters.unsealer

`spec.parameters.unsealer` contains the Vault unsealer information, e.g. secret shares & threshold of unseal keys, the location of unseal-keys, etc.

```yaml
unsealer:
  mode:
    googleKmsGcs:
      bucket: vault-testing-keys
      credentialSecretRef:
        name: gcp-cred
      kmsCryptoKey: vault-testing-key
      kmsKeyRing: vault-testing
      kmsLocation: global
      kmsProject: appscode-testing
  secretShares: 5
  secretThreshold: 3
```

### spec.clientConfig

`spec.clientConfig` is a required field that specifies the information to make a connection with an app.

```yaml
spec:
  clientConfig:
    service:
      name: vault
      port: 8200
      scheme: HTTPS
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9ERXlNamN3TkRVNU1qVmFGdzB5T0RFeU1qUXdORFU1TWpWYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMVhid2wyQ1NNc2VQTU5RRzhMd3dUVWVOCkI1T05oSTlDNzFtdUoyZEZjTTlUc1VDQnlRRk1weUc5dWFvV3J1ZDhtSWpwMVl3MmVIUW5udmoybXRmWGcrWFcKSThCYkJUaUFKMWxMMFE5MlV0a1BLczlXWEt6dTN0SjJUR1hRRDhhbHZhZ0JrR1ViOFJYaUNqK2pnc1p6TDRvQQpNRWszSU9jS0xnMm9ldFZNQ0hwNktpWTBnQkZiUWdJZ1A1TnFwbksrbU02ZTc1ZW5hWEdBK2V1d09FT0YwV0Z2CmxGQmgzSEY5QlBGdTJKbkZQUlpHVDJKajBRR1FNeUxodEY5Tk1pZTdkQnhiTWhRVitvUXp2d1EvaXk1Q2pndXQKeDc3d29HQ2JtM0o4cXRybUg2Tjl6Tlc3WlR0YTdLd05PTmFoSUFEMSsrQm5rc3JvYi9BYWRKT0tMN2dLYndJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFXeWFsdUt3Wk1COWtZOEU5WkdJcHJkZFQyZnFTd0lEOUQzVjN5anBlaDVCOUZHN1UKSS8wNmpuRVcyaWpESXNHNkFDZzJKOXdyaSttZ2VIa2Y2WFFNWjFwZHRWeDZLVWplWTVnZStzcGdCRTEyR2NPdwpxMUhJb0NrekVBMk5HOGRNRGM4dkQ5WHBQWGwxdW5veWN4Y0VMeFVRSC9PRlc4eHJxNU9vcXVYUkxMMnlKcXNGCmlvM2lJV3EvU09Yajc4MVp6MW5BV1JSNCtSYW1KWjlOcUNjb1Z3b3R6VzI1UWJKWWJ3QzJOSkNENEFwOUtXUjUKU2w2blk3NVMybEdSRENsQkNnN2VRdzcwU25seW5mb3RaTUpKdmFzbStrOWR3U0xtSDh2RDNMMGNGOW5SOENTSgpiTjBiZzczeVlWRHgyY3JRYk0zcko4dUJnY3BsWlRpUy91SXJ2QT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
```

It has the following fields:

- `clientConfig.url` : `Optional`. Specifies the location of the app, in standard URL form (`[scheme://]host:port/path`).

- `clientConfig.service`: `Optional`. Specifies the reference of the Kubernetes service for this app. It has the following fields:

  - `service.scheme` : `Optional`. Specifies which scheme to use, for example, HTTP, https. If specified, then it will be applied as the prefix in this format: `scheme://`. If not specified, then nothing will be prefixed.
  - `service.name` : `Required`. Specifies the name of the service. This `service.name` and AppBinding's namespace will be used to create app DNS.
  - `service.port` : `Required`. Specifies the port that will be exposed by this app.
  - `service.path` : `Optional`. Specifies the URL path which will be sent in any request to this service.
  - `service.query` : `Optional`. Specifies the encoded query string, without '?' which will be sent in any request to this service.

> Note: Either `clientConfig.url` or `clientConfig.service` must be specified.

- `clientConfig.caBundle`: `Optional`. Specifies the PEM encoded CA bundle which will be used to validate the serving certificate of this app.

- `clientConfig.insecureSkipTLSVerify`: `Optional`. To skip TLS certificate verification when communicating with this app. This is strongly discouraged.  You should use the `clientConfig.caBundle` instead.

### spec.secret

`spec.secret` is an optional field that specifies the name of the secret containing credentials associated with AppBinding. It must be in AppBinding's namespace.

```yaml
spec:
  secret:
    name: vault-token
```

### spec.parameters

`spce.parameters` is an optional field that specifies the list of parameters to be used to connect to the app, backup, restore, etc. The Parameters field is *not* secret or secured in any way and should *never* be used to hold sensitive information.

```yaml
spec:
  parameters:
    apiVersion: config.kubevault.com/v1alpha1
    backend: raft
    backupTokenSecretRef:
      name: vault-backup-token
    kind: VaultServerConfiguration
    kubernetes:
      serviceAccountName: vault
      tokenReviewerServiceAccountName: vault-k8s-token-reviewer
      usePodServiceAccountForCSIDriver: true
    path: kubernetes
    stash:
      addon:
        backupTask:
          name: vault-backup-1.10.3
          params:
            - name: keyPrefix
              value: k8s.75cf1449-0675-4cce-9fbe-965441450355.demo.vault
        restoreTask:
          name: vault-restore-1.10.3
          params:
            - name: keyPrefix
              value: k8s.75cf1449-0675-4cce-9fbe-965441450355.demo.vault
    unsealer:
      mode:
        googleKmsGcs:
          bucket: vault-testing-keys
          credentialSecretRef:
            name: gcp-cred
          kmsCryptoKey: vault-testing-key
          kmsKeyRing: vault-testing
          kmsLocation: global
          kmsProject: appscode-testing
      secretShares: 5
      secretThreshold: 3
    vaultRole: vault-policy-controller
```

### spec.secretTransforms

`spec.secretTransforms` is an optional field that contains the list of transformations that should be applied to the credentials associated with the AppBinding before they are inserted into the Secret. For example, the credential secret specified in `spec.secret.name` has the key `USERNAME`, but the consumer requires the username to be exposed under the key `DB_USER` instead. To have the KubeVault operator transform the secret, the following secret transformation must be specified in `spec.secretTransforms`.

```yaml
spec:
  secretTransforms:
    - renameKey:
        from: USERNAME
        to: DB_USER
```

It has the following fields:

- `secretTransforms[].renameKey`: `Optional`. Specifies a transform that renames a credentials secret entry's key. It has the following fields:
  - `renameKey.from`: `Required`. Specifies the name of the key to rename.
  - `renameKey.to`: `Required`. Specifies the new name for the key.

- `secretTransforms[].addKey`: `Optional`. Specifies a transform that adds an additional key to the credentials secret.
  - `addKey.key`: `Required`. Specifies the name of the key to add.
  - `addKey.value`: `Required`. Specifies the value (possibly non-binary) to add to the secret under the specified key.
  - `addKey.stringValue`: `Required`. Specifies the string value to add to the secret under the specified key. If both `addKey.value` and `addKey.stringValue` are specified, then `addKey.value` is ignored and `addKey.stringValue` is stored.
  - `addKey.jsonPathExpression`: `Required`. Specifies the JSONPath expression, the result of which will be added to the Secret under the specified key. For example, given the following credentials: `{ "foo": { "bar": "foobar" } }` and the jsonPathExpression `{.foo.bar}`, the value `foobar` will be stored in the credentials secret under the specified key.

- `secretTransforms[].addKeysFrom`: `Optional`. Specifies a transform that merges all the entries of an existing secret into the credentials secret.
  - `addKeysFrom.secretRef.name`: `Optinal`. Specifies the name of the secret.
  - `addKeysFrom.secretRef.namespace`: `Optinal`. Specifies the namespace of the secret.

- `secretTransforms[].removeKey`: `Optional`. Specifies a transform that removes a credentials secret entry.
  - `removeKey.key`. `Required`. Specifies the key to remove from secret.

