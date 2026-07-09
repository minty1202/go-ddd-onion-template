#!/usr/bin/env bash
set -euo pipefail

MIGRATIONS_DIR="${MIGRATIONS_DIR:-sql/migrations}"
DB_URL="${DATABASE_URL:?DATABASE_URL is not set}"

case "${1:-}" in
  up)
    go tool goose -dir "$MIGRATIONS_DIR" postgres "$DB_URL" up
    ;;
  down)
    go tool goose -dir "$MIGRATIONS_DIR" postgres "$DB_URL" down
    ;;
  reset)
    go tool goose -dir "$MIGRATIONS_DIR" postgres "$DB_URL" reset
    ;;
  create)
    if [ -z "${2:-}" ]; then
      echo "Usage: $0 create <name>"
      exit 1
    fi
    go tool goose -dir "$MIGRATIONS_DIR" create "$2" sql
    ;;
  status)
    go tool goose -dir "$MIGRATIONS_DIR" postgres "$DB_URL" status
    ;;
  *)
    echo "Usage: $0 {up|down|reset|create <name>|status}"
    exit 1
    ;;
esac
