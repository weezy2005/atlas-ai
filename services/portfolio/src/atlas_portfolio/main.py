"""Portfolio service. Normalized positions, transactions, balances and snapshots (Phase 4)."""

from atlas_common.app import create_app
from atlas_common.config import ServiceSettings

settings = ServiceSettings(service_name="portfolio")
app = create_app(settings)
