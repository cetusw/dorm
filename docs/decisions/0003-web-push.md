# ADR-0003: Browser Web Push

## Статус

Accepted

## Контекст

Resident приложение работает как web/PWA интерфейс. Нужны уведомления:

- без открытой вкладки;
- с server-side отправкой;
- без отдельного native mobile приложения.

## Решение

Использовать стандартный browser Web Push:

- Service Worker на фронтенде;
- `PushManager` в браузере;
- VAPID keys в backend env;
- `webpush-go` для доставки;
- backend таблицы `push_subscription` и `notification`;
- сочетание scheduler-triggered и event-triggered уведомлений.

## Последствия

- уведомления зависят от HTTPS и browser push support;
- iOS-like устройства требуют standalone PWA install;
- backend хранит subscriptions и чистит устаревшие endpoint-ы;
- доставка отделена от доменного решения "кому и когда отправлять".
