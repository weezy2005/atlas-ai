# ADR-0002: Go for the gateway and brokerage sync

Status: Accepted (2026-10-22)

## Context
The gateway handles every request; sync fans out many concurrent broker HTTP calls with retries and backoff.

## Decision
Write these two in Go (chi, pgx, franz-go). Goroutines suit high-concurrency I/O with small memory use and static binaries. Everything numerical stays in Python.

## Consequences
Two toolchains and two CI paths. Event schemas must be language-neutral (see ADR-0005).
