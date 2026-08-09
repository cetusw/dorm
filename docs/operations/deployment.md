# Deployment

## Источник

Фактический flow описан в:

- `Dockerfile`
- `docker-compose.yml`
- `docker-compose.prod.yml`
- `bin/deploy.sh`

## Что собирается

### Frontend stage

`Dockerfile` сначала:

1. берёт `node:22-alpine`;
2. делает `npm ci` в `frontend/`;
3. выполняет `npm run build`;
4. получает готовый `web/app`.

### Backend stage

Потом:

1. берёт `golang:1.25-alpine`;
2. скачивает `go mod`;
3. копирует repository;
4. копирует built frontend в `./web/app`;
5. собирает:
   - `dorm` из `cmd/main.go`
   - `dorm-migrate` из `cmd/migrate`

### Runtime image

Финальный слой:

- `alpine:3.22`
- `ca-certificates`, `tzdata`
- бинарники `dorm`, `dorm-migrate`
- `config.json`
- `data/mysql/migrations`
- `web/`

## Compose topology

Base compose:

- `app`
- `migrate`
- `db`

Production override:

- `app` публикуется только на `127.0.0.1:8080:8080`
- `db` не публикует порт наружу

Следствие: внешний reverse proxy находится вне repository.

## Переменные окружения

Для deploy используются `.env.prod` -> на сервере копируется как `.env`.

Документировать можно только имена:

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_ROOT_PASSWORD`
- `DB_NAME`
- `DB_DRIVER`
- `AUTH_SECRET`
- `TZ`
- `NOTIFICATION_SCHEDULER_ENABLED`
- `WEB_PUSH_ENABLED`
- `WEB_PUSH_PUBLIC_KEY`
- `WEB_PUSH_PRIVATE_KEY`
- `WEB_PUSH_SUBJECT`

## Фактический flow `bin/deploy.sh`

1. Проверить наличие:
   - `Dockerfile`
   - `docker-compose.yml`
   - `docker-compose.prod.yml`
   - `.env.prod`
   - `mysql/`
2. Собрать image `dorm-app:latest`
3. Сохранить image в `dorm-app.tar`
4. Скопировать tar, compose-файлы, `.env.prod` и каталог `mysql/` на сервер
5. На сервере:
   - загрузить image `docker load`
   - поднять `db`
   - дождаться `healthy`
   - применить миграции
   - пересоздать `app`
   - проверить `http://127.0.0.1:8080/app/`
   - дополнительно попробовать `https://dormkit.ru/app/`
   - вывести `docker compose ps` и tail логов
   - сделать `docker image prune -f`

## Health boundary

Проверка deploy считает приложение готовым, если отвечает:

- локальный `http://127.0.0.1:8080/app/`
- или публичный `https://dormkit.ru/app/`

Собственного `/healthz` endpoint в repository нет.

## Reverse proxy

Repository не содержит конфигурацию Nginx/Caddy/Traefik.

Следовательно:

- TLS termination;
- публичный домен;
- basic auth или другая perimeter-security логика

находятся вне этого кода и должны сопровождаться отдельно.
