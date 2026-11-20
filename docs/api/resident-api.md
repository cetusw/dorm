# Resident API

Ниже описаны resident self-service, предупреждения, склад, push и используемые SPA сценарии управления. Backend prefix — `/api/v1`; UUID передаются строками, dormitory/area ID — положительными числами. Успешные ответы JSON, кроме 204. API не имеет общего механизма пагинации: не добавляйте неописанные page/limit.

## Общие условия

Cookie и права описаны в [авторизации](../architecture/authorization.md). На всех защищённых маршрутах проверяется серверная v2-сессия и активность пользователя. Любой POST/PUT/PATCH/DELETE, включая login/logout, требует допустимый `Origin`; отсутствие/null/чужой origin даёт 403 до use case. Для JSON body используйте Content-Type application/json.

Ошибки обычно имеют `{"error":"сообщение"}`; исключение task conflict может дополнительно содержать task. Ответ содержит `X-Request-ID`. Общие статусы: 400 — параметры/валидация, 401 — сессия/credentials, 403 — права/CSRF, 404 — отсутствующая сущность, 409 — доменный конфликт, 429 — login limiter, 500 — скрытая внутренняя ошибка, 503 — недоступность проверки сессии/отмена, 504 — deadline. SQL/network ошибки других handlers не гарантируют 503: централизованный ErrorHandler может вернуть общий 500. Временные ошибки не означают logout. API deadline обычно 30 секунд, создания дежурства группы — 90; middleware сессии отдельно 5 секунд.

## Аутентификация

### `POST /api/v1/auth/login`

- Назначение: логин жителя.
- Auth: не требуется.
- Body: `{ "login": string, "password": string }`
- Response `200`: `{ "redirect_url": "/app/tasks" }`
- Ошибки:
  - `400` некорректный JSON
  - `401` неверный логин или пароль
  - `403` недопустимый Origin
  - `429` limiter с фактическим Retry-After
  - `503/504` инфраструктурная ошибка/timeout
- Side effects:
  - создаёт независимую серверную сессию и только затем устанавливает cookie `resident_session`; неудачная попытка сохраняет прежнюю cookie

### `POST /api/v1/auth/logout`

- Назначение: logout.
- Auth: не требуется, но действует на текущую cookie.
- Response: `204 No Content`
- Side effects:
  - отзывает текущую v2-сессию и очищает cookie `resident_session`, не затрагивая другие устройства
- Повторный logout, отсутствие/невалидный формат cookie безопасны. При ошибке серверного отзыва возвращается ошибка, а не ложный 204.

### `GET /api/v1/auth/me`

- Назначение: получить текущего пользователя и его UI-права.
- Auth: обязательна.
- Response `200`:
  - `id`
  - `first_name`
  - `last_name`
  - `can_manage_dormitories`
  - `can_manage_penalties`
  - `can_manage_warehouse`
- Ошибки:
  - `401` cookie отсутствует или невалидна

## Текущее дежурство и история

### `GET /api/v1/resident/current-duty`

- Назначение: текущий resident view дежурства.
- Auth: обязательна.
- Query:
  - `group_id` — опционально, только для доступного выбора группы
  - `dormitory_id` — опционально, используется главой общежития
- Response `200`:
  - `dormitory_id`, `my_dormitory_id` (omitempty), `selected_group_id`, `show_group_select`, `groups`, `my_group` (omitempty)
  - `has_active_duty`
  - `can_manage_tasks`
  - `can_manage_duty_settings`
  - `read_only`
  - `visible_tabs`
  - `notice_message`, `notice_tone`
  - `period_status`
  - `duty_id`, `group`, `team`, `start_date`, `end_date`
  - `cost_per_resident_goal`, `my_taken_cost_sum`
  - `team_members`
  - `tasks`
- Основные ошибки:
  - `400` некорректный `group_id` или `dormitory_id`
  - `403` доступ запрещён
  - `404` группа не найдена

### `GET /api/v1/resident/duties`

- Назначение: история дежурств выбранной группы.
- Auth: обязательна.
- Query:
  - `group_id` — опционально
- Permissions:
  - только глава группы
- Response `200`:
  - `selected_group_id`
  - `show_group_select`
  - `groups`
  - `duties[]` с `id`, `start_date`, `end_date`, `team_leader_name`, `period_status`, `progress`
- Ошибки:
  - `403` доступ запрещён
  - `404` группа не найдена

