# Миграции

## Инструмент

Проект использует `golang-migrate` через локальный CLI `cmd/migrate`.

Основной код:

- `data/mysql/migrator.go`
- `cmd/migrate/main.go`

Путь миграций:

- `data/mysql/migrations`

## Создание миграции

```bash
./bin/create-migration.sh add_some_change
```

Скрипт создаёт пару файлов:

- `..._add_some_change.up.sql`
- `..._add_some_change.down.sql`

## Локальное применение

Через Docker окружение:

```bash
./bin/dorm-migrate.sh
```

Или напрямую через бинарник:

```bash
go run ./cmd/migrate up
```

## Поддерживаемые команды CLI

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down
go run ./cmd/migrate version
go run ./cmd/migrate force <version>
```

Что делает каждая:

- `up` — применяет все новые миграции;
- `down` — откатывает ровно один шаг;
- `version` — печатает `version=<n> dirty=<bool>`;
- `force` — вручную выставляет версию при dirty-state.

## Почему нельзя менять старые миграции

Применённая миграция уже могла отработать:

- локально у других разработчиков;
- на production;
- на дампах и резервных копиях.

Редактирование старого файла создаёт расхождение между одинаковой версией migration chain и разным фактическим schema state.

Правило проекта: добавлять новую миграцию поверх существующих.

## Требования к `up/down`

- у каждого `up.sql` должен быть `down.sql`;
- `down.sql` должен откатывать схему настолько корректно, насколько это возможно;
- destructive migrations всё равно должны иметь осмысленный rollback, даже если он частичный.

## Проверка schema после миграции

Минимальный цикл:

1. поднять чистую БД;
2. применить `up`;
3. проверить нужные таблицы и индексы;
4. по возможности выполнить критические репозитории/use-case;
5. при необходимости проверить `down` на тестовой базе.

Практически:

```bash
./bin/dorm-up.sh
./bin/dorm-db.sh
```

Дальше можно использовать SQL вроде:

```sql
SHOW CREATE TABLE duty;
SHOW INDEX FROM team;
```

## Dirty migration

Проверка версии:

```bash
go run ./cmd/migrate version
```

Если `dirty=true`, сначала нужно понять, на каком шаге миграция оборвалась. Только после этого использовать:

```bash
go run ./cmd/migrate force <version>
```

`force` исправляет служебную версию migration table, но не чинит схему сам по себе.

## Production flow

Production deploy script использует:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml run --rm -T --no-deps migrate ./dorm-migrate up
```

То есть миграции применяются:

1. после старта БД;
2. до пересоздания приложения.
