# ADR-0008: Envelope encryption for broker tokens

Status: Accepted (2026-10-22)

## Context
Broker OAuth tokens give read access to users' finances.

## Decision
Each record gets a data key from KMS (a static local key in dev) and is encrypted with AES-256-GCM, using account_id as AAD so a ciphertext can't be moved to another row.

## Consequences
Prod depends on KMS; key rotation re-wraps data keys rather than re-encrypting data.
