#!/usr/bin/env bash
# Creates (if missing) + migrates + seeds the dedicated ims_e2e database.
# Safe to re-run: database creation is conditional, migrate/seed/seed-e2e are
# all idempotent. Never touches ims_dev.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
ENV_FILE="$REPO_ROOT/infra/env/.env.e2e"

if [ ! -f "$ENV_FILE" ]; then
  echo "Missing $ENV_FILE — copy infra/env/.env.e2e.example to infra/env/.env.e2e first." >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

: "${DB_USER:=ims_app}"
: "${DB_NAME:=ims_e2e}"
: "${DB_PORT:=5437}"

echo "==> Starting postgres (infra/docker-compose.yml)"
(cd "$REPO_ROOT/infra" && docker compose up -d postgres)

echo "==> Waiting for postgres to accept connections"
until (cd "$REPO_ROOT/infra" && docker compose exec -T postgres pg_isready -U "$DB_USER" -d ims_dev -p "$DB_PORT") >/dev/null 2>&1; do
  sleep 1
done

echo "==> Ensuring database \"$DB_NAME\" exists"
EXISTS="$(cd "$REPO_ROOT/infra" && docker compose exec -T postgres psql -U "$DB_USER" -d ims_dev -p "$DB_PORT" -tAc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'")"
if [ "$EXISTS" != "1" ]; then
  (cd "$REPO_ROOT/infra" && docker compose exec -T postgres psql -U "$DB_USER" -d ims_dev -p "$DB_PORT" -c "CREATE DATABASE \"$DB_NAME\"")
fi

echo "==> Running migrations against $DB_NAME"
(cd "$REPO_ROOT/backend" && go run cmd/migrate/main.go up)

echo "==> Running base seed (permission catalog + reference/demo data) against $DB_NAME"
(cd "$REPO_ROOT/backend" && go run ./cmd/seed)

echo "==> Running E2E fixture seed (e2e_admin, e2e_manager, ...) against $DB_NAME"
(cd "$REPO_ROOT/backend" && go run ./cmd/seed-e2e)

echo "==> ims_e2e ready"
