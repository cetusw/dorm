# Resident API

Ниже описаны фактически зарегистрированные resident-facing маршруты Fiber. Все ответы — JSON, если не указано иное.

## Аутентификация

### `POST /api/v1/auth/login`

- Назначение: логин жителя.
- Auth: не требуется.
- Body: `{ "login": string, "password": string }`
- Response `200`: `{ "redirect_url": "/app/tasks" }`
- Ошибки:
  - `400` некорректный JSON
  - `401` неверный логин или пароль
- Side effects:
  - устанавливает cookie `resident_session`

### `POST /api/v1/auth/logout`

- Назначение: logout.
- Auth: не требуется, но действует на текущую cookie.
- Response: `204 No Content`
- Side effects:
  - очищает cookie `resident_session`

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
  - `selected_group_id`, `groups`, `my_group`
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
  - `duties[]` с `id`, `start_date`, `end_date`, `team_leader_name`, `progress`
- Ошибки:
  - `403` доступ запрещён
  - `404` группа не найдена

### `GET /api/v1/resident/duties/:dutyId`

- Назначение: read-only детали конкретного дежурства из истории.
- Auth: обязательна.
- Permissions:
  - только глава доступной группы
- Response: тот же контракт, что `current-duty`, но в read-only режиме.
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
  - только глава команды текущего `duty.team_id`

### `POST /api/v1/resident/tasks/:taskId/reopen`

- Назначение: вернуть завершённую задачу в работу со стороны главы команды.
- Permissions:
  - только глава команды

## Push notifications

### `GET /api/v1/notifications/config`

- Назначение: узнать, включён ли Web Push и какой public key использовать.
- Auth: обязательна.
- Response:
  - `enabled`
  - `public_key | null`

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

- Назначение: закрыть предупреждение.

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
  - `created_at DESC`, при равном времени tie-breaker по `id`
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

## Management API, доступный из той же SPA

Эти маршруты не относятся к resident self-service, но реально потребляются тем же frontend и требуют auth:

- `/api/v1/dormitories`
- `/api/v1/groups`
- `/api/v1/groups/:id/duty-settings/**`
- `/api/v1/teams`
- `/api/v1/areas`
- `/api/v1/task-definitions`
- `/api/v1/users`

Общее правило:

- доступ в основном ограничен главой общежития;
- исключения внутри `/groups/:id/duty-settings/**` — это сценарии главы группы.
