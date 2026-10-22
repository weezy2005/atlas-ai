# ADR-0007: Copilot tool calling with number verification

Status: Accepted (2026-10-22)

## Context
An LLM that invents portfolio numbers is worse than useless.

## Decision
Claude only gets typed tools that call Portfolio/Analytics/Risk. Before an answer is returned, every number in it is extracted and matched against tool results within rounding tolerance. If any fail, the answer is regenerated once, then returned with the unverified figures stripped.

## Consequences
More latency and some false positives (dates, counts), which are handled with an allow-list.