### `GET /api/v1/resident/duties/:dutyId`

- Назначение: read-only детали конкретного дежурства из истории.
- Auth: обязательна.
- Permissions:
  - только глава доступной группы
- Response `200`: тот же контракт, что `current-duty`, но в read-only режиме. Query group_id/dormitory_id здесь не используются: доступ определяется по duty и текущей группе пользователя.
- Ошибки:
  - `400` некорректный UUID
  - `403` доступ запрещён
  - `404` дежурство не найдено

## Действия с задачами

Все маршруты ниже:

- Method: `POST`
- Auth: обязательна
- Query:
  - `group_id` — опционально
  - `dormitory_id` — опционально
- Успешный ответ: `200` и обновлённый `ResidentCurrentDuty`

### `POST /api/v1/resident/tasks/:taskId/take`

- Назначение: взять свободную задачу.
- Permissions:
  - участник команды этого дежурства;
  - duty должно быть активным или прошлым с outstanding tasks.
- Ошибки:
  - `400` некорректный UUID
  - `403` доступ запрещён или действия с duty недоступны
  - `404` задача не найдена
  - `409` задача уже взята
- Особенность `409`:
  - response может содержать `{ error, task }`, где `task` — актуальное состояние конфликтующей задачи.

### `POST /api/v1/resident/tasks/:taskId/return`

- Назначение: вернуть свою незавершённую задачу.
- Ошибки:
  - `403` доступ запрещён / duty actions unavailable
  - `404` задача не найдена
  - `409` состояние уже изменилось

### `POST /api/v1/resident/tasks/:taskId/complete`

- Назначение: завершить свою задачу.
- Side effects:
  - если это закрывает все задачи недели, публикуется `tasks.ready_for_review`

### `POST /api/v1/resident/tasks/:taskId/open`

- Назначение: отменить своё завершение до проверки.

### `POST /api/v1/resident/tasks/:taskId/verify`

- Назначение: подтвердить завершённую задачу.
- Permissions:
  - текущий участник и глава команды `duty.team_id`, в доступном для действий периоде

### `POST /api/v1/resident/tasks/:taskId/reopen`

- Назначение: вернуть завершённую задачу в работу со стороны главы команды.
- Permissions:
  - текущий участник и глава команды, в доступном для действий периоде

## Push notifications

### `GET /api/v1/notifications/config`

- Назначение: узнать, включён ли Web Push и какой public key использовать.
- Auth: обязательна.
- Response:
  - `enabled`
  - `publicKey | null`

### `POST /api/v1/notifications/subscriptions`

- Назначение: сохранить browser push subscription.
- Auth: обязательна.
- Body:
  - `endpoint`
  - `keys.p256dh`
  - `keys.auth`
- Response: `204`
- Ошибки:
  - `400` невалидные данные подписки
  - `500` не удалось сохранить подписку

### `DELETE /api/v1/notifications/subscriptions`

- Назначение: удалить подписку по endpoint.
- Auth: обязательна.
- Body:
  - `endpoint`
- Response: `204`

## Предупреждения

### `GET /api/v1/penalties/me`

- Любой авторизованный пользователь, без управленческих прав.
- `200`: `{penalties: [{id, reason, weight, issued_on}]}` — собственные активные предупреждения.


### `GET /api/v1/penalties`

- Назначение: список жителей с активными предупреждениями в доступном scope.
- Auth: обязательна.
- Permissions:
  - глава общежития или глава группы своего текущего общежития
- Response:
  - `residents[]` с `user_id`, `full_name`, `total_weight`, `threshold_reached`

### `GET /api/v1/penalties/residents`

- Назначение: поиск жителей, которым можно выдать предупреждение.
- Query:
  - `q` — опциональная строка поиска
- Scope:
  - глава общежития видит жителей своего общежития;
  - глава группы видит тот же список жителей своего текущего общежития.

### `GET /api/v1/penalties/residents/:userId`

- Назначение: активные предупреждения конкретного жителя.
- Response:
  - `user_id`
  - `full_name`
  - `penalties[]`

### `POST /api/v1/penalties`

- Назначение: выдать предупреждение.
- Body:
  - `user_id`
  - `reason`
  - `weight`
  - `issued_on` (`YYYY-MM-DD`)
- Response `201`:
  - `id`, `reason`, `weight`, `issued_on`

### `DELETE /api/v1/penalties/:penaltyId`

