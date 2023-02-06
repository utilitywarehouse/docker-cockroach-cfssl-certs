# cockroach-cfssl-certs 
Utility to get TLS certificates for cockroach nodes and clients from cfssl CA.
It is inspired by a similar tool that uses Kubernetes CA from
[cockroach](https://github.com/cockroachdb/k8s/tree/master/request-cert).

## Kubernetes manifest examples
The Docker image for this tool is intended to be used in initContainers.
Below are example configurations of initContainers for a database client 
and a database node itself.

### Client
Note that `USER` should be the name of the database user that 
this certificate is going to be used for. The user name will be
used as a common name in the certificate.

```
  initContainers:
  - name: init-certs
    image: registry.uw.systems/cockroach-cfssl-certs:initial
    imagePullPolicy: Always
    command: ["cockroach-certs"]
    env:
    - name: CERTIFICATE_TYPE
      value: "client"
    - name: USER
      value: "root"
    - name: CERTS_DIR
      value: "/cockroach-certs"
    - name: CA_PROFILE
      value: "client"
    - name: CA_ADDRESS
      valueFrom:
        configMapKeyRef:
          name: config
          key: ca.endpoint
    - name: CA_AUTH_KEY
      valueFrom:
        secretKeyRef:
          key: auth.key
          name: ca-auth-key
    volumeMounts:
    - name: client-certs
      mountPath: /cockroach-certs
```
### Node
```
  initContainers:
  - name: init-certs
    image: registry.uw.systems/cockroach-cfssl-certs:initial
    imagePullPolicy: Always
    command:
    - "sh"
    - "-c"
    - >
      cockroach-certs
      --host=localhost
      --host=127.0.0.1
      --host=$(hostname -f)
      --host=$(hostname -f|cut -f 1-2 -d '.')
      --host=cockroachdb-public
      --host=cockroachdb-public.$(hostname -f|cut -f 3- -d '.')
    env:
    - name: CERTIFICATE_TYPE
      value: "node"
    - name: CERTS_DIR
      value: "/cockroach-certs"
    - name: CA_PROFILE
      value: "client-server"
    - name: CA_ADDRESS
      valueFrom:
        configMapKeyRef:
          name: config
          key: ca.endpoint
    - name: CA_AUTH_KEY
      valueFrom:
        secretKeyRef:
          key: auth.key
          name: ca-auth-key
    volumeMounts:
    - name: certs
      mountPath: /cockroach-certs
```
