FROM golang:1.18 as builder

WORKDIR /app

RUN apt-get update \
    && apt-get install \
        ca-certificates \
    && update-ca-certificates

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build main.go

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/main /bin/duckdns-operator

USER nonroot:nonroot

CMD ["/bin/duckdns-operator"]
