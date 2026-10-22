#!/bin/sh
# Cluster bootstrap (ADR-0003): one login role and one schema per service.
# Runs once, on an empty data directory. Tables are created only by each
# service's migrations (Alembic / golang-migrate), never here.
set -eu

create() {
  role="$1"; password="$2"; shift 2
  psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
    -v role="$role" -v pw="$password" <<'SQL'
CREATE ROLE :"role" LOGIN PASSWORD :'pw';
GRANT CONNECT ON DATABASE atlas TO :"role";
SQL
  for schema in "$@"; do
    psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
      -v role="$role" -v schema="$schema" <<'SQL'
CREATE SCHEMA :"schema" AUTHORIZATION :"role";
SQL
  done
  # First schema is the service's default; it can't see anyone else's.
  psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
    -v role="$role" -v path="$(echo "$@" | tr ' ' ',')" <<'SQL'
ALTER ROLE :"role" SET search_path = :path;
SQL
}

psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" <<'SQL'
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON DATABASE atlas FROM PUBLIC;
SQL

create auth_svc          "$AUTH_DB_PASSWORD"          auth audit
create sync_svc          "$SYNC_DB_PASSWORD"          sync
create portfolio_svc     "$PORTFOLIO_DB_PASSWORD"     portfolio
create analytics_svc     "$ANALYTICS_DB_PASSWORD"     analytics
create risk_svc          "$RISK_DB_PASSWORD"          risk
create ai_svc            "$AI_DB_PASSWORD"            ai
create notifications_svc "$NOTIFICATIONS_DB_PASSWORD" notifications
