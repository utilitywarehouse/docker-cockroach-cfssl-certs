# -
# Build workspace
# -
FROM golang:1.11 AS compile

RUN apt-get update -y && \
    apt-get install --no-install-recommends -y -q build-essential ca-certificates

WORKDIR /build
ADD . .
RUN make install
RUN make static

# -
# Runtime
# -
FROM alpine:3.8 AS runtime

COPY --from=compile /build/cockroach-certs /bin/cockroach-certs
COPY --from=compile /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["cockroach-certs"]
CMD ["--help"]
