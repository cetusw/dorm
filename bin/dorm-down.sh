#!/bin/bash
set -euo pipefail

COMPOSE_FILES="-f docker-compose.yml -f docker-compose.local.yml"

echo "Stopping local environment..."
docker compose $COMPOSE_FILES down

echo "Local environment stopped."