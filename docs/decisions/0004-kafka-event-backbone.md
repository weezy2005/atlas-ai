# ADR-0004: Kafka as the event backbone

Status: Accepted (2026-10-22)

## Context
State changes (sync results, portfolio updates) fan out to several consumers that must stay decoupled and be able to replay.

## Decision
Kafka in KRaft mode (MSK in prod). Topics are keyed by user_id for per-user ordering, and every topic has a .retry and .dlq companion.

## Consequences
Heavier to run than SQS or Redis Streams, but gives replay, consumer groups and ordered fan-out.
