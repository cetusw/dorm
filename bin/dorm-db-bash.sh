#!/bin/bash
set -euo pipefail

COMPOSE_FILES="-f docker-compose.yml -f docker-compose.local.yml"

docker compose $COMPOSE_FILES exec db sh