"""AI service. Copilot: Claude tool calling over Portfolio/Analytics/Risk (Phase 6)."""

from atlas_common.app import create_app
from atlas_common.config import ServiceSettings

settings = ServiceSettings(service_name="ai")
app = create_app(settings)
