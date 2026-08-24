#!/bin/bash
set -euo pipefail

COMPOSE_FILES="-f docker-compose.yml -f docker-compose.local.yml"

if [ "$#" -eq 0 ]; then
  set -- up
fi

if [ ! -f ".env" ]; then
  echo "Missing .env file."
  echo "Create it from .env.example:"
  echo "  cp .env.example .env"
  exit 1
fi

echo "Starting database if needed..."
docker compose $COMPOSE_FILES up -d db

echo "Waiting for database health..."
until [ "$(docker inspect -f '{{.State.Health.Status}}' dorm-db)" = "healthy" ]; do
  sleep 2
done

echo "Applying migrations..."
docker compose $COMPOSE_FILES run --rm -T migrate ./dorm-migrate "$@" </dev/null

echo "Migrations applied."
