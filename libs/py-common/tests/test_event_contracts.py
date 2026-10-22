"""Contract tests: libs/events schemas are valid and match the Python envelope model."""

import json
import uuid
from pathlib import Path
from typing import Any

import pytest
from jsonschema import Draft202012Validator, FormatChecker

from atlas_common.events import EventEnvelope, uuid7

EVENTS_DIR = Path(__file__).resolve().parents[2] / "events"
SCHEMAS = EVENTS_DIR / "schemas"
TOPICS = (EVENTS_DIR / "topics.txt").read_text().split()


def _schema(name: str) -> dict[str, Any]:
    return json.loads((SCHEMAS / f"{name}.json").read_text())  # type: ignore[no-any-return]


def _validator(name: str) -> Draft202012Validator:
    return Draft202012Validator(_schema(name), format_checker=FormatChecker())


@pytest.mark.parametrize("path", sorted(SCHEMAS.glob("*.json")), ids=lambda p: p.stem)
def test_schema_is_valid_draft_2020_12(path: Path) -> None:
    Draft202012Validator.check_schema(json.loads(path.read_text()))


def test_every_topic_has_a_payload_schema_and_no_orphans() -> None:
    payload_schemas = {p.stem for p in SCHEMAS.glob("*.json")} - {"envelope.v1"}
    assert payload_schemas == set(TOPICS)


@pytest.mark.parametrize("topic", TOPICS)
def test_topics_are_versioned(topic: str) -> None:
    assert topic.rsplit(".", 1)[1].startswith("v")


def test_python_envelope_serializes_to_valid_schema_instance() -> None:
    env = EventEnvelope(
        event_type="portfolio.updated.v1",
        schema_version=1,
        user_id=uuid.uuid4(),
        producer="portfolio",
        trace_context={"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"},
        payload={
            "snapshot_id": str(uuid.uuid4()),
            "as_of": "2026-10-22T20:00:00Z",
            "total_value": "125034.17",
            "position_count": 14,
        },
    )
    instance = json.loads(env.to_json_bytes())
    _validator("envelope.v1").validate(instance)
    _validator(env.event_type).validate(instance["payload"])


def test_money_must_be_decimal_string() -> None:
    bad = {
        "snapshot_id": str(uuid.uuid4()),
        "as_of": "2026-10-22T20:00:00Z",
        "total_value": 125034.17,
        "position_count": 1,
    }
    assert not _validator("portfolio.updated.v1").is_valid(bad)


def test_envelope_rejects_unversioned_event_type() -> None:
    with pytest.raises(ValueError):
        EventEnvelope(
            event_type="portfolio.updated",
            schema_version=1,
            user_id=uuid.uuid4(),
            producer="x",
            payload={},
        )


def test_uuid7_is_version_7_and_time_ordered() -> None:
    ids = [uuid7() for _ in range(1000)]
    assert all(i.version == 7 for i in ids)
    assert [i.int >> 80 for i in ids] == sorted(i.int >> 80 for i in ids)
    assert len(set(ids)) == len(ids)
