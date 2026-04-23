#!/bin/bash
set -e

# Resolve project root (script-safe)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Project root: $PROJECT_ROOT"

# Load environment variables from .env if it exists
# if [ -f ../.env ]; then
#   export $(grep -v '^#' ../.env | xargs)
# fi

# Ensure variables are set
# : "${DB_URL:?DB_URL is not set}"

MIGRATIONS_DIR="$PROJECT_ROOT/internal/db/migrations"

CMD=${1:-up}
STEPS=${2:-}

echo "Running migrations..."

# migrate -path "$MIGRATIONS" -database "$DB_URL" up

# Use Docker to run the migrate tool
docker run --rm \
  --network chat-demo-net \
  -v $MIGRATIONS_DIR:/migrations migrate/migrate  \
  -path=/migrations \
  -database "postgresql://postgres:postgres@postgres:5432/chat-message?sslmode=disable" \
  $CMD $STEPS

echo "Done."

# To create a new migration file, use:
# migrate create -ext sql -dir ./internal/db/migrations -seq <migration_name>

# to run migrate for postgres locally without docker, use:
# go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
# the tags 'postgres' is needed to build the postgres driver