- Назначение: закрыть предупреждение; повторное закрытие идемпотентно, `204`.
- Ошибки: 400 невалидный UUID, 403 вне scope, 404 предупреждение не найдено.

Списки предупреждений не пагинируются: сводка сортируется total_weight DESC, затем по ФИО/ID; поиск — ФИО/ID, детали — issued_on DESC, created_at DESC, id ASC. DTO/сортировки: [penalty DTO](../../pkg/core/ports/dto/penalty.go), [query](../../pkg/infrastructure/mysql/query/penalty.go). Поиск возвращает `{residents: [{id, name}]}`; детали — активные записи, не архив закрытых.

## Склад

### `GET /api/v1/warehouse`

- Назначение: получить общий склад текущего общежития.
- Auth: обязательна.
- Permissions:
  - только глава общежития или глава группы своего текущего общежития
- Response `200`:
  - `items[]` с `id`, `name`, `quantity`
- Сортировка:
  - `name ASC`
- Ошибки:
  - `403` доступ запрещён

### `GET /api/v1/warehouse/items/:itemId/history`

- Назначение: получить историю движений одной позиции склада.
- Auth: обязательна.
- Permissions:
  - только глава общежития или глава группы своего текущего общежития
- Response `200`:
  - `item` с `id`, `name`, `quantity`
  - `movements[]` с `id`, `type`, `quantity`, `comment`, `created_by`, `created_at`, `balance_after`
- Сортировка:
  - `created_at DESC`, затем `id DESC`
- Ошибки:
  - `400` некорректный UUID
  - `403` доступ запрещён, в том числе для позиции другого общежития
  - `404` позиция не найдена

### `POST /api/v1/warehouse/items`

- Назначение: создать новую позицию склада.
- Auth: обязательна.
- Body:
  - `name` — обязательно
  - `quantity` — опционально, по умолчанию `0`
  - `comment` — опционально
- Response `201`:
  - `id`, `name`, `quantity`
- Ошибки:
  - `400` некорректный JSON, пустое название, отрицательное `quantity`, слишком длинный `comment`
  - `403` доступ запрещён
  - `409` позиция с таким названием уже существует

### `PUT /api/v1/warehouse/items/:itemId`

- Назначение: переименовать позицию склада.
- Auth: обязательна.
- Body:
  - `name`
- Response `200`:
  - `id`, `name`, `quantity`
- Ошибки:
  - `400` некорректный UUID, некорректное название
  - `403` доступ запрещён, в том числе для позиции другого общежития
  - `404` позиция не найдена
  - `409` позиция с таким названием уже существует

### `DELETE /api/v1/warehouse/items/:itemId`

- Назначение: удалить позицию склада вместе со всей историей движений.
- Auth: обязательна.
- Response: `204 No Content`
- Ошибки:
  - `400` некорректный UUID
  - `403` доступ запрещён, в том числе для позиции другого общежития
  - `404` позиция не найдена

### `POST /api/v1/warehouse/items/:itemId/add`

- Назначение: оприходовать предметы по существующей позиции.
- Auth: обязательна.
- Body:
  - `quantity` — положительное число
  - `comment` — опционально
- Response `200`:
  - `id`, `name`, `quantity`
- Ошибки:
  - `400` некорректный UUID, `quantity <= 0`, слишком длинный `comment`
  - `403` доступ запрещён, в том числе для позиции другого общежития
  - `404` позиция не найдена

### `POST /api/v1/warehouse/items/:itemId/write-off`

- Назначение: списать предметы по существующей позиции.
- Auth: обязательна.
- Body:
  - `quantity` — положительное число
  - `comment` — опционально
- Response `200`:
  - `id`, `name`, `quantity`
- Ошибки:
  - `400` некорректный UUID, `quantity <= 0`, слишком длинный `comment`
  - `403` доступ запрещён, в том числе для позиции другого общежития
  - `404` позиция не найдена
  - `409` недостаточный остаток

### `PUT /api/v1/warehouse/items/:itemId/movements/:movementId`

- Назначение: изменить количество и комментарий существующего поступления или списания.
- Auth: обязательна.
- Body:
  - `quantity` — положительное число
  - `comment` — опционально
- Response `200`:
  - `id`, `name`, `quantity`
- Ошибки:
  - `400` некорректный UUID, `quantity <= 0`, слишком длинный `comment`
  - `403` доступ запрещён, в том числе для позиции другого общежития
  - `404` позиция или движение не найдены
  - `409` изменение делает историю с отрицательным остатком

