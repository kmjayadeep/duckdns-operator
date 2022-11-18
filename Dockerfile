FROM golang:1.18 as builder

WORKDIR /app

RUN apt-get update \
    && apt-get install \
        ca-certificates \
    && update-ca-certificates

COPY . .
RUN go build main.go

FROM gcr.io/distroless/static

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/main /bin/duckdns-operator

# Run as UID for nobody since k8s pod securityContext runAsNonRoot can't resolve the user ID:
# https://github.com/kubernetes/kubernetes/issues/40958
USER 65534

ENTRYPOINT ["/bin/duckdns-operator"]
