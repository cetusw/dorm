# Frontend

## Общая структура

Frontend находится в `frontend/` и собирается Vite в `web/app`. Это React 19 SPA/PWA без `react-router-dom`.

Структура каталогов фактически близка к FSD, но не является строго формализованной Feature-Sliced схемой:

- `src/app` — shell, провайдеры, ручная навигация и роутинг;
- `src/pages` — page-level composition;
- `src/features` — feature hooks, API и UI;
- `src/entities` — небольшие доменные типы и утилиты, например `duty-task` и `push-subscription`;
- `src/shared` — общий API client, стили и UI primitives;
- `src/widgets` — составные UI-блоки поверх features.

## Зависимости и стек

- UI: Mantine 9
- Icons: Phosphor Icons
- Drag and drop: `@dnd-kit/*`
- Build: Vite 8
- Styling: CSS Modules + `globals.css`

## Навигация

Приложение использует собственную навигацию из `src/app/navigation.ts`:

- `navigateTo()` делает `history.pushState`;
- `replaceTo()` делает `history.replaceState`;
- `useAppPathname()` слушает `popstate` и кастомное событие `app:navigation`.

Это важно: в проекте нет `react-router-dom`, поэтому нельзя предполагать наличие route params, loaders и route guards из router-библиотек.

Роутинг реализован вручную в `src/app/router.tsx` по `pathname`:

- `/app/login`
- `/app/account`
- `/app/warehouse`
- `/app/tasks`
- `/app/duties/history`
- `/app/duties/:dutyId`
- `/app/groups`
- `/app/groups/:groupId/teams`
- `/app/groups/:groupId/duty-settings`
- `/app/dormitories`
- `/app/residents`
- `/app/areas`
- `/app/task-definitions`
- `/app/penalties`

## App shell и условная навигация

`ResidentAppShell` строит меню на основе `currentUser`:

- основной экран обычного жителя — текущие задачи; маршрут истории существует, но данные истории доступны только текущему главе группы;
- глава общежития видит разделы управления общежитиями, группами, жителями, территориями и каталогом задач;
- управление предупреждениями показывается при `can_manage_penalties=true`, складом — при `can_manage_warehouse=true`; уведомления включаются кнопкой shell.

Frontend только отображает права, полученные от backend. Он не является источником правил доступа.

## API layer

Все HTTP-запросы идут через `shared/api/apiClient.ts`:

- `credentials: 'same-origin'` для cookie-сессии;
- исходный `401` остаётся `ApiError`, без повторения исходного запроса;
- перед глобальным redirect выполняется отдельный `/api/v1/auth/me`: только его `401` подтверждает потерю сессии; 200, 403, 5xx, сеть или timeout не перенаправляют;
- параллельные 401 одной версии используют общий Promise проверки с AbortController и лимитом 10 секунд; Promise очищается после любого исхода, рекурсивной проверки `/auth/me` нет;
- login/logout отключают глобальное подтверждение 401; начало/завершение изменения авторизации меняют версию с подпиской через `useSyncExternalStore`; пока операции не завершены, новая проверка пользователя не запускается, устаревшие ответы игнорируются;
- ошибки поднимаются как `ApiError`.

Feature-level API модули:

- `features/auth/api/authApi.ts`
- `features/current-duty/api/currentDutyApi.ts`
- `features/duty-history/api/dutyHistoryApi.ts`
- `features/penalties/api/penaltiesApi.ts`
- `entities/push-subscription/api/pushSubscriptionApi.ts`
- CRUD API для dormitories/groups/teams/areas/users/task-definitions.

## Текущие доменные экраны

- `CurrentDutyPage` — основной экран жителя;
- `DutyHistoryPage`, `DutyHistoryDetailsPage` — история группы;
- `DutySettingsPage` — настройка команд, территорий и задач для группы;
- `DormitoriesPage`, `GroupsPage`, `ResidentsPage`, `AreasPage`, `TaskCatalogPage` — управление структурой;
- `PenaltiesPage` — warnings/penalties;
- `LoginPage` — вход по логину и паролю;
- `AccountPage` — данные пользователя, собственные предупреждения и выход;
- `WarehousePage` — склад и история движений.

## PWA и Service Worker

