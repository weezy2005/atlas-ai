"""Analytics service. P&L, returns, sector allocation and attribution (Phase 4)."""

from atlas_common.app import create_app
from atlas_common.config import ServiceSettings

settings = ServiceSettings(service_name="analytics")
app = create_app(settings)
