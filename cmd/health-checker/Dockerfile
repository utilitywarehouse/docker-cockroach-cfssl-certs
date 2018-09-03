# -
# Build workspace
# -
FROM golang:1.10 AS compile

RUN apt-get update -y && \
    apt-get install --no-install-recommends -y -q build-essential ca-certificates

WORKDIR /go/src/github.com/utilitywarehouse/docker-cockroach-tools
ADD . .

ENV COMMAND=health-checker
RUN make install-packages
RUN make static

# -
# Runtime
# -
FROM alpine:3.8 AS runtime

COPY --from=compile /go/src/github.com/utilitywarehouse/docker-cockroach-tools/health-checker /bin/health-checker
COPY --from=compile /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["health-checker"]
CMD ["--help"]
