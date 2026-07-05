#!/bin/bash
set -euo pipefail

COMPOSE_FILES="-f docker-compose.yml -f docker-compose.local.yml"

if [ ! -f ".env" ]; then
  echo "Missing .env file."
  echo "Create it from .env.example:"
  echo "  cp .env.example .env"
  exit 1
fi

echo "Starting local database..."
docker compose $COMPOSE_FILES up -d db

echo "Waiting for database health..."
until [ "$(docker inspect -f '{{.State.Health.Status}}' dorm-db)" = "healthy" ]; do
  sleep 2
done

echo "Applying migrations..."
docker compose $COMPOSE_FILES run --rm -T migrate </dev/null

echo "Starting local application and tools..."
docker compose $COMPOSE_FILES up -d --build app adminer

echo "Local environment is up."
docker compose $COMPOSE_FILES ps

echo ""
echo "Admin panel: http://localhost:8080/admin"
echo "Adminer:     http://localhost:8081"
echo "MySQL:       localhost:3307"