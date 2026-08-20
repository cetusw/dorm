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
   - обычный deploy использует локальный Docker cache без принудительного обновления base images;
   - для явного обновления base images используется `bin/deploy.sh --pull`
3. Сохранить image в `dorm-app.tar`
4. Скопировать tar, compose-файлы, `.env.prod` и каталог `mysql/` на сервер
5. На сервере:
   - загрузить image `docker load`
   - поднять `db`
   - дождаться `healthy`
   - применить миграции
   - пересоздать `app`
   - предупредить о высоком заполнении `/`, если root filesystem заполнен на 85%+;
   - проверить локальный `http://127.0.0.1:8080/app/`
   - убедиться, что HTML содержит ссылку на реальный `/app/assets/...`
   - убедиться, что найденный asset действительно отвечает как static file, а не как HTML fallback
   - вывести `docker compose ps` и tail логов
   - сделать `docker image prune -f`

## Health boundary

Проверка deploy считает приложение готовым, если локальный `http://127.0.0.1:8080/app/`:

- отвечает успешно;
- возвращает актуальный `index.html`;
- содержит ссылку на существующий `/app/assets/...`.

Собственного `/healthz` endpoint в repository нет.

## Docker build cache

Текущий `Dockerfile` уже устроен так, чтобы не разрушать dependency cache без необходимости:

- `frontend/package*.json` копируются отдельно перед `npm ci`, поэтому изменение frontend-кода само по себе не заставляет повторно устанавливать npm dependencies;
- `go.mod` и `go.sum` копируются отдельно перед `go mod download`, поэтому изменение backend/frontend-кода не заставляет повторно скачивать Go dependencies;
- обычный deploy не должен без явного запроса делать `docker build --pull`.

Это важно для production deploy: базовые images не должны без необходимости заново проверяться и скачиваться при каждом запуске.

## Static delivery semantics для `/app`

Production SPA обслуживается backend-ом так, чтобы разные типы ресурсов имели разную cache policy:

- `index.html` и SPA deep links:
  - `Cache-Control: no-cache`
  - причина: `index.html` содержит ссылки на конкретные hash-based assets текущего deploy и должен всегда revalidate-иться;
- `/app/assets/*`:
  - `Cache-Control: public, max-age=31536000, immutable`
  - причина: Vite генерирует content-hashed filenames, поэтому существующий asset безопасно кешировать долго;
- `/app/service-worker.js`:
  - `Cache-Control: no-cache`
  - причина: браузер должен своевременно проверять новую версию Service Worker;
- отсутствующий `/app/assets/*`:
  - обязан возвращать настоящий `404 Not Found`
  - и не должен попадать в SPA fallback.

Критический инвариант: missing hashed asset никогда не должен получать `200 text/html` с содержимым `index.html`, иначе после deploy браузер может показать страницу без актуальных стилей и скриптов.

## Reverse proxy

Repository не содержит конфигурацию Nginx/Caddy/Traefik.

Следовательно:

- TLS termination;
- публичный домен;
- basic auth или другая perimeter-security логика

находятся вне этого кода и должны сопровождаться отдельно.

Для production reverse proxy необходимо вручную проверить:

- что он не переопределяет backend cache policy;
- что `index.html` не получает долгий TTL;
- что `/app/assets/*` может оставаться immutable-cache;
- что отсутствующий asset не превращается proxy-слоем обратно в HTML fallback.
