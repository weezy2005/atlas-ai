"""Readiness checks against real Postgres and Redis containers."""

from collections.abc import Iterator

import pytest
from fastapi.testclient import TestClient

from atlas_common.app import create_app
from atlas_common.config import ServiceSettings

pytestmark = pytest.mark.integration

testcontainers = pytest.importorskip("testcontainers")
from testcontainers.community.postgres import PostgresContainer  # noqa: E402
from testcontainers.community.redis import RedisContainer  # noqa: E402


@pytest.fixture(scope="module")
def deps() -> Iterator[tuple[str, str]]:
    try:
        pg = PostgresContainer("postgres:16-alpine", driver="asyncpg")
        rd = RedisContainer("redis:7-alpine")
        pg.start()
        rd.start()
    except Exception as exc:  # Docker not available
        pytest.skip(f"docker unavailable: {exc}")
    try:
        redis_url = f"redis://{rd.get_container_host_ip()}:{rd.get_exposed_port(6379)}/0"
        yield pg.get_connection_url(), redis_url
    finally:
        pg.stop()
        rd.stop()


def test_ready_against_real_postgres_and_redis(deps: tuple[str, str]) -> None:
    db_url, redis_url = deps
    settings = ServiceSettings(service_name="it", database_url=db_url, redis_url=redis_url)
    with TestClient(create_app(settings)) as client:
        resp = client.get("/ready")
    assert resp.status_code == 200, resp.text
    assert resp.json()["checks"] == {"postgres": "ok", "redis": "ok"}


def test_not_ready_when_postgres_unreachable(deps: tuple[str, str]) -> None:
    _, redis_url = deps
    settings = ServiceSettings(
        service_name="it",
        database_url="postgresql+asyncpg://x:y@127.0.0.1:1/none",
        redis_url=redis_url,
        readiness_timeout_seconds=1,
    )
    with TestClient(create_app(settings)) as client:
        resp = client.get("/ready")
    assert resp.status_code == 503
    assert resp.json()["checks"]["postgres"].startswith("error:")
    assert resp.json()["checks"]["redis"] == "ok"
