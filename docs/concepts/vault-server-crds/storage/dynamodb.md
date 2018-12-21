# DynamoDB

In DynamoDB storage backend, data will be stored in [DynamoDB](https://aws.amazon.com/dynamodb/). Vault documentation for DynamoDB storage can be found in [here](https://www.vaultproject.io/docs/configuration/storage/dynamodb.html).


```yaml
apiVersion: kubevault.com/v1alpha1
kind: VaultServer
metadata:
  name: vault-with-dynamoDB
  namespace: demo
spec:
  nodes: 1
  version: "0.11.1"
  backend:
    dynamoDB:
      table: "my-vault-table"
      region: "us-west-1"
      readCapacity: 5
      writeCapacity: 5
```

## spec.backend.dynamoDB

To use dynamoDB as backend storage in Vault specify `spec.backend.dynamoDB` in [VaultServer](/docs/concepts/vault-server-crds/vaultserver.md) CRD.

```yaml
spec:
  backend:
    dynamoDB:
      table: <table_name>
      region: <region>
      endPoint: <endpoint>
      haEnabled: <true/false>
      readCapacity: <read_capacity>
      writeCapacity: <write_capacity>
      credentialSecret: <secret_name>
      sessionTokenSecret: <secret_name>
      maxParallel: <max_parallel>
```

`spec.backend.dynamoDB` has following fields:

#### dynamoDB.table

`dynamoDB.table` is a required field that specifies the name of the DynamoDB table. If the specified table does not exist, then Vault will create it during initialization. If it is not initialized, then Vault will set value `vault-dynamodb-backend`.

```yaml
spec:
  backend:
    dynamoDB:
      table: "my-vault-table"
```

#### dynamoDB.endPoint

`dynamoDB.endPoint` is an optional field that specifies an alternative, AWS compatible, DynamoDB endpoint.

```yaml
spec:
  backend:
    dynamoDB:
      endPoint: "endpoint.com"
```

#### dynamoDB.region

`dynamoDB.region` is an optional field that specifies the AWS region. If this field is not specified, then Vault will set value `us-east-1`.

```yaml
spec:
  backend:
    dynamoDB:
      region: "us-east-1"
```

#### dynamoDB.credentialSecret

`dynamoDB.credentialSecret` is an optional field that specifies the secret name containing AWS access key and AWS secret key. The secret contains following keys:
  
- `access_key`
- `secret_key`

Leaving the `access_key` and `secret_key` fields empty will cause Vault to attempt to retrieve credentials from the AWS metadata service.

```yaml
spec:
  backend:
    dynamoDB:
      credentialSecret: "aws-credential"
```

#### dynamoDB.sessionTokenSecret

`dynamoDB.sessionTokenSecret` is an optional field that specifies the secret name containing AWS session token. The secret contains following key:
  
- `session_token`

```yaml
spec:
  backend:
    dynamoDB:
      sessionTokenSecret: "aws-session-token"
```

#### dynamoDB.maxParallel

`maxParallel` is an optional field that specifies the maximum number of parallel operations to take place. This field accepts integer value. If this field is not specified, then Vault will set value `128`.

```yaml
spec:
  backend:
    dynamoDB:
      maxParallel: 124
```

#### dynamoDB.readCapacity

`dynamoDB.readCapacity` is an optional field that specifies the maximum number of reads consumed per second on the table. If it is not specifies, then Vault will set value `5`.

```yaml
spec:
  backend:
    dynamoDB:
      readCapacity: 10
```

#### dynamoDB.writeCapacity

`dynamoDB.writeCapacity` is an optional field that specifies the maximum number of writes performed per second on the table. If it is not specifies, then Vault will set value `5`.

```yaml
spec:
  backend:
    dynamoDB:
      writeCapacity: 10
```

#### dynamoDB.haEnabled

`dynamoDB.haEnabled` is an optional field that specifies whether this backend should be used to run Vault in high availability mode. This field accepts boolean value. Default value is `false`.

```yaml
spec:
  backend:
    dynamoDB:
      haEnabled: true
```
