# ADR-0001: Microservices over a modular monolith

Status: Accepted (2026-10-22)

## Context
One developer, low traffic. A modular monolith would be faster to build and simpler to test and operate.

## Decision
Split into 8 services anyway, on workload lines: broker sync is bursty I/O, analytics/risk are CPU-bound pandas jobs, the copilot is latency-bound on an external LLM. Each scales on a different signal.

## Consequences
Pay for it in network hops, eventual consistency, distributed tracing and more deployables. Boundaries follow data ownership so services could be folded back into a monolith without redesign.
