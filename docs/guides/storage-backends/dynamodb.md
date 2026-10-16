---
title: Vault Server DynamoDB Backend
menu:
  docs_{{ .version }}:
    identifier: vault-server-dynamodb-backend
    name: DynamoDB
    parent: storage-backends-guides
    weight: 80
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeVault? Please start [here](/docs/concepts/README.md).

# Deploy VaultServer with DynamoDB Backend

You can easily deploy and manage [OpenBao](https://openbao.org/) in the Kubernetes cluster using KubeVault operator. In this tutorial, we are going to deploy Vault on the Kubernetes cluster using KubeVault operator, with [Amazon DynamoDB](https://aws.amazon.com/dynamodb/) as the storage backend.

![Vault Server](/docs/images/guides/vault-server/vault_server_guide.svg)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy VaultServer with DynamoDB Backend

### Deploy Sample DynamoDB

To keep this tutorial self-contained, we are going to use [DynamoDB Local](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.html) instead of the AWS service. Vault creates the table if it does not exist.

Below is the YAML of the `Service` and `Deployment` that we are going to create for this tutorial:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: dynamodb
  namespace: demo
  labels:
    app: dynamodb
spec:
  ports:
  - port: 8000
    name: http
  selector:
    app: dynamodb
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dynamodb
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: dynamodb
  template:
    metadata:
      labels:
        app: dynamodb
    spec:
      containers:
      - name: dynamodb
        image: amazon/dynamodb-local:latest
        imagePullPolicy: IfNotPresent
        command:
        - "-jar"
        - "DynamoDBLocal.jar"
        - "-sharedDb"
        - "-inMemory"
        ports:
        - containerPort: 8000
          name: http
        resources:
          limits:
            cpu: "500m"
            memory: "512Mi"
          requests:
            cpu: "100m"
            memory: "256Mi"
```

Create the above resources,

```bash
$ kubectl apply -f dynamodb-local.yaml
service/dynamodb created
deployment.apps/dynamodb created

$ kubectl wait -n demo --for=condition=Ready pod -l app=dynamodb --timeout=5m
pod/dynamodb-5c8d7f6b9-p4x7q condition met
```

**Create Credential Secret:**

The credential secret holds the AWS credentials in the keys `access_key`, `secret_key` and `session_token`. DynamoDB Local accepts any values, so we use dummy ones.

```bash
$ kubectl create secret generic dynamodb-cred -n demo \
    --from-literal=access_key=test \
    --from-literal=secret_key=test \
    --from-literal=session_token=test
secret/dynamodb-cred created
```

> To use the AWS service, remove `endpoint` from the `VaultServer`, set `region` to your region and put your own credentials in the secret. `session_token` is the session token of temporary credentials. The `VaultServer` reads this key from the secret, so it must be present in it. The IAM identity needs permission to read and write the table, and to create it if the table does not exist yet.

## Deploy Vault using KubeVault

We're going to use Kubernetes secret to store the unseal-keys & root-token. A sample `VaultServer` with DynamoDB backend manifest file may look like this:

```yaml
apiVersion: kubevault.com/v1alpha2
kind: VaultServer
metadata:
  name: vault
  namespace: demo
spec:
  version: sigilr-2.6.1.2
  replicas: 2
  allowedSecretEngines:
    namespaces:
      from: All
  backend:
    dynamodb:
      endpoint: "http://dynamodb.demo.svc:8000"
      region: "us-east-1"
      table: "vault-data"
      credentialSecretRef:
        name: dynamodb-cred
      haEnabled: true
      readCapacity: 5
      writeCapacity: 5
  unsealer:
    secretShares: 5
    secretThreshold: 3
    mode:
      kubernetesSecret:
        secretName: vault-keys
  terminationPolicy: WipeOut
```

`readCapacity` and `writeCapacity` are only used when Vault creates the table. With `haEnabled: true`, the `VaultServer` can run with more than one replica: one pod is the active node and the others are standby.

Now, let's deploy the `VaultServer`:

```bash
$ kubectl apply -f vaultserver.yaml
vaultserver.kubevault.com/vault created
```

`KubeVault` operator will create a `AppBinding` CRD on `VaultServer` deployment, which contains the necessary information
to take backup of the Vault instances. It'll have the same name & be created on the same namespace as the `Vault`.
Read more about `AppBinding` [here](/docs/concepts/vault-server-crds/appbinding.md).

```bash
$ kubectl get appbinding -n demo vault -oyaml
```

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
    backend: dynamodb
    backupTokenSecretRef:
      name: vault-backup-token
    deploymentMode: Local
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
              value: k8s.62e86b47-6e7c-4849-89fb-88062a22f451.demo.vault
        restoreTask:
          name: vault-restore-1.10.3
          params:
            - name: keyPrefix
              value: k8s.62e86b47-6e7c-4849-89fb-88062a22f451.demo.vault
    unsealer:
      mode:
        kubernetesSecret:
          secretName: vault-keys
      secretShares: 5
      secretThreshold: 3
    vaultRole: vault-policy-controller
  type: VaultServer
  version: sigilr-2.6.1.2
```

Now, let's wait until all the vault pods come up & `VaultServer` phase becomes `Ready`.

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=vault
NAME      READY   STATUS    RESTARTS   AGE
vault-0   2/2     Running   0          2m8s
vault-1   2/2     Running   0          91s
```

```bash
$ kubectl get vaultserver -n demo
NAME    REPLICAS   VERSION          STATUS   AGE
vault   2          sigilr-2.6.1.2   Ready    2m50s
```

At this stage, we've successfully deployed `Vault` using `KubeVault` operator with DynamoDB as the storage backend.

Let's write some data in a `KV` secret engine. Let's export the necessary environment variables & port-forward from `vault` service
or exec into the vault pod in order to interact with it.

```bash
$ export VAULT_TOKEN=$(kubectl vault root-token get vaultserver vault -n demo --value-only)
$ export VAULT_ADDR='http://127.0.0.1:8200'
$ kubectl port-forward -n demo svc/vault 8200
```
Now check whether Vault server can be accessed (run this in another terminal):

```bash
$ vault status
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    5
Threshold       3
Version         2.6.1
Storage Type    dynamodb
Cluster Name    vault-cluster-3f8a1c92
Cluster ID      56c3a681-f799-2310-9813-d67877d25450
HA Enabled      true
HA Cluster      https://vault.demo.svc:8201
HA Mode         active
Active Since    2026-10-12T11:41:37.13040169Z
```

We can see the currently enabled list of secret engines.

```bash
$ vault secrets list
Path                                       Type         Accessor              Description
----                                       ----         --------              -----------
cubbyhole/                                 cubbyhole    cubbyhole_bb7c56f9    per-token private secret storage
identity/                                  identity     identity_fa8431fa     identity store
k8s.kubevault.com.kv.demo.vault-health/    kv           kv_5129d194           n/a
sys/                                       system       system_c7e0879a       system endpoints used for control, policy and debugging
```

Let's enable a `KV` type secret engine:

```bash
$ vault secrets enable kv
Success! Enabled the kv secrets engine at: kv/
```

Write some dummy data in the secret engine path:

```bash
$ vault kv put kv/name name=appscode
Success! Data written to: kv/name
```

Verify data written in `KV` secret engine:

```bash
$ vault kv get kv/name
==== Data ====
Key     Value
---     -----
name    appscode
```

## Cleanup

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete vaultserver -n demo vault
$ kubectl delete secret -n demo dynamodb-cred
$ kubectl delete -f dynamodb-local.yaml
$ kubectl delete ns demo
```

Delete the `VaultServer` before DynamoDB. With `terminationPolicy: WipeOut`, the operator connects to DynamoDB to remove the items Vault stored there, so the endpoint has to be reachable until the `VaultServer` is gone.

To see all the fields you can set under `spec.backend.dynamodb`, please check the [DynamoDB storage backend](/docs/concepts/vault-server-crds/storage/dynamodb.md) concept.

For more details on how to interact with the vault server, please check the [Vault Server guide](/docs/guides/vault-server/vault-server.md).
