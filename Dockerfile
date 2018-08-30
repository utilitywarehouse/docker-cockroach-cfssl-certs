# -
# Build workspace
# -
FROM golang:1.10 AS compile
ARG SERVICE

RUN apt-get update -y && \
    apt-get install --no-install-recommends -y -q build-essential ca-certificates

WORKDIR /go/src/github.com/utilitywarehouse/$SERVICE
ADD . .
RUN make install-packages
RUN make static

# -
# Runtime
# -
FROM scratch AS runtime
ARG SERVICE

COPY --from=compile /go/src/github.com/utilitywarehouse/$SERVICE/$SERVICE /bin/request-certs
COPY --from=compile /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["request-certs"]
CMD ["--help"]
