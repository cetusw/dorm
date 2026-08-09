# Уведомления

## Назначение

Система уведомлений разделена на две части:

- доменная запись `notification` в БД;
- опциональная доставка через Web Push.

Этот документ описывает именно предметные сценарии.

## Кто получает уведомления

- участники дежурной команды;
- глава команды, когда все задачи готовы к проверке.

Получатели выбираются по текущему членству команды и текущему `team.leader_id`.

## Типы уведомлений

- `duty_started`
- `take_tasks_saturday_reminder`
- `take_tasks_sunday_reminder`
- `finish_tasks_sunday_reminder`
- `tasks_ready_for_review`
- `duty_completed`

Из этих типов фактически используются:

- `duty_started`
- `take_tasks_saturday_reminder`
- `take_tasks_sunday_reminder`
- `finish_tasks_sunday_reminder`
- `tasks_ready_for_review`

`duty_completed` существует в enum, но в текущем коде не отправляется.

## Источники событий

### События use-case / event bus

- `week.started` -> уведомления о старте дежурства;
- `tasks.ready_for_review` -> уведомление главе команды о готовности к проверке.

### Scheduler

- суббота 09:00 — напомнить взять задачи;
- воскресенье 09:00 — напомнить пользователям ниже goal взять свободные задачи;
- воскресенье 18:00 — напомнить завершить уборку.

## Таблица сценариев

| Событие / время | Получатель | Условие | Тип | Target |
| --- | --- | --- | --- | --- |
| `week.started` | все текущие участники команды | создано новое дежурство | `duty_started` | `/app/current-duty` |
| суббота 09:00 | все участники активного duty | notification scheduler включён | `take_tasks_saturday_reminder` | `/app/current-duty?tab=free` |
| воскресенье 09:00 | участники, у которых баллов задач меньше goal | есть свободные задачи | `take_tasks_sunday_reminder` | `/app/current-duty?tab=free` |
| воскресенье 18:00 | участники с незавершёнными задачами и/или ниже goal | в duty ещё есть работа | `finish_tasks_sunday_reminder` | `/app/current-duty` |
| `tasks.ready_for_review` | глава команды | все задачи завершены, но ещё не проверены | `tasks_ready_for_review` | `/app/current-duty?tab=review` |

## Deduplication

Каждое уведомление использует `deduplication_key` вида:

`<type>:<dutyID>:<userID>`

Ограничение `UNIQUE (user_id, deduplication_key)` в таблице `notification` гарантирует, что повторная отправка того же сценария не создаст дубликат.

## Target URL

`target_url` должен быть:

- относительным;
- внутри `/app` или `/app/...`;
- не абсолютным URL;
- не внешним доменом.

Это проверяется доменной валидацией `ValidateNotificationTargetURL`.

## Read/unread

В БД есть поле `read_at`, но в текущем коде:

- нет resident API для списка уведомлений;
- нет сценария пометки прочитанным;
- нет UI центра уведомлений.

То есть запись хранится как журнал доставки и дедупликации, а не как полноценный inbox.

## Где реализовано

- домен: `pkg/core/domain/notification/*.go`
- reminder use-case: `pkg/core/usecase/notification/dutyreminderservice.go`
- event handlers: `tasksreadyhandler.go`, `weekstartedhandler.go`
- delivery: `deliveryservice.go`
