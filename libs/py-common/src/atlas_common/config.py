from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class ServiceSettings(BaseSettings):
    """Settings common to every Python service, read from the environment."""

    model_config = SettingsConfigDict(extra="ignore")

    service_name: str
    environment: str = "local"
    log_level: str = "INFO"
    database_url: str | None = None
    redis_url: str | None = None
    kafka_bootstrap_servers: str | None = None
    otel_exporter_otlp_endpoint: str | None = None
    readiness_timeout_seconds: float = Field(default=2.0, gt=0)