### `DELETE /api/v1/warehouse/items/:itemId/movements/:movementId`

- Назначение: удалить существующее поступление или списание.
- Auth: обязательна.
- Response: `204 No Content`
- Ошибки:
  - `400` некорректный UUID
  - `403` доступ запрещён, в том числе для позиции другого общежития
  - `404` позиция или движение не найдены
  - `409` удаление делает историю с отрицательным остатком

## Управление жителями

Все `/api/v1/users` требуют сессию и `CanManageDormitories`: лидерство хотя бы одного общежития, без модели роли admin. Этот gate даёт административный доступ; его нельзя документировать как ограничение только своим общежитием.

| Method и URL | Параметры/body | Успешный ответ |
| --- | --- | --- |
| `GET /api/v1/users` | обязательный query dormitory_id | 200 `{users: [{id,name,room_number}]}` |
| `GET /api/v1/users/options` | без параметров | 200 `{users: [{id,name}]}` |
| `GET /api/v1/users/:id` | UUID | 200 ResidentDetails |
| `POST /api/v1/users` | CreateResidentRequest | 201 ResidentDetails |
| `PUT /api/v1/users/:id` | UUID, UpdateResidentRequest | 200 ResidentDetails |
| `DELETE /api/v1/users/:id` | UUID | 204, soft delete и отзыв сессий |

Create/Update body: first_name, last_name, middle_name (nullable), login, **password (обязателен и для PUT)**, dormitory_id, floor (nullable), room_number (nullable). ResidentDetails содержит id, first_name, last_name, middle_name, login, floor, room_number; не возвращает password_hash. Полный DTO — [user.go](../../pkg/core/ports/dto/user.go), ограничения — [жители](../domain/residents.md). Неверные данные и бизнес-конфликт логина здесь дают 400, отсутствующие детали/PUT — 404, отказ gate — 403. Нельзя обещать 409 для конфликта login: handler маппит бизнес-ошибку в 400.

## Настройки дежурства группы

Prefix `/api/v1/groups/:id`, id — UUID группы. Здесь требуется именно текущий глава этой группы; лидерство общежитием само по себе не даёт права. Изменения защищены Origin. Некорректные ID/body — 400, отсутствие ресурса — 404, отказ — 403; ошибки инвариантов settings в основном 400.

| Method и URL | Вход | Ответ |
| --- | --- | --- |
| `POST /api/v1/groups/:id/duties` | start_date/end_date | 204; 409 пересечение или уже созданное дежурство |
| `GET /api/v1/groups/:id/duty-settings` | — | 200 DutySettingsResponse |
| `GET /api/v1/groups/:id/duty-settings/teams/members` | — | 200 options |
| `GET /api/v1/groups/:id/duty-settings/teams/:teamId` | UUID команды | 200 TeamDetails |
| `GET /api/v1/groups/:id/duty-settings/teams/:teamId/members` | UUID команды | 200 TeamMembersResponse |
| `GET /api/v1/groups/:id/duty-settings/teams/:teamId/member-search` | q | 200 TeamSearchResponse |
| `POST /api/v1/groups/:id/duty-settings/teams/:teamId/members` | user_id | 204 |
| `PUT /api/v1/groups/:id/duty-settings/teams/:teamId/leader` | user_id | 204 |
| `DELETE /api/v1/groups/:id/duty-settings/teams/:teamId/members/:userId` | UUID | 204 |
| `POST /api/v1/groups/:id/duty-settings/teams` | CreateResidentTeamRequest | 201 TeamDetails |
| `PUT /api/v1/groups/:id/duty-settings/teams/:teamId` | UpdateResidentTeamRequest | 200 TeamDetails |
| `DELETE /api/v1/groups/:id/duty-settings/teams/:teamId` | UUID | 204 |
| `POST /api/v1/groups/:id/duty-settings/teams/reorder` | team_ids[] | 204 |
| `POST /api/v1/groups/:id/duty-settings/teams/:teamId/assign-active-duty` | — | 204, сброс состояния задач |
| `POST /api/v1/groups/:id/duty-settings/areas` | name, group_id, floor | 201 AreaDetails |
| `PUT /api/v1/groups/:id/duty-settings/areas/:areaId` | name, group_id, floor | 200 AreaDetails |
| `DELETE /api/v1/groups/:id/duty-settings/areas/:areaId` | числовой ID | 204 |
| `POST /api/v1/groups/:id/duty-settings/areas/:areaId/tasks` | CreateTaskRequest | 201 TaskDetails |
| `PUT /api/v1/groups/:id/duty-settings/tasks/:taskId` | UpdateTaskRequest | 200 TaskDetails |
| `DELETE /api/v1/groups/:id/duty-settings/tasks/:taskId` | UUID | 204 |
| `POST /api/v1/groups/:id/duty-settings/tasks/:taskId/include` | — | 204 |
| `DELETE /api/v1/groups/:id/duty-settings/tasks/:taskId/include` | — | 204 |

