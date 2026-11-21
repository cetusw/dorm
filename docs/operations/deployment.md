# Production deployment

## Границы и конфигурация

Источник истины: `bin/deploy.sh`, `release.sh`, `rollback.sh`, `backup.sh`, Dockerfile и production Compose. Команды с SSH запускаются из локального checkout; серверный каталог определяется как `/home/<SSH-user>/dorm`. Нужны доступ по SSH в BatchMode, Docker/Compose v2, Bash, rsync/scp и GNU утилиты release manager (включая flock, timeout, sync). Пароли SSH интерактивно не запрашиваются.

Локально подготовьте защищённый `.env.prod`: DB_HOST=db, DB_PORT=3306, DB_USER, DB_PASSWORD, DB_ROOT_PASSWORD, DB_NAME, DB_DRIVER=mysql, TZ; `AUTH_ORIGIN` — единственный публичный HTTPS-origin, `AUTH_TRUSTED_PROXIES` — точные IP/CIDR доверенного reverse proxy. Production Compose задаёт `APP_ENV=production`. Proxy должен заменять X-Real-IP адресом клиента. `AUTH_SECRET` не нужен. Web Push и scheduler настраиваются отдельно, см. [уведомления](./notifications.md) и [авторизацию](../architecture/authorization.md).

`.env.prod` копируется как `.env` конкретного релиза с закреплённым DORM_IMAGE. `dorm --check-config` проверяет конфигурацию без подключения к БД; отдельно проверяются БД, схема и готовность credentials. Пример env не содержит production AUTH_ORIGIN/proxy: их надо задать явно. TLS/reverse proxy сопровождаются вне репозитория; app публикуется на `127.0.0.1:8080`, БД наружу не публикуется.

## Состав образа

Dockerfile: Node 22 собирает frontend; Go 1.25 — backend. Runtime Alpine 3.22 содержит CA certificates, tzdata, `dorm`, `dorm-migrate`, `dorm-hash-password`, `config.json`, SQL-миграции и `web/`. Секреты и файл паролей в image не помещают.

## Обычный deployment

Из именованной Git-ветки:

```bash
./bin/deploy.sh admin@server
```

Для явного обновления базовых images:

```bash
./bin/deploy.sh admin@server --pull
```

Без --pull используется обычный Docker cache; script предупреждает о незакоммиченных изменениях, но собирает текущее дерево. Detached HEAD запрещён. Имя `dorm-app:<нормализованная-ветка>-<UTC timestamp>` резервируется, существующий image не перезаписывается.

Последовательность:

1. Проверить локальные инструменты, .env.prod, Compose, mysql/ и серверный доступ, собрать образ и docker save.
2. Зарезервировать `releases/<release>/`, проверить место, доставить image.tar, Compose, env и mysql/; исходники backend/frontend не передаются. Общий release.sh доставляется в `$ROOT/bin/`.
3. Под flock `$ROOT/.deployment.lock` проверить подтверждённый release-state, активный контейнер и image identity, загрузить image, поднять/проверить БД и конфигурацию.
4. Создать проверенный штатный backup. Подготовить shared assets и импортировать assets кандидата/сохраняемых релизов.
5. Проверить clean SQL schema и выполнить SQL up (внешний лимит 900 секунд; прочие Compose-команды обычно 180).
6. Атомарно выбрать `active`, пересоздать только app без зависимостей. Production-приложение при старте проверяет корректность Argon2id всех активных пользователей.
7. До подтверждения проверить image ID и Compose identity, health, ready, HTML, реальные CSS/JS с SHA-256 из выбранного image и shared assets mount. Бюджет проверки около 120 секунд, отдельные HTTP-запросы ограничены.
8. Атомарно сохранить подтверждённые current/previous и версии схемы в `release-state`.

`/health/live` возвращает 204 (Compose healthcheck), `/health/ready` — 204 либо 503 после проверки MySQL. Успешная отдача HTML сама по себе не доказывает readiness.

Images, releases, backups и старые assets сохраняются. Автоматического image prune, удаления volume и rollback БД нет. Retention и свободное место контролирует администратор. Deploy не требует файла паролей или специальной процедуры авторизации.

## Сбои релиза

Ошибка конфигурации или backup останавливает deploy до SQL-миграций. После ошибки DDL проверьте фактическую схему и schema_migrations: SQL может продолжать выполняться после timeout клиента. Скрипт не делает автоматических retry/down/force/restore и сохраняет backup.

После неудачного switch/verification восстановление CURRENT допускается только при неизменной clean schema, совпадающей с подтверждённой версией CURRENT. Перед восстановлением проверяются image identity, конфигурация, символьные ссылки и shared assets; здоровый подтверждённый контейнер без необходимости не пересоздаётся.

При сбое confirmation скрипт перечитывает атомарный release-state и сообщает, было ли подтверждение записано. Он не угадывает результат и не восстанавливает БД. Если состояние невозможно однозначно прочитать, нужна ручная сверка контейнера, selected release и metadata.

## Rollback приложения

```bash
./bin/rollback.sh admin@server
```

Цель — предыдущий подтверждённый image; при рассогласовании selected/container с CURRENT сначала восстанавливается CURRENT. При отличии схемы rollback остановится; только после самостоятельной проверки совместимости допустим `--confirm-schema-compatible`. Этот флаг является подтверждением администратора, а не доказательством совместимости; dirty/unreadable schema по-прежнему блокирует операцию. SQL down и восстановление dump не выполняются.

Release manager не проверяет исторические поколения авторизации Docker-образов. Администратор должен выбирать для rollback только релизы с действующей Argon2id/server-session реализацией и совместимые с текущими данными. Совпадение версии схемы само по себе не подтверждает совместимость старого бинарника.

## Статические ресурсы

`web/app` входит в image. `$ROOT/shared/assets` хранится независимо от релизов; `assets.override.yml` монтирует его read-only в `/app/web/app/assets`. Ресурсы выбранного/сохраняемых images импортируются рекурсивно, существующий путь с другим содержимым вызывает отказ. Старые CSS/JS остаются доступны вкладкам предыдущего релиза.

- index.html, SPA deep links, service-worker.js: `Cache-Control: no-cache`;
- `/app/assets/*`: `public, max-age=31536000, immutable`;
- root PWA-файлы и `/app/icons/*`: revalidation;
- отсутствующие assets/icons/root files: настоящий 404, без HTML fallback.

Inter из `public/assets/fonts/inter-v1/` публикуется тем же механизмом; обновление требует inter-v2 и нового CSS URL. Дополнительный volume или proxy-route для шрифтов не нужен. SW не кеширует SPA через fetch handler. Reverse proxy не должен переопределять эти правила.

## Состояние и резервные копии

В `release-state`: current/previous и их schema versions, attempt/result, backup. Это метаданные подтверждения/последней попытки, а не средство отката БД. Каталоги releases/, backups/ и shared/ сохраняются между релизами. Подробности резервного копирования и ручного восстановления — в [backup-restore](./backup-restore.md).
