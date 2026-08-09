# Web Push и эксплуатация уведомлений

## Что нужно для работы

- `WEB_PUSH_ENABLED=true`
- `WEB_PUSH_PUBLIC_KEY`
- `WEB_PUSH_PRIVATE_KEY`
- `WEB_PUSH_SUBJECT`
- `NOTIFICATION_SCHEDULER_ENABLED=true` для cron reminder-уведомлений
- HTTPS или другой secure context для браузера
- корректно зарегистрированный Service Worker

## Генерация VAPID

В проекте есть CLI:

```bash
go run ./cmd/generate-vapid
```

Он печатает:

- `WEB_PUSH_PUBLIC_KEY=...`
- `WEB_PUSH_PRIVATE_KEY=...`

Значения нужно положить в env, но не коммитить в документацию.

## Включение локально

1. Заполнить в `.env`:
   - `WEB_PUSH_ENABLED=true`
   - `WEB_PUSH_PUBLIC_KEY=...`
   - `WEB_PUSH_PRIVATE_KEY=...`
   - `WEB_PUSH_SUBJECT=mailto:...` или `https://...`
2. При необходимости включить reminder scheduler:
   - `NOTIFICATION_SCHEDULER_ENABLED=true`
3. Перезапустить приложение.

## Как выключить

Отключить Web Push:

```text
WEB_PUSH_ENABLED=false
```

Отключить только cron-напоминания, но оставить возможность immediate event notifications:

```text
NOTIFICATION_SCHEDULER_ENABLED=false
```

## Требования браузера

Клиентский код проверяет:

- `window.isSecureContext`
- наличие `serviceWorker`
- наличие `Notification`
- наличие `PushManager`

На iOS-like устройствах дополнительно требуется standalone installation PWA.

## Service Worker

Используется `frontend/public/service-worker.js`.

Он:

- показывает push notification;
- фильтрует `target_url`, разрешая только локальные `/app/**`;
- по клику открывает или фокусирует окно приложения.

## Что происходит при невалидной подписке

Если push endpoint отвечает `404` или `410`:

- sender возвращает `ErrPushSubscriptionExpired`;
- `DeliveryService` удаляет подписку из `push_subscription`.

Это штатный сценарий очистки устаревших browser subscriptions.

## Где искать проблемы

1. Проверить env и валидацию `WebPushConfig`.
2. Проверить `GET /api/v1/notifications/config`.
3. Проверить, что frontend действительно зарегистрировал Service Worker.
4. Проверить, что браузер дал permission.
5. Проверить наличие записи в `push_subscription`.
6. Проверить логи приложения во время `NotifyUser`.

Типовые причины:

- `WEB_PUSH_ENABLED=false`;
- некорректный `WEB_PUSH_SUBJECT`;
- insecure context;
- iOS приложение не установлено как standalone;
- endpoint устарел и был удалён после `404/410`.
