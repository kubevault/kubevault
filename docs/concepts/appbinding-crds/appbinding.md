# AppBinding CRD

AppBinding CRD provides a way to specify connection information, credential and parameters that are necessary for communicate will app/service. In Vault operator, AppBinding used to communicate with vault, database, etc. This also provides flexibility to use external app/service. 

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
      scheme: http
    insecureSkipTLSVerify: true
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
      scheme: http
    insecureSkipTLSVerify: true
```
It has following fields:

- `clientConfig.url` : `Optional`. Specifies the location of the app, in standard URL form (`[scheme://]host:port/path`).

- `clientConfig.service`: `Optional`. Specifies the reference of the kubernetes service for this app. It has following fields:
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

