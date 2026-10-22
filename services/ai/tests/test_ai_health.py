from fastapi.testclient import TestClient

from atlas_ai.main import app


def test_health() -> None:
    resp = TestClient(app).get("/health")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ok", "service": "ai"}


def test_metrics_exposed() -> None:
    assert "http_requests_total" in TestClient(app).get("/metrics").text
