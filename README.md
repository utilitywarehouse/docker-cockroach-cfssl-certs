# cockroach-cfssl-certs
Utility to get ssl certificates for cockroach nodes and clients from cfssl CA.
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
    image: registry.uw.systems/cockroach-cfssl-certs:latest
    imagePullPolicy: Always
    args: []
    env:
    - name:  CERTIFICATE_TYPE
      value: "client"
    - name:  USER
      value: "root"
    - name:  CERTS_DIR
      value: "/cockroach-certs"
    - name:  CA_PROFILE
      value: "client"
    - name:  CA_ADDRESS
      value: "certificate-authority:8080"
    - name: CA_AUTH_KEY
      valueFrom:
        secretKeyRef:
          key: auth.key
          name: cfssl-auth-key
    volumeMounts:
    - name: client-certs
      mountPath: /cockroach-certs
```
### Node
```
  initContainers:
  - name: init-certs
    image: registry.uw.systems/cockroach-cfssl-certs:latest
    imagePullPolicy: Always
    args:
    - "--hosts"
    - "localhost,127.0.0.1,$(hostname -f),$(hostname -f|cut -f 1-2 -d '.'),cockroachdb-public,cockroachdb-public.$(hostname -f|cut -f 3- -d '.')"
    env:
    - name:  CERTIFICATE_TYPE
      value: "node"
    - name:  CERTS_DIR
      value: "/cockroach-certs"
    - name:  CA_PROFILE
      value: "client"
    - name:  CA_ADDRESS
      value: "certificate-authority:8080"
    - name: CA_AUTH_KEY
      valueFrom:
        secretKeyRef:
          key: auth.key
          name: cfssl-auth-key
    volumeMounts:
    - name: certs
      mountPath: /cockroach-certs
```