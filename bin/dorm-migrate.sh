#!/bin/bash

set -e

COMMAND="${1:-up}"

docker compose build migrate
docker compose up -d db
docker compose run --rm migrate ./dorm-migrate "$COMMAND" "${@:2}"
