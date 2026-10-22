# ADR-0003: Schema-per-service on a shared Postgres cluster

Status: Accepted (2026-10-22)

## Context
Each service must own its data. Seven separate databases is a lot of operational weight for local dev and RDS cost.

## Decision
One Postgres 16 cluster with one schema and one login role per service. A role can only use its own schema. Roles and schemas are provisioned at cluster bootstrap (init script locally, Terraform in prod); tables only ever come from Alembic or golang-migrate.

## Consequences
Isolation comes from grants, not physical separation, so noisy neighbours share I/O. Moving a service to its own instance only means changing its connection string.
