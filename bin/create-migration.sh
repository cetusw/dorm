#!/bin/bash

MIGRATION_DIR="./data/mysql/migrations"

if [ -z "$1" ]; then
  echo "Ошибка: Необходимо указать название миграции."
  echo "Пример использования: $0 <migration_name>"
  exit 1
fi

mkdir -p "$MIGRATION_DIR"

TIMESTAMP=$(date +%s)

RAW_NAME="$1"
SANITIZED_NAME="${RAW_NAME// /_}"
SANITIZED_NAME="${SANITIZED_NAME//-/_}"

UP_FILE="${MIGRATION_DIR}/${TIMESTAMP}_${SANITIZED_NAME}.up.sql"
DOWN_FILE="${MIGRATION_DIR}/${TIMESTAMP}_${SANITIZED_NAME}.down.sql"

touch "$UP_FILE"
touch "$DOWN_FILE"

echo "Созданы файлы миграции:"
echo "  - $UP_FILE"
echo "  - $DOWN_FILE"

exit 0