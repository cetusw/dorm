# Backend

## Общая модель

Backend реализован как практическая гексагональная схема:

- `pkg/core/domain` хранит сущности и инварианты;
- `pkg/core/ports` описывает интерфейсы use-case и DTO;
- `pkg/core/usecase` реализует бизнес-сценарии;
- `pkg/adapters/http` принимает HTTP и маппит его в use-case;
- `pkg/infrastructure/mysql` реализует хранилище;
- `pkg/infrastructure/di` собирает контейнер;
- `pkg/infrastructure/scheduler` запускает фоновые сценарии;
- `pkg/infrastructure/eventbus` связывает use-case и внешние эффекты без прямой зависимости.

## Направление зависимостей

- `domain` не знает про HTTP, MySQL, cron и browser push.
- `ports` знают только про контракты.
- `usecase` преимущественно зависит от `domain` и `ports`; фактическое исключение — cleaning/distribution.go импортирует тип ошибки MySQL для распознавания конфликта создания дежурства. SQL остаётся в repositories.
- `adapters/http` зависит от `ports` и конкретных use-case интерфейсов.
- реализации `infrastructure` зависят от core-контрактов; business core не импортирует пакеты инфраструктуры проекта.

## Слои

### `pkg/core/domain`

Содержит:

- `catalog`: `Area`, `TaskDefinition`, `DutyTaskOverride`;
- `duty`: `Duty`, `DutyTask`, переходы состояний и ошибки;
- `structure`: `Dormitory`, `Group`, `Team`;
- `user`: `User`;
- `penalty`: `Penalty`;
- `notification`: `Notification`, `PushSubscription`, `NotificationType`.
- `warehouse`: `Item`, `Movement`, `MovementType`.

Здесь должны жить инварианты и допустимые переходы. Здесь не должно быть SQL, HTTP-статусов и сериализации API.

### `pkg/core/ports`

Содержит:

- use-case интерфейсы (`CleaningUseCase`, `ResidentDutyUseCase`, `PenaltyUseCase` и т. д.);
- DTO для HTTP/frontend;
- query-service интерфейсы для read-моделей;
- event bus и push sender контракты;
- `SessionRepository`, `SessionUseCase` и DTO сессии (без bearer token в БД).

### `pkg/core/usecase`

Ключевые сервисы:

- `authentication.Service`: создание, проверка и отзыв серверных v2-сессий;
- `cleaning.Service`: создание дежурств и действия с `DutyTask`;
- `resident.Service`: формирование resident view текущего дежурства и истории;
- `dutysettings.Service`: настройки группы, команд, территорий и задач;
- `penalty.Service`: warnings/penalties и scope доступа;
- `warehouse.Service`: доступ к складу общежития, движения и read-модели истории;
- `notification.*`: подписки, хранение уведомлений и reminder-сценарии;
- `user.Service`, `dormitory.Service`, `team.Service`, `catalog.Service`: CRUD и вспомогательные сценарии.

Use-case слой не должен содержать SQL или прямую работу с Fiber.

### `pkg/adapters/http`

HTTP-адаптеры регистрируются в `pkg/infrastructure/di/container.go`.

Актуальные группы маршрутов:

- `/api/v1/auth`
- `/api/v1/resident`
- `/api/v1/notifications`
- `/api/v1/penalties`
- `/api/v1/warehouse`
- `/api/v1/dormitories`
- `/api/v1/groups`
- `/api/v1/teams`
- `/api/v1/areas`
- `/api/v1/task-definitions`
- `/api/v1/users`

Роль handler-ов:

- прочитать cookie и параметры;
- распарсить body;
- вызвать use-case;
- вернуть JSON и HTTP-код.

Проверка сессии и CSRF выполняется централизованно. Административные handlers дополнительно требуют `CanManageDormitories` и возвращают ненулевую ошибку при отказе. Сценарии resident, duty settings, предупреждений и склада проверяют отношения и область доступа в use cases. Общего RBAC нет; см. [авторизацию](./authorization.md).

### `pkg/infrastructure/mysql`

Содержит:

- `repository/*`: write/read для доменных агрегатов;
- `query/*`: read-модели для списков, истории и уведомлений;
- `connection.go`: SQL connection pool.

