# Production deployment

## Границы и конфигурация

Источник истины: `bin/deploy.sh`, `release.sh`, `rollback.sh`, `backup.sh`, `migrate-passwords.sh`, Dockerfile и production Compose. Команды с SSH запускаются из локального checkout; серверный каталог определяется как `/home/<SSH-user>/dorm`. Нужны доступ по SSH в BatchMode, Docker/Compose v2, Bash, rsync/scp и GNU утилиты release manager (включая flock, timeout, sync). Пароли SSH интерактивно не запрашиваются.

Локально подготовьте защищённый `.env.prod`: DB_HOST=db, DB_PORT=3306, DB_USER, DB_PASSWORD, DB_ROOT_PASSWORD, DB_NAME, DB_DRIVER=mysql, TZ; `AUTH_ORIGIN` — единственный публичный HTTPS-origin, `AUTH_TRUSTED_PROXIES` — точные IP/CIDR доверенного reverse proxy. Production Compose задаёт `APP_ENV=production`. Proxy должен заменять X-Real-IP адресом клиента. `AUTH_SECRET` не нужен. Web Push и scheduler настраиваются отдельно, см. [уведомления](./notifications.md) и [авторизацию](../architecture/authorization.md).

`.env.prod` копируется как `.env` конкретного релиза с закреплённым DORM_IMAGE. `dorm --check-config` проверяет конфигурацию без подключения к БД; отдельно проверяются БД, схема и готовность credentials. Пример env не содержит production AUTH_ORIGIN/proxy: их надо задать явно. TLS/reverse proxy сопровождаются вне репозитория; app публикуется на `127.0.0.1:8080`, БД наружу не публикуется.

## Состав образа

Dockerfile: Node 22 собирает frontend; Go 1.25 — backend. Runtime Alpine 3.22 содержит CA certificates, tzdata, `dorm`, `dorm-migrate`, `dorm-hash-password`, `dorm-password-migration`, `config.json`, SQL-миграции и `web/`. Label `org.dorm.auth-v2=true` означает поддержку Argon2id и серверных сессий. Секреты и файл паролей в image не помещают.

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
2. Зарезервировать `releases/<release>/`, проверить место, доставить image.tar, Compose, env и mysql/; исходники backend/frontend не передаются. Общие release.sh/migrate-passwords.sh доставляются в `$ROOT/bin/`.
3. Под flock `$ROOT/.deployment.lock` проверить подтверждённый release-state, активный контейнер и image identity, загрузить image, поднять/проверить БД и конфигурацию.
4. Создать проверенный предварительный backup. Подготовить shared assets и импортировать assets кандидата/сохраняемых релизов.
5. Проверить marker и готовность авторизации, выполнить SQL up (внешний лимит 900 секунд; прочие Compose-команды обычно 180).
6. Проверить Argon2id-readiness, атомарно выбрать `active`, пересоздать только app без зависимостей.
7. До подтверждения проверить image ID и Compose identity, health, ready, HTML, реальные CSS/JS с SHA-256 из выбранного image и shared assets mount. Бюджет проверки около 120 секунд, отдельные HTTP-запросы ограничены.
8. Атомарно сохранить подтверждённые current/previous и версии схемы в `release-state`.

`/health/live` возвращает 204 (Compose healthcheck), `/health/ready` — 204 либо 503 после проверки MySQL. Успешная отдача HTML сама по себе не доказывает readiness.

Images, releases, backups и старые assets сохраняются. Автоматического image prune, удаления volume и rollback БД нет. Retention и свободное место контролирует администратор. Обычный последующий deploy не требует файла паролей и отдельного maintenance ради второго backup.

## Первый переход на новую авторизацию

Это процедура для существующей первоначальной SHA-256-установки; наличие процедуры в коде не означает выполненного production-перехода. Все пользователи после него входят заново прежними логинами и паролями: старые stateless-cookie не принимаются.

До начала:

