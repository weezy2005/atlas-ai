# ADR-0006: Idempotent consumers via processed_events

Status: Accepted (2026-10-22)

## Context
Kafka delivers at-least-once, so redelivery after a rebalance or crash is normal.

## Decision
Each consumer writes the event_id into its own processed_events table in the same transaction as its side effect, and skips event_ids it has already seen.

## Consequences
One extra insert per event; the table needs periodic pruning past the retention window.
