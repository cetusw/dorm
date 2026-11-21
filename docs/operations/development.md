# Локальная разработка

## Требования и конфигурация

Команды выполняются из корня checkout. Нужны Docker с Compose v2 и Bash. В Dockerfile используются Go 1.25 и Node.js 22; `go.mod` требует Go 1.25. Для локального frontend подходит Node 22.13+ (нужен `--experimental-strip-types`, используемый генератором планов). `AGENTS.md` упоминает идиомы Go 1.27; фактическая версия сборки — 1.25, это расхождение инструкций, а не требование установленного toolchain.

```bash
cp .env.example .env
```

Замените TODO в `DB_USER`, `DB_PASSWORD`, `DB_ROOT_PASSWORD`, проверьте `DB_NAME=dorm`, `DB_DRIVER=mysql`, `DB_HOST=db`, `DB_PORT=3306`, `TZ=Europe/Moscow`. Не коммитьте `.env`. Пустой APP_ENV означает development; `AUTH_SECRET` не нужен. Development origins и proxy-настройки описаны в [авторизации](../architecture/authorization.md). Web Push по умолчанию выключен.

## Контейнерный запуск

При первом запуске заранее соберите образ app/migrate, чтобы migrate не выбрал старый локальный image:

```bash
docker compose -f docker-compose.yml -f docker-compose.local.yml build app migrate
./bin/dorm-up.sh
```

`dorm-up.sh` поднимает db, ждёт healthy, выполняет migrate up и запускает app/adminer. Скрипт не имеет собственного предела ожидания healthy; при ошибках смотрите состояние db. Адреса:

| Сервис | Адрес |
| --- | --- |
| SPA | `http://localhost:8080/app/` |
| Adminer | `http://localhost:8081` |
| MySQL с хоста | `127.0.0.1:3308` |
| MySQL внутри Compose | `db:3306` |

Сообщения dorm-up/restart о порте 3307 устарели: источник истины — `docker-compose.local.yml` с `3308:3306`.

```bash
./bin/dorm-down.sh
./bin/dorm-restart.sh
./bin/dorm-migrate.sh
./bin/dorm-db-bash.sh
./bin/dorm-db.sh
```

Это отдельные операции: down останавливает окружение без `-v`; restart повторяет запуск с пересборкой app; migrate поднимает БД и применяет up; db-bash открывает sh в db. `dorm-db.sh` открывает mysql с root credentials из `.env`, но имя базы в нём фиксировано как `dorm`; для другого DB_NAME выбирайте базу явно. `.env` должен быть доверенным: bootstrap его выполняет через source, db helper разбирает shell-словами.

## Первый пользователь

После SQL-миграций на пустой установке:

```bash
bash bin/bootstrap.sh
```

В Git bootstrap.sh не имеет executable bit, поэтому используется bash, без изменения прав файла. Нужен работающий `dorm-db`, схема и актуальный образ app с `dorm-hash-password`; работающий `dorm-app` не нужен. Bootstrap использует одноразовый Compose-контейнер с entrypoint утилиты и передаёт пароль через stdin. Затем в транзакции создаёт первого пользователя, общежитие и назначает руководителя. Требует отсутствие активных пользователей и отсутствие общежитий; не предназначен для повторного заполнения существующей базы. Пароль вводится скрыто дважды, минимум 8 символов. Хеш — Argon2id.

Bootstrap не доставляется обычным deploy автоматически. Если он нужен для новой пустой production-установки, администратор отдельно доставляет проверенный скрипт и задаёт `PROJECT_DIR`/`ENV_FILE` каталога конкретного релиза с закреплённым DORM_IMAGE; случайный latest не используют.

## Frontend dev server

```bash
cd frontend
npm ci
cd ..
./bin/frontend.sh dev
```

Vite читает `frontend/.cert/dev-key.pem` и `dev-cert.pem`. Их нужно подготовить самостоятельно (например, доверенным локальным центром сертификатов) с SAN localhost/127.0.0.1; скрипт их не генерирует. Сертификаты не коммитят. URL `https://localhost:5173/app/`, bind `0.0.0.0:5173`. `/api` проксируется на `http://127.0.0.1:8080`; changeOrigin меняет Host, но браузерный Origin должен оставаться разрешённым. Порт 5173 должен быть свободен: другой автоматически выбранный Vite порт не входит в стандартный CSRF allowlist.

`./bin/frontend.sh install` выполняет npm install; для воспроизводимой установки предпочтителен npm ci. `./bin/frontend.sh build` запускает генератор планов, TypeScript и Vite в `web/app`.

## Backend вне Docker

БД должна быть подготовлена. Экспортируйте DB_* с `DB_HOST=127.0.0.1`, `DB_PORT=3308`, `DB_DRIVER=mysql` и TZ; `.env` не загружается самим Go-приложением. Из корня:

```bash
go run ./cmd/main.go
```

Порт 8080 должен быть свободен; для SPA нужен собранный `web/app`. Live — `/health/live`, ready — `/health/ready`.
