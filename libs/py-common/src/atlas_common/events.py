"""Event envelope shared by every Kafka message (libs/events/schemas/envelope.v1.json)."""

import secrets
import time
import uuid
from datetime import UTC, datetime
from typing import Any

from pydantic import BaseModel, ConfigDict, Field


def uuid7() -> uuid.UUID:
    """Time-ordered UUIDv7 (RFC 9562), so event ids sort by creation time."""
    ms = time.time_ns() // 1_000_000
    rand_a = secrets.randbits(12)
    rand_b = secrets.randbits(62)
    value = (ms << 80) | (0x7 << 76) | (rand_a << 64) | (0b10 << 62) | rand_b
    return uuid.UUID(int=value)


class EventEnvelope(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True)

    event_id: uuid.UUID = Field(default_factory=uuid7)
    event_type: str = Field(pattern=r"^[a-z_]+(\.[a-z_]+)+\.v[0-9]+$")
    schema_version: int = Field(ge=1)
    occurred_at: datetime = Field(default_factory=lambda: datetime.now(UTC))
    user_id: uuid.UUID
    producer: str
    trace_context: dict[str, str] = Field(default_factory=dict)
    payload: dict[str, Any]

    def to_json_bytes(self) -> bytes:
        return self.model_dump_json().encode()
