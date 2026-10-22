# ADR-0005: JSON Schema over Avro for events

Status: Accepted (2026-10-22)

## Context
Event schemas are shared between Go and Python and must be versioned.

## Decision
JSON Schema (draft 2020-12) in libs/events, registered in Confluent Schema Registry. The version is part of the topic name (e.g. portfolio.updated.v1); breaking changes create a new topic version.

## Consequences
Larger messages than Avro and weaker compatibility tooling. In exchange, payloads are human-readable and codegen is easy in both languages.
