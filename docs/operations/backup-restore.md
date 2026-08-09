# Backup и restore

В repository нет отдельного backup script. Ниже описана ручная процедура для MySQL.

## Важно

Restore перезапишет локальные данные в целевой базе. Перед восстановлением убедитесь, что база disposable или у вас есть свежий дамп.

## Backup

Пример с переменными окружения:

```bash
export DB_HOST=127.0.0.1
export DB_PORT=3306
export DB_USER=...
export DB_PASSWORD=...
export DB_NAME=dorm

mysqldump \
  --single-transaction \
  --quick \
  --set-gtid-purged=OFF \
  --default-character-set=utf8mb4 \
  -h "$DB_HOST" \
  -P "$DB_PORT" \
  -u "$DB_USER" \
  -p"$DB_PASSWORD" \
  "$DB_NAME" | gzip > "dorm-$(date +%F-%H%M%S).sql.gz"
```

## Transfer

Копирование с production на локальную машину:

```bash
scp user@server:/path/to/dorm-2026-08-09-120000.sql.gz .
```

Если дамп создаётся прямо на сервере:

```bash
ssh user@server '...mysqldump command...' > dorm.sql.gz
```

## Restore локально

Если используется локальный docker compose:

1. поднять БД:

```bash
docker compose -f docker-compose.yml -f docker-compose.local.yml up -d db
```

2. дождаться `healthy`
3. восстановить дамп:

```bash
gzip -dc dorm-2026-08-09-120000.sql.gz | \
docker compose -f docker-compose.yml -f docker-compose.local.yml exec -T db \
  mysql --default-character-set=utf8mb4 -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME"
```

Если нужно полностью заменить локальную схему, сначала можно пересоздать БД вручную внутри mysql shell.

## Verification

Минимальная проверка после restore:

```bash
./bin/dorm-db.sh
```

И в mysql:

```sql
SELECT COUNT(*) FROM user;
SELECT COUNT(*) FROM duty;
SELECT COUNT(*) FROM duty_task;
SELECT COUNT(*) FROM notification;
```

После этого можно поднять приложение и открыть:

- `http://localhost:8080/app/`

Если данные восстановлены некорректно, первым обычно ломается resident login или current-duty экран.