Особенно важные детали:

- UUID сериализуются как `BINARY(16)`;
- soft delete есть у `user` и `task`;
- INSERT/UPDATE пользователя разделены; смена пароля и soft delete транзакционно отзывают сессии, выдача сессии блокирует пользователя и сверяет credentials;
- создание `duty` выполняется в транзакции с блокировкой группы и проверкой пересечения периода.

### `pkg/infrastructure/di`

`NewContainer` собирает:

- подключение к БД;
- репозитории и query services;
- use-case сервисы;
- event bus;
- scheduler;
- Fiber app.

Это единственная точка композиции runtime.

### Scheduler и event bus

- `scheduler.Scheduler` читает cron из `config.json` и TZ из env.
- `StartNewWeek` вызывается по cron.
- notification reminder jobs включаются только при `NOTIFICATION_SCHEDULER_ENABLED=true`.
- `cleaning.Service` публикует события:
  - `week.started`
  - `task.assigned`
  - `task.completed`
  - `task.uncompleted`
  - `tasks.ready_for_review`
- сейчас на bus подписаны notification handlers для `week.started` и `tasks.ready_for_review`;
- bus in-memory, не является durable queue; остановка ожидает обработчики;
- cron использует SkipIfStillRunning и контекст задания до 5 минут; при shutdown активные задания отменяются.

### Web Push infrastructure

- `notification.DeliveryService` всегда создаёт запись `notification` в БД.
- При `WEB_PUSH_ENABLED=true` сервис ищет `push_subscription` пользователя и отправляет payload через `webpush-go`.
- Просроченные подписки (`410/404`) удаляются из БД автоматически.

## Реальный сценарий: житель завершает задачу

```mermaid
sequenceDiagram
    participant Browser
    participant HTTP as ResidentAPIHandler
    participant UC as resident.Service
    participant Cleaning as cleaning.Service
    participant DutyRepo as DutyTaskRepository
    participant Bus as EventBus
    participant Notif as TasksReadyForReviewHandler

    Browser->>HTTP: POST /api/v1/resident/tasks/:taskId/complete
    HTTP->>UC: CompleteTask(userID, taskID)
    UC->>Cleaning: CompleteTask(taskID, userID)
    Cleaning->>DutyRepo: UPDATE duty_task SET completion_date=...
    DutyRepo-->>Cleaning: ok/conflict
    Cleaning->>Bus: task.completed
    Cleaning->>Bus: tasks.ready_for_review (если все задачи завершены)
    Bus->>Notif: Handle(event)
    HTTP->>UC: GetCurrentDuty(...)
    UC-->>HTTP: ResidentCurrentDutyResponse
    HTTP-->>Browser: 200 JSON
```

## Что сюда не помещать

- frontend navigation;
- browser-only permission checks;
- ad hoc SQL в handler-ах;
- прямые вызовы push/web API из use-case кроме через порты;
- доменную логику в DTO.

## HTTP lifecycle и диагностика

Fiber: ReadTimeout 15 секунд, WriteTimeout 120, IdleTimeout 60. Контекст API — 30 секунд, ручного создания дежурства группы — 90. Проверка сессии отдельно ограничена 5 секундами. Это cooperative deadlines, не гарантия отмены уже отправленного SQL/DDL.

`diagnostics.go` генерирует Request ID (`X-Request-ID`), делает access logging через slog, panic recovery и централизованный JSON ErrorHandler. Deadline возвращает 504, отмена 503, внутренние детали скрыты. Логи используют шаблон route, без query и credentials; медленный запрос — от 2 секунд.

`GET /health/live` возвращает 204; `/health/ready` делает PingContext до 2 секунд и возвращает 204 либо 503. Docker healthcheck использует live, deploy дополнительно ready.

`cmd/main.go` обрабатывает SIGINT/SIGTERM с общим бюджетом 20 секунд: останавливает scheduler, завершает HTTP, ожидает event handlers, закрывает SQL pool. Compose даёт контейнеру 30 секунд. Production DI перед стартом проверяет готовность Argon2id; `dorm --check-config` проверяет только конфигурацию, без подключения к БД.
