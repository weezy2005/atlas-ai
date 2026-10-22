# ADR-0009: Mock broker as a first-class adapter

Status: Accepted (2026-10-22)

## Context
The system has to run end to end without real broker credentials.

## Decision
The mock broker implements the same BrokerAdapter interface as Alpaca, with deterministic realistic portfolios, and can inject failures (rate limits, 5xx, timeouts) for chaos testing.

## Consequences
It needs maintenance to keep it realistic.
