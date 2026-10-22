"""Notifications service. Rule-based alerts from analytics and risk events (Phase 7)."""

from atlas_common.app import create_app
from atlas_common.config import ServiceSettings

settings = ServiceSettings(service_name="notifications")
app = create_app(settings)