- проверить происхождение установки и release-state; отсутствие marker не доказывает, что переход никогда не выполнялся;
- зарезервировать окно обслуживания, исключить внешние приложения, cron/скрипты, прямые SQL-записи и автоматический запуск контейнеров на всё время операции;
- иметь актуальную копию БД и доступ к исходным паролям активных пользователей;
- проверить конфигурацию нового релиза и возможность forward recovery. Скрипт не умеет автоматически восстанавливать схему/пароли.

Из checkout на рабочей машине:

```bash
./bin/deploy.sh admin@server --prepare-auth --confirm-exclusive-db-writers
```

Подтверждение writers — реальное обязательство администратора, не автоматическая гарантия. `--prepare-auth` также подтверждает, что это исходная старая установка, а не неизвестная восстановленная БД. Наличие auth_session, post-transition schema, marker или следов уже начатой границы блокирует этот первичный режим.

Release manager создаёт preliminary backup при работающем backend, готовит image/assets, останавливает app и проверяет чистое завершение (`false 0 false`, без OOM). Проверяются другие контейнеры Compose/DB-сетей, MySQL Events/работающие event workers и активные InnoDB-транзакции. Внешних writers на всём интервале нельзя исключить снимком processlist — это обязанность администратора. Например, Adminer в сети БД должен быть остановлен.

Затем создаётся отдельный final backup после остановки записи и до DDL, проверяется gzip/схема, сохраняются пути обеих копий и состояние `auth_prepared`. Backend остаётся остановленным. До этого момента SQL/credentials не меняются. Ошибка final backup блокирует дальнейшие шаги, preliminary сохраняется; автоматический старт старого backend в режиме подготовки не выполняется.

На сервере, в `/home/admin/dorm`, подготовьте `/absolute/path/passwords.txt` доверенным локальным редактором при umask 077. Одна строка — логин, пробел, исходный пароль; пустые строки игнорируются, пробелы внутри пароля не поддерживаются. Файл принадлежит запускающему пользователю, обычный, без symlink/hardlinks, права строго 0600, до 1 MiB. Не передавайте содержимое в чат, Git, аргументы процессов, image или логи; не используйте set -x. Файл вводится непосредственно на сервере, скрипт не выполняет SSH/SCP.

```bash
cd /home/admin/dorm
./bin/migrate-passwords.sh --dry-run /absolute/path/passwords.txt
./bin/migrate-passwords.sh /absolute/path/passwords.txt
bash bin/release.sh "$PWD" activate-auth
```

Это корень deployment с bin/, release-state и releases/, не каталог конкретного релиза и не исходный checkout. Используется строго `attempt` подготовленного release-state, а не latest. Пароли передаются контейнеру через stdin, без дополнительного plaintext-файла.

Dry-run блокирует активных пользователей для проверки, но не записывает изменения. Нормализуются логины; дубликат, неизвестный/удалённый логин, неверный пароль, некорректная строка или неполный набор ещё не мигрированных пользователей останавливают операцию. Активные конфликты логинов требуют ручного разрешения. Удалённые записи не изменяются.

Apply повторяет dry-run и проверку writers, фиксирует необратимую границу, выполняет SQL up, затем обновляет пароли одной транзакцией. Старый SHA-256 проверяется только этой утилитой; для каждого нового хеша — собственная случайная соль. Уже корректный Argon2id не заменяется: если пользователь указан, пароль проверяется; отсутствовать в списке могут только уже мигрированные пользователи. Пустой файл не принимается даже при повторном запуске.

После apply проверяется готовность всех активных credentials и сохраняется `auth_ready`; backend по-прежнему остановлен. `activate-auth` заново проверяет image, marker, схему, unique index и readiness, затем выполняет обычное переключение/verification/confirmation. После проверки login, me, logout и отказа отозванной cookie окно обслуживания можно закрыть. Чувствительный входной файл далее удаляется по принятой политике хранения, не автоматически release manager.

## Необратимая граница и восстановление

