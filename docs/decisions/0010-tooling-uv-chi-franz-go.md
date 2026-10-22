# ADR-0010: uv, chi and franz-go

Status: Accepted (2026-10-22)

## Context
The build prompt offered alternatives.

## Decision
uv (a fast workspace with one lockfile across Python services) over Poetry; chi (stdlib net/http compatible, so otelhttp and middleware work unchanged) over gin; franz-go (pure Go, no cgo, so static distroless images) over confluent-kafka-go.

## Consequences
franz-go has a smaller community than librdkafka bindings.
