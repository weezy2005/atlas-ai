# syntax=docker/dockerfile:1.7
# Build context: repository root.
FROM python:3.12-slim AS build
COPY --from=ghcr.io/astral-sh/uv:0.13.5 /uv /bin/uv
ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy UV_PYTHON_DOWNLOADS=never
WORKDIR /app
COPY pyproject.toml uv.lock ./
COPY libs/py-common libs/py-common
COPY services/notifications services/notifications
# extra_ca: see EXTRA_CA_FILE in .env.example.
RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=secret,id=extra_ca,required=false \
    cat /etc/ssl/certs/ca-certificates.crt /run/secrets/extra_ca > /tmp/ca.pem 2>/dev/null \
      || cp /etc/ssl/certs/ca-certificates.crt /tmp/ca.pem; \
    SSL_CERT_FILE=/tmp/ca.pem uv sync --frozen --no-dev --no-editable --package atlas-notifications

FROM python:3.12-slim
RUN useradd --system --uid 10001 --no-create-home atlas
COPY --from=build /app/.venv /app/.venv
ENV PATH=/app/.venv/bin:$PATH PYTHONUNBUFFERED=1
USER 10001
EXPOSE 8000
HEALTHCHECK --interval=10s --timeout=3s --start-period=45s --retries=3 \
    CMD ["python", "-c", "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8000/health', timeout=2)"]
CMD ["uvicorn", "atlas_notifications.main:app", "--host", "0.0.0.0", "--port", "8000", "--no-access-log"]
