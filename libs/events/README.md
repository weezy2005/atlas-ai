# libs/events

Versioned event contracts shared by the Go and Python services (ADR-0005).

- `topics.txt`: every domain topic, one per line. The Kafka init job creates each topic plus `<topic>.retry` and `<topic>.dlq`.
- `schemas/envelope.v1.json`: the envelope every message uses (`event_id`, `event_type`, `occurred_at`, `user_id`, `trace_context`, `payload`).
- `schemas/<topic>.json`: payload schema for each topic.

Money and quantities are decimal strings, so values never pass through a float.
A breaking change means a new topic version (`foo.v2`); additive optional fields keep the same version.