Ручное создание принимает даты YYYY-MM-DD; parseDutyPeriod приводит их к 09:00 в TZ. Настройки возвращают group, areas/tasks, teams, active_duty_team_id, task_editor_state, task_editor_alert, active_duty. Детали включения/исключения зависят от активного периода и доступности редактора (его календарные границы отличаются от resident view, см. [duties](../domain/duties.md#отличие-редактора-duty-settings)); это не произвольное изменение исторических DutyTask.

JSON-поля настроек описаны в [dutysettings.go](../../pkg/core/ports/dto/dutysettings.go), команд — в [team.go](../../pkg/core/ports/dto/team.go), территорий/задач — в [taskcatalog.go](../../pkg/core/ports/dto/taskcatalog.go). CreateTask: title, cost, **recurrenceInterval**, area_id, include_in_current_duty. UpdateTask — без include_in_current_duty. `startSequence` возвращается, но не задаётся этими request DTO. API использует смесь snake_case, publicKey и этих двух исторических camelCase имён; не переименовывайте их в клиенте.

## Остальной административный API

Область этого документа — Resident API и сценарии SPA, а не полная спецификация всех CRUD DTO. Ниже точный индекс остальных зарегистрированных маршрутов; детали полей см. в связанных DTO/handlers. Все требуют CanManageDormitories и общую сессию, POST/PUT/DELETE — также Origin. CRUD: чтение/PUT 200, POST 201, DELETE 204; параметры/бизнес-валидация обычно 400, отсутствие при чтении 404, gate 403.

| Ресурс | Маршруты | Фильтры и контракт |
| --- | --- | --- |
| Общежития | GET/POST `/api/v1/dormitories`, GET/PUT/DELETE `/api/v1/dormitories/:id` | положительный числовой id; [structure DTO](../../pkg/core/ports/dto/structure.go) |
| Группы | GET/POST `/api/v1/groups`, GET/PUT/DELETE `/api/v1/groups/:id`, GET `/api/v1/groups/options` | GET списка/options требует dormitory_id; [handlers](../../pkg/adapters/http/group.go) |
| Команды | GET/POST `/api/v1/teams`, GET/PUT/DELETE `/api/v1/teams/:id`, GET `/api/v1/teams/members` | GET списка/members: group_id; members также optional team_id; [team DTO](../../pkg/core/ports/dto/team.go) |
| Территории | GET/POST `/api/v1/areas`, GET/PUT/DELETE `/api/v1/areas/:id` | список/POST/PUT/DELETE требуют dormitory_id; числовой id; [handlers](../../pkg/adapters/http/area.go) |
| Каталог | GET/POST `/api/v1/task-definitions`, GET/PUT/DELETE `/api/v1/task-definitions/:id` | все операции требуют dormitory_id, id UUID; [handlers](../../pkg/adapters/http/task_catalog.go) |

## Детали resident DTO

`team` в current-duty — отображаемое имя руководителя команды, не техническое team.name. `tasks` содержит id, area_id/name/floor, title, cost, status, is_mine, can_take/return/complete/open/verify/review_open и optional assignee_id/name (omitempty). Статусы: free, assigned, completed, verified. TeamMembers: id/name, только текущие участники. Пустые коллекции и omitempty определяются DTO и сборщиком ответа; клиент не должен выводить доступ из отсутствия кнопки.

История не пагинируется и сортируется start_date DESC, sequence_number DESC, id DESC; progress содержит total_cost_sum, taken_cost_sum и четыре счётчика задач. Алгоритм выбора текущего/незавершённого прошлого дежурства — [duties](../domain/duties.md); подробные DTO — [residentduty.go](../../pkg/core/ports/dto/residentduty.go).