`$ROOT/.auth-v2-activated` хранит `auth-v2` с переводом строки, владелец текущий, режим 0600, одна ссылка. Создаётся после final backup и успешной проверки паролей, непосредственно перед DDL: временный файл в том же каталоге, sync, hard link без перезаписи, удаление временного имени, sync каталога. Это независимый от MySQL, постоянный marker; его не понижают и не удаляют при rollback/restore. Числовых security floors и auth_security в БД нет.

Любой запуск через deploy, rollback, select/restore проверяет marker, label image, clean schema, наличие структуры сессий/уникальности логина и Argon2id-readiness. Marker без совместимой БД запрещает автоматический запуск даже нового образа. Неизвестная история без marker требует аудита; чистая установка допускается только без прежних следов deployment, пользовательских таблиц и app-контейнера.

| Отказ | Безопасное действие |
| --- | --- |
| До maintenance/backup | Устранить причину; старое приложение могло остаться работающим, проверить report_state |
| После остановки, до marker | Проверить writers, schema/credentials и backups; не запускать старый backend на основании одного отсутствия marker |
| SQL migration | Оставить maintenance; проверить фактический DDL и dirty state, дождаться завершения SQL; подготовить проверенное forward recovery |
| Транзакция паролей | При обычной ошибке rollback всей транзакции; при ошибке commit исход неизвестен — повторный dry-run и проверка БД до решения о retry |
| После marker/credentials, при старте/verification | Только совместимый образ и forward recovery, старый backend запрещён |
| Сбой атомарного confirmation | Проверить durable release-state и фактический контейнер; script не угадывает результат подтверждения |

При подходящем сохранённом состоянии migrate-passwords допускает повторный dry-run/apply без rehash уже мигрированных пользователей. Dirty schema сначала требует ручного аудита; слепой повтор не является ремонтом DDL. `activate-auth` поддерживает проверку уже запущенного правильного кандидата после сбоя switch/verify. Не редактируйте release-state/marker, чтобы обойти проверку.

## Rollback приложения

```bash
./bin/rollback.sh admin@server
```

Цель — предыдущий подтверждённый image; при рассогласовании selected/container с CURRENT сначала восстанавливается CURRENT. При отличии схемы rollback остановится; только после самостоятельной проверки совместимости допустим `--confirm-schema-compatible`. Этот флаг не обходит marker, capability, readiness и dirty checks. SQL down и восстановление dump не выполняются.

Автоматическое восстановление CURRENT после неудачного switch/verify допускается только без пересечения security boundary, при неизменной clean schema и повторной проверке авторизации. После первой миграции возврат к прежнему SHA-256/stateless backend запрещён даже при формально совместимой схеме.

## Статические ресурсы

`web/app` входит в image. `$ROOT/shared/assets` хранится независимо от релизов; `assets.override.yml` монтирует его read-only в `/app/web/app/assets`. Ресурсы выбранного/сохраняемых images импортируются рекурсивно, существующий путь с другим содержимым вызывает отказ. Старые CSS/JS остаются доступны вкладкам предыдущего релиза.

- index.html, SPA deep links, service-worker.js: `Cache-Control: no-cache`;
- `/app/assets/*`: `public, max-age=31536000, immutable`;
- root PWA-файлы и `/app/icons/*`: revalidation;
- отсутствующие assets/icons/root files: настоящий 404, без HTML fallback.

Inter из `public/assets/fonts/inter-v1/` публикуется тем же механизмом; обновление требует inter-v2 и нового CSS URL. Дополнительный volume или proxy-route для шрифтов не нужен. SW не кеширует SPA через fetch handler. Reverse proxy не должен переопределять эти правила.

## Состояние и резервные копии

В `release-state`: current/previous и их schema versions, attempt/result, backup, preliminary_backup, final_backup, irreversible_started. Это метаданные подтверждения/последней попытки, а не средство отката БД. Каталоги releases/, backups/, shared/ и marker сохраняются между релизами. Подробности резервного копирования и ручного восстановления — в [backup-restore](./backup-restore.md).
