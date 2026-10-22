"""Auth service. Signup, login, JWT issuance and refresh-token rotation (Phase 2)."""

from atlas_common.app import create_app
from atlas_common.config import ServiceSettings

settings = ServiceSettings(service_name="auth")
app = create_app(settings)
