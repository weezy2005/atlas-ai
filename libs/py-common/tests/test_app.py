from fastapi.testclient import TestClient

from atlas_common.app import create_app
from atlas_common.config import ServiceSettings
from atlas_common.health import ReadinessCheck


async def _ok() -> None:
    return None


async def _down() -> None:
    raise ConnectionRefusedError


def _client(checks: list[ReadinessCheck]) -> TestClient:
    return TestClient(create_app(ServiceSettings(service_name="test"), checks=checks))


def test_health_is_always_ok() -> None:
    resp = _client([ReadinessCheck("postgres", _down)]).get("/health")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ok", "service": "test"}


def test_ready_when_all_checks_pass() -> None:
    resp = _client([ReadinessCheck("postgres", _ok), ReadinessCheck("redis", _ok)]).get("/ready")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ready", "checks": {"postgres": "ok", "redis": "ok"}}


def test_not_ready_names_failing_dependency() -> None:
    resp = _client([ReadinessCheck("postgres", _ok), ReadinessCheck("kafka", _down)]).get("/ready")
    assert resp.status_code == 503
    assert resp.json()["checks"] == {"postgres": "ok", "kafka": "error: ConnectionRefusedError"}


def test_slow_check_times_out() -> None:
    import asyncio

    async def hang() -> None:
        await asyncio.sleep(10)

    app = create_app(
        ServiceSettings(service_name="test", readiness_timeout_seconds=0.05),
        checks=[ReadinessCheck("redis", hang)],
    )
    resp = TestClient(app).get("/ready")
    assert resp.status_code == 503
    assert resp.json()["checks"]["redis"] == "error: TimeoutError"


def test_metrics_use_route_template_labels() -> None:
    client = _client([])
    client.get("/health")
    client.get("/does-not-exist")
    body = client.get("/metrics").text
    assert 'http_requests_total{method="GET",route="/health",service="test",status="200"}' in body
    assert 'route="unmatched"' in body
    assert 'route="/metrics"' not in body
