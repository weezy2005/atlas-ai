# syntax=docker/dockerfile:1.7
# Build context: repository root.
FROM golang:1.26-alpine AS build
ENV CGO_ENABLED=0 GOWORK=off
WORKDIR /src
COPY libs/go-common/go.mod libs/go-common/go.sum libs/go-common/
COPY services/gateway/go.mod services/gateway/go.sum services/gateway/
RUN --mount=type=cache,target=/go/pkg/mod cd services/gateway && go mod download
COPY libs/go-common libs/go-common
COPY services/gateway services/gateway
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build     cd services/gateway && go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/gateway

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
USER nonroot:nonroot
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 CMD ["/app", "-healthcheck"]
ENTRYPOINT ["/app"]