- `main.tsx` регистрирует Service Worker при старте.
- `public/manifest.webmanifest`: start_url и scope `/app/`, display standalone; PNG any-иконки 192×192 и 512×512 в `/app/icons/`.
- HTML использует `/app/favicon.svg`, `/app/manifest.webmanifest`, `/app/icons/apple-touch-icon.png` (180×180).
- `/`, `/app`, `/app/` в router проверяют существующую сессию: 200 → replace `/app/tasks`, 401 → replace `/app/login`; до ответа loading, при 5xx/сети error и «Повторить». Backend раздаёт SPA на `/app`; корневой `/` требует доставки index внешней конфигурацией, отдельного root route backend нет.
- На защищённой странице без загруженного currentUser показывается loading/error, а не полноценный интерфейс. Хук повторно проверяет сессию после login/logout и при Retry, не применяет ответы после размонтирования.
- `public/service-worker.js`:
  - принимает push payload;
  - показывает notification;
  - по клику открывает только локальные `/app/**` URL.
  - не содержит `fetch` handler и не реализует offline/cache-first delivery SPA assets.

## Production static delivery

Frontend собирается Vite в `web/app` с hash-based asset именами вида `/app/assets/index-<hash>.css`.

Для production это означает разные требования к кешированию:

- `index.html` и SPA deep links должны revalidate-иться (`Cache-Control: no-cache`), потому что HTML ссылается на конкретные asset hash текущего deploy;
- `/app/assets/*` имеют immutable cache на год; для Vite assets версия в hash имени, для public assets требуется версия пути;
- `/app/service-worker.js` не должен иметь immutable-cache, чтобы браузер своевременно замечал новую версию Service Worker;
- отсутствующий `/app/assets/*` должен возвращать настоящий `404`, а не SPA fallback.

Это особенно важно для первой загрузки после deploy: если старый `index.html` или отсутствующий asset ошибочно превращаются в HTML fallback, пользователь может увидеть страницу без актуальных стилей и скриптов.

## Push notifications на клиенте

`usePushNotifications`:

- проверяет secure context;
- проверяет наличие Service Worker, Notifications API и PushManager;
- требует standalone installation на iOS-like устройствах;
- запрашивает `/api/v1/notifications/config`;
- создаёт или удаляет browser subscription;
- синхронизирует её с backend.

## Адаптивность

Фронтенд ориентирован на mobile-first использование:

- sticky/mobile thresholds вынесены в `shared/ui/mobileStickyThreshold.ts`;
- для текущего дежурства есть мобильные карточки, drawer и отдельная подача списка задач;
- shell сочетает sidebar и compact navigation.

## Что важно не ломать

- ручную навигацию и глубокие ссылки;
- server-authoritative permissions;
- cookie-based auth flow;
- PWA scope `/app/`;
- контракт имён полей backend response, включая исторические publicKey/recurrenceInterval/startSequence.

## Локальный Inter

Google Fonts и внешних font CSS нет. `src/shared/styles/globals.css` регистрирует один normal `@font-face` Inter Variable WOFF2, веса 100–900, `font-display: swap`, без `local()`, italic и ограничивающего unicode-range. Mantine и body сохраняют Inter с системными fallback; используются в том числе 400/500/600/700.

Файлы находятся в `frontend/public/assets/fonts/inter-v1/InterVariable.woff2` и `OFL.txt` (оба присутствуют в текущем дереве). Vite копирует их без переименования в `web/app/assets/fonts/inter-v1/`. URL — `/app/assets/fonts/inter-v1/InterVariable.woff2`; кириллица зависит от покрытия файла, дополнительных CSS-настроек не требует.

Сборка допускает отсутствие WOFF2: браузер получает настоящий 404 и использует системные fallback, JavaScript от этого не ломается. Новое содержимое опубликованного immutable-шрифта требует нового каталога (`inter-v2`) и обновления URL; существующий файл не перезаписывают. Docker и shared assets release manager доставляют его тем же механизмом, что CSS/JS, без отдельного volume для шрифтов. См. [deployment](../operations/deployment.md#статические-ресурсы).

SVG-планы этажей в `frontend/assets/building-plans/` преобразуются `npm run generate:plans` в `src/features/current-duty/building-plan/generated/plans.ts`; build и dev запускают генератор. Генерируемый файл не редактируют вручную.

## Ограничение явного выхода

AccountPanel вызывает logout и в finally выполняет `window.location.assign('/app/login')`, даже если серверный отзыв завершился ошибкой. Такой переход UI не подтверждает успешного logout: контракт backend по-прежнему возвращает ошибку при неподтверждённом отзыве. Это известное ограничение клиента; глобальная обработка обычных 401 ему не эквивалентна.
