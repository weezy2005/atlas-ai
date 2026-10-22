"""Risk service. Concentration, volatility, beta, VaR and stress scenarios (Phase 5)."""

from atlas_common.app import create_app
from atlas_common.config import ServiceSettings

settings = ServiceSettings(service_name="risk")
app = create_app(settings)
