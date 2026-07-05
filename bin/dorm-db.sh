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

echo "Database is ready."
echo "MySQL: localhost:3307"