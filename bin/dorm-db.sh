#!/bin/bash

if [ ! -f .env ]; then
    echo "Ошибка: .env файл не найден!"
    echo "Запустите скрипт из корневой директории проекта."
    exit 1
fi

export $(grep -v '^#' .env | xargs)

if [ -z "$DB_ROOT_PASSWORD" ]; then
    echo "Ошибка: Переменная DB_ROOT_PASSWORD не найдена в .env файле."
    exit 1
fi

docker compose exec db mysql --default-character-set=utf8mb4 -u root -p"$DB_ROOT_PASSWORD" dorm