import json
import logging

from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider

from atlas_common.logging import JsonFormatter, user_id_var


def _format(record: logging.LogRecord) -> dict[str, object]:
    return json.loads(JsonFormatter("portfolio").format(record))  # type: ignore[no-any-return]


def _record(msg: str, **extra: object) -> logging.LogRecord:
    record = logging.LogRecord("atlas.test", logging.INFO, __file__, 1, msg, None, None)
    record.__dict__.update(extra)
    return record


def test_every_line_has_service_trace_and_user_fields() -> None:
    entry = _format(_record("hello"))
    assert entry["service"] == "portfolio"
    assert entry["msg"] == "hello"
    assert entry["level"] == "info"
    assert "trace_id" in entry and "user_id" in entry


def test_user_id_comes_from_context() -> None:
    token = user_id_var.set("0192f0c1-0000-7000-8000-000000000001")
    try:
        assert _format(_record("x"))["user_id"] == "0192f0c1-0000-7000-8000-000000000001"
    finally:
        user_id_var.reset(token)


def test_trace_id_comes_from_active_span() -> None:
    tracer = TracerProvider().get_tracer("test")
    with tracer.start_as_current_span("op") as span:
        expected = format(span.get_span_context().trace_id, "032x")
        assert _format(_record("x"))["trace_id"] == expected
    assert trace.get_current_span().get_span_context().is_valid is False


def test_extra_fields_are_included() -> None:
    assert _format(_record("synced", account_id="a1", positions=12))["positions"] == 12
