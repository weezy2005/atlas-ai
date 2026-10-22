# ADR-0011: Alpaca as the first market data provider

Status: Accepted (2026-10-22)

## Context
Analytics and risk need prices and benchmark history.

## Decision
Alpaca market data first (same free account as the broker) behind a MarketDataProvider adapter, cached in Redis. Polygon can be plugged in later.

## Consequences
The free tier is IEX-only for real-time quotes, which is acceptable for daily analytics.
