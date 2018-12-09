
## How to install vault on kubernetes


### What is vault?

To get introduced on `vault` click [here](https://www.vaultproject.io/docs/index.html)

### Install vault on kubernetes

To install vault on kubernetes, create file `vault.yaml` and put the following data

```yaml
apiVersion: v1
kind: Service
metadata:
  name: vault
spec:
  ports:
  - name: http
    nodePort: 30001
    port: 8200
  selector:
    app: vault
  type: NodePort
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: vault
  labels:
    app: vault
spec:
  serviceName: "vault"
  selector:
    matchLabels:
      app: vault
  replicas: 1
  template:
    metadata:
      labels:
        app: vault
    spec:
      containers:
      - name: vault
        image: "vault:0.10.4"
        args:
        - "server"
        - "-dev"
        - "-dev-root-token-id=root"
        ports:
        - name: http
          containerPort: 8200
          protocol: "TCP"
        - name: server
          containerPort: 8201
          protocol: "TCP"
```

then run

```bash
$ kubectl apply -f vault.yaml
```

