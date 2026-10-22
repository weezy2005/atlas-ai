"""Readiness checks for the dependencies a service talks to."""

import asyncio
from collections.abc import Awaitable, Callable
from dataclasses import dataclass

from redis.asyncio import Redis
from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncEngine

CheckFn = Callable[[], Awaitable[None]]


@dataclass(frozen=True)
class ReadinessCheck:
    name: str
    check: CheckFn


def postgres_check(engine: AsyncEngine) -> ReadinessCheck:
    async def check() -> None:
        async with engine.connect() as conn:
            await conn.execute(text("SELECT 1"))

    return ReadinessCheck("postgres", check)


def redis_check(client: "Redis") -> ReadinessCheck:
    async def check() -> None:
        await client.ping()

    return ReadinessCheck("redis", check)


def kafka_check(bootstrap_servers: str) -> ReadinessCheck:
    """TCP reachability of the first bootstrap broker.

    Replaced by a metadata request once services hold a real Kafka client (Phase 3).
    """
    host, _, port = bootstrap_servers.split(",")[0].strip().rpartition(":")

    async def check() -> None:
        _, writer = await asyncio.open_connection(host, int(port))
        writer.close()
        await writer.wait_closed()

    return ReadinessCheck("kafka", check)


async def run_checks(checks: list[ReadinessCheck], per_check_seconds: float) -> dict[str, str]:
    async def one(c: ReadinessCheck) -> tuple[str, str]:
        try:
            await asyncio.wait_for(c.check(), per_check_seconds)
        except Exception as exc:
            return c.name, f"error: {type(exc).__name__}"
        return c.name, "ok"

    return dict(await asyncio.gather(*(one(c) for c in checks)))
