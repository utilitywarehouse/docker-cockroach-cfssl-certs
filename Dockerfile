FROM golang:1.20-alpine AS compile

WORKDIR /build
COPY . .
RUN apk --no-cache add git \
      && go get -d -v ./... \
      && go generate \
      && CGO_ENABLED=0 go build -o=cockroach-certs .

FROM alpine:3.17 AS runtime

COPY --from=compile /build/cockroach-certs /bin/cockroach-certs
COPY --from=compile /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 8000

ENTRYPOINT ["cockroach-certs"]
CMD ["--help"]
