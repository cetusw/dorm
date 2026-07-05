#!/bin/bash
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage:"
  echo "  ./bin/create-migration.sh migration_name"
  echo ""
  echo "Example:"
  echo "  ./bin/create-migration.sh add_user_status"
  exit 1
fi

NAME="$1"
TIMESTAMP="$(date +%s)"
DIR="data/mysql/migrations"

mkdir -p "$DIR"

UP_FILE="$DIR/${TIMESTAMP}_${NAME}.up.sql"
DOWN_FILE="$DIR/${TIMESTAMP}_${NAME}.down.sql"

touch "$UP_FILE"
touch "$DOWN_FILE"

echo "Created migration:"
echo "  $UP_FILE"
echo "  $DOWN_FILE"