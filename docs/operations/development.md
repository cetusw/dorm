# Локальная разработка

## Требования

- Docker и Docker Compose
- Node.js 22+ для локального frontend dev
- npm
- Go 1.25, если запускать backend вне контейнера

## Подготовка env

1. Создать `.env`:

```bash
cp .env.example .env
```

2. Заполнить минимум:

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_ROOT_PASSWORD`
- `DB_NAME`
- `AUTH_SECRET`
- `TZ`

Для локального Docker Compose обычно подходят:

- `DB_HOST=db`
- `DB_PORT=3306`
- `TZ=Europe/Moscow`

## Быстрый локальный запуск в контейнерах

Основной скрипт:

```bash
./bin/dorm-up.sh
```

Он делает:

1. поднимает `db`;
2. ждёт `healthy`;
3. запускает миграции через сервис `migrate`;
4. поднимает `app` и `adminer`.

Адреса:

- приложение: `http://localhost:8080/app/`
- Adminer: `http://localhost:8081`
- MySQL host port: `localhost:3307`

Остановить:

```bash
./bin/dorm-down.sh
```

Перезапустить с пересборкой app:

```bash
./bin/dorm-restart.sh
```

## Только миграции

```bash
./bin/dorm-migrate.sh
```

## Frontend dev server

Локально frontend можно запускать отдельно от контейнерного backend:

```bash
./bin/frontend.sh dev
```

Особенности:

- Vite слушает `https://0.0.0.0:5173`
- dev cert лежит в `frontend/.cert/`
- `/api` проксируется на `http://127.0.0.1:8080`

Если зависимости ещё не установлены:

```bash
./bin/frontend.sh install
```

## Запуск backend вне Docker

Если БД уже поднята и env настроен:

```bash
go run ./cmd/main.go
```

Сервер слушает `:8080`.

## Инициализация первого пользователя

Скрипт `./bin/bootstrap.sh` предназначен для пустой БД:

```bash
./bin/bootstrap.sh
```

Он:

- проверяет, что контейнер `dorm-db` запущен;
- требует пустые таблицы `user` и `dormitory`;
- интерактивно создаёт первого пользователя;
- создаёт первое общежитие;
- назначает пользователя его главой.

Скрипт нельзя безопасно запускать на уже заполненной базе.

## Полезные вспомогательные команды

Открыть mysql shell в контейнере:

```bash
./bin/dorm-db.sh
```

Открыть shell контейнера БД:

```bash
./bin/dorm-db-bash.sh
```
