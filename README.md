# Cockroach Tools [![CircleCI](https://circleci.com/gh/utilitywarehouse/docker-cockroach-tools.svg?style=svg&circle-token=d220b3fb97a38ee8321d564e9e4443dd858650c5)]
A collection of tools to make easier running secure Cockroach DB cluster
on Kubernetes. The project contains following two utilities:
* [Request Certificates](#request-certificates)
* [Health Checker](#health-checker)

## Request Certificates
Request Certs is a utility to get ssl certificates for cockroach nodes and clients from cfssl CA.
It is inspired by a similar tool that uses Kubernetes CA from
[cockroach](https://github.com/cockroachdb/k8s/tree/master/request-cert).

### Kubernetes manifest examples
The Docker image for this tool is intended to be used in initContainers.
Below are example configurations of initContainers for a database client 
and a database node itself.

#### Client
Note that `USER` should be the name of the database user that 
this certificate is going to be used for. The user name will be
used as a common name in the certificate.

```
  initContainers:
  - name: init-certs
    image: registry.uw.systems/cockroach-cfssl-certs:latest
    imagePullPolicy: Always
    command: ["request-certs"]
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
#### Node
```
  initContainers:
  - name: init-certs
    image: registry.uw.systems/cockroach-cfssl-certs:latest
    imagePullPolicy: Always
    command:
    - "sh"
    - "-c"
    - >
      request-certs
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

## Health Checker
Health checker is a small http service that exposes a health endpoint
and is intended to be run as a sidecar of a Cockroach node.
When called the endpoint checks expiry of the certificate and then 
forwards the request to the provided health endpoint of a Cockroach instance.
