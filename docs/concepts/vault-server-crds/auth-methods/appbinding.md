---
title: AppBinding | KubeVault Concepts
menu:
  docs_{{ .version }}:
    identifier: appbinding-auth-methods
    name: AppBinding
    parent: auth-methods-vault-server-crds
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# AppBinding CRD

AppBinding CRD provides a way to specify connection information, credential and parameters that are necessary for communicating with app/service. In Vault operator, AppBinding used to communicate with vault, database, etc. This also provides flexibility to use external app/service.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  ...
```

## AppBinding Spec

AppBinding `spec` contains connection, credential and parameters information.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: vault-app
  namespace: demo
spec:
  type: vault
  secret:
    name: vault-token
  clientConfig:
    service:
      name: vault
      port: 8200
      scheme: HTTPS
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9ERXlNamN3TkRVNU1qVmFGdzB5T0RFeU1qUXdORFU1TWpWYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMVhid2wyQ1NNc2VQTU5RRzhMd3dUVWVOCkI1T05oSTlDNzFtdUoyZEZjTTlUc1VDQnlRRk1weUc5dWFvV3J1ZDhtSWpwMVl3MmVIUW5udmoybXRmWGcrWFcKSThCYkJUaUFKMWxMMFE5MlV0a1BLczlXWEt6dTN0SjJUR1hRRDhhbHZhZ0JrR1ViOFJYaUNqK2pnc1p6TDRvQQpNRWszSU9jS0xnMm9ldFZNQ0hwNktpWTBnQkZiUWdJZ1A1TnFwbksrbU02ZTc1ZW5hWEdBK2V1d09FT0YwV0Z2CmxGQmgzSEY5QlBGdTJKbkZQUlpHVDJKajBRR1FNeUxodEY5Tk1pZTdkQnhiTWhRVitvUXp2d1EvaXk1Q2pndXQKeDc3d29HQ2JtM0o4cXRybUg2Tjl6Tlc3WlR0YTdLd05PTmFoSUFEMSsrQm5rc3JvYi9BYWRKT0tMN2dLYndJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFXeWFsdUt3Wk1COWtZOEU5WkdJcHJkZFQyZnFTd0lEOUQzVjN5anBlaDVCOUZHN1UKSS8wNmpuRVcyaWpESXNHNkFDZzJKOXdyaSttZ2VIa2Y2WFFNWjFwZHRWeDZLVWplWTVnZStzcGdCRTEyR2NPdwpxMUhJb0NrekVBMk5HOGRNRGM4dkQ5WHBQWGwxdW5veWN4Y0VMeFVRSC9PRlc4eHJxNU9vcXVYUkxMMnlKcXNGCmlvM2lJV3EvU09Yajc4MVp6MW5BV1JSNCtSYW1KWjlOcUNjb1Z3b3R6VzI1UWJKWWJ3QzJOSkNENEFwOUtXUjUKU2w2blk3NVMybEdSRENsQkNnN2VRdzcwU25seW5mb3RaTUpKdmFzbStrOWR3U0xtSDh2RDNMMGNGOW5SOENTSgpiTjBiZzczeVlWRHgyY3JRYk0zcko4dUJnY3BsWlRpUy91SXJ2QT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
```

AppBinding Spec has following fields:

### spec.type

`spec.type` is an optional field that specifies the type of app.

```yaml
spec:
  type: vault
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
It has following fields:

- `clientConfig.url` : `Optional`. Specifies the location of the app, in standard URL form (`[scheme://]host:port/path`).

- `clientConfig.service`: `Optional`. Specifies the reference of the Kubernetes service for this app. It has following fields:
    - `service.scheme` : `Optional`. Specifies which scheme to use, for example: http, https. If specified, then it will applied as prefix in this format: `scheme://`. If not specified, then nothing will be prefixed.
    - `service.name` : `Required`. Specifies the name of the service. This `service.name` and AppBinding's namespace will used to create app DNS.
    - `service.port` : `Required`. Specifies the port that will be exposed by this app.
    - `service.path` : `Optional`. Specifies the URL path which will be sent in any request to this service.
    - `service.query` : `Optional`. Specifies the encoded query string, without '?' which will be sent in any request to this service.

> Note: Either `clientConfig.url` or `clientConfig.service` must be specified.

- `clientConfig.caBundle`: `Optional`. Specifies the PEM encoded CA bundle which will be used to validate the serving certificate of this app.

- `clientConfig.insecureSkipTLSVerify`: `Optional`. To skip TLS certificate verification when communicating with this app. This is strongly discouraged.  You should use the `clientConfig.caBundle` instead.

### spec.secret

`spec.secret` is an optional field that specifies the name of secret containing credential associated with AppBinding. It must be in AppBinding's namespace.

```yaml
spec:
  secret:
    name: vault-token
```

### spec.parameters

`spce.parameters` is an optional field that specifies the list of parameter to be used to connect to the app. The Parameters field is NOT secret or secured in any way and should NEVER be used to hold sensitive information.

```yaml
spec:
  parameters:
    authPath: "kubernetes"
    policyControllerRole: "demo"
    foo: "bar"
```

### spec.secretTransforms

`spec.secretTransforms` is an optional field that contains the list of transformations that should be applied to the credentials associated with the AppBinding before they are inserted into the Secret. For example, the credential secret specified in `spec.secret.name` has the key `USERNAME`, but the consumer requires the username to be exposed under the key `DB_USER` instead. To have the Vault operator transform the secret, the following secret transformation must be specified in `spec.secretTransforms`.

```yaml
spec:
  secretTransforms:
    - renameKey:
        from: USERNAME
        to: DB_USER
```

It has following fields:

- `secretTransforms[].renameKey`: `Optional`. Specifies a transform that renames a credentials secret entry's key. It has following fields:
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

