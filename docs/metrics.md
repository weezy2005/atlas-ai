# Measured metrics

Only numbers that were actually measured go here, each with its method, so every
claim about the project can be reproduced. Machine: Windows 11, Docker Desktop
with 8 CPUs / 5.7 GB RAM allocated.

## Phase 1: skeleton (2026-10-22)

| Metric | Value | How measured |
|---|---|---|
| Services up and passing readiness | 8 / 8 | `make up`, then `make smoke` (`/ready` on each service checks Postgres, Redis, Kafka) |
| Kafka topics provisioned | 27 (9 domain + 9 retry + 9 DLQ) | `kafka-init` container log |
| Cold start to all-healthy (images cached, empty volumes) | 165 s | `time docker compose up -d --wait` after `down -v` |
| Go service image size | 40–46 MB | `docker images` (distroless static, non-root) |
| Python service image size | 323 MB | `docker images` (python:3.12-slim, non-root) |
| Automated tests | 47 Python (incl. 2 Testcontainers integration) + 6 Go packages | `make test` |
| Cross-schema access | denied by Postgres | `auth_svc` querying `portfolio.*` → `permission denied for schema portfolio` |

Observations:
- Cold start is dominated by the JVM services (Kafka, Schema Registry) and six
  Python services importing in parallel on a memory-constrained VM. Analytics
  briefly failed its health check during startup, so the Python health-check
  start period was raised from 10 s to 45 s.
- The Schema Registry image is 4.19 GB on disk. Worth revisiting (e.g. a lighter
  registry such as Apicurio or Redpanda's) if local resource use becomes a problem.

## Planned measurements

| Phase | Metric |
|---|---|
| 2 | Gateway throughput and p95/p99 latency under k6 load; rate limiter accuracy under burst |
| 3 | Sync duration per account; behaviour under injected broker failures |
| 4 | Event lag from `positions.synced` to `analytics.updated`; duplicate-delivery handling |
| 5 | Risk computation time vs. portfolio size |
| 6 | Share of copilot answers whose numbers all trace to tool results |
| 9 | Full-stack load test; chaos-test recovery time |
