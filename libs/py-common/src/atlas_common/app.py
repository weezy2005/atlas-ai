"""FastAPI app factory: logging, tracing, metrics, /health, /ready, /metrics."""

import logging
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from fastapi import FastAPI, Response
from fastapi.responses import JSONResponse
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from prometheus_client import CONTENT_TYPE_LATEST, generate_latest
from redis.asyncio import Redis
from sqlalchemy.ext.asyncio import create_async_engine

from atlas_common.config import ServiceSettings
from atlas_common.health import (
    ReadinessCheck,
    kafka_check,
    postgres_check,
    redis_check,
    run_checks,
)
from atlas_common.logging import configure_logging
from atlas_common.metrics import metrics_middleware
from atlas_common.tracing import configure_tracing

log = logging.getLogger(__name__)


def default_checks(settings: ServiceSettings, app: FastAPI) -> list[ReadinessCheck]:
    """Build a check for every dependency the service is configured with."""
    checks: list[ReadinessCheck] = []
    if settings.database_url:
        engine = create_async_engine(settings.database_url, pool_pre_ping=True)
        app.state.db_engine = engine
        checks.append(postgres_check(engine))
    if settings.redis_url:
        client: Redis = Redis.from_url(settings.redis_url)
        app.state.redis = client
        checks.append(redis_check(client))
    if settings.kafka_bootstrap_servers:
        checks.append(kafka_check(settings.kafka_bootstrap_servers))
    return checks


def create_app(settings: ServiceSettings, checks: list[ReadinessCheck] | None = None) -> FastAPI:
    configure_logging(settings.service_name, settings.log_level)
    configure_tracing(settings.service_name, settings.otel_exporter_otlp_endpoint)

    @asynccontextmanager
    async def lifespan(app: FastAPI) -> AsyncIterator[None]:
        log.info("service starting", extra={"environment": settings.environment})
        yield
        if engine := getattr(app.state, "db_engine", None):
            await engine.dispose()
        if client := getattr(app.state, "redis", None):
            await client.aclose()

    app = FastAPI(title=f"atlas-{settings.service_name}", lifespan=lifespan)
    readiness = checks if checks is not None else default_checks(settings, app)
    app.middleware("http")(metrics_middleware(settings.service_name))
    FastAPIInstrumentor.instrument_app(app, excluded_urls="health,ready,metrics")

    @app.get("/health")
    async def health() -> dict[str, str]:
        return {"status": "ok", "service": settings.service_name}

    @app.get("/ready")
    async def ready() -> JSONResponse:
        results = await run_checks(readiness, settings.readiness_timeout_seconds)
        ok = all(v == "ok" for v in results.values())
        return JSONResponse(
            {"status": "ready" if ok else "not_ready", "checks": results},
            status_code=200 if ok else 503,
        )

    @app.get("/metrics")
    async def metrics() -> Response:
        return Response(generate_latest(), media_type=CONTENT_TYPE_LATEST)

    return app
