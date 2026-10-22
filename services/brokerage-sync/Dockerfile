# syntax=docker/dockerfile:1.7
# Build context: repository root.
FROM golang:1.26-alpine AS build
ENV CGO_ENABLED=0 GOWORK=off
WORKDIR /src
COPY libs/go-common/go.mod libs/go-common/go.sum libs/go-common/
COPY services/brokerage-sync/go.mod services/brokerage-sync/go.sum services/brokerage-sync/
# extra_ca: see EXTRA_CA_FILE in .env.example.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=secret,id=extra_ca,required=false \
    cat /etc/ssl/certs/ca-certificates.crt /run/secrets/extra_ca > /tmp/ca.pem 2>/dev/null \
      || cp /etc/ssl/certs/ca-certificates.crt /tmp/ca.pem; \
    cd services/brokerage-sync && SSL_CERT_FILE=/tmp/ca.pem go mod download
COPY libs/go-common libs/go-common
COPY services/brokerage-sync services/brokerage-sync
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    cd services/brokerage-sync && go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/brokerage-sync

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
USER nonroot:nonroot
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 CMD ["/app", "-healthcheck"]
ENTRYPOINT ["/app"]
