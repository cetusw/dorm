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

- обычный житель видит текущие задачи и историю;
- глава общежития видит разделы управления общежитиями, группами, жителями, территориями и каталогом задач;
- управление предупреждениями показывается при `can_manage_penalties=true`.

Frontend только отображает права, полученные от backend. Он не является источником правил доступа.

## API layer

Все HTTP-запросы идут через `shared/api/apiClient.ts`:

- `credentials: 'same-origin'` для cookie-сессии;
- автоматический redirect на `/app/login` при `401`;
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
- `PenaltiesPage` — warnings/penalties и управление индивидуальными задачами жителей через формы и адаптивный Drawer;
- `LoginPage` — вход по логину и паролю.

## PWA и Service Worker

- `main.tsx` регистрирует Service Worker при старте.
- `public/manifest.webmanifest` задаёт standalone scope `/app/`.
- `public/service-worker.js`:
  - принимает push payload;
  - показывает notification;
  - по клику открывает только локальные `/app/**` URL.
  - не содержит `fetch` handler и не реализует offline/cache-first delivery SPA assets.

## Production static delivery

Frontend собирается Vite в `web/app` с hash-based asset именами вида `/app/assets/index-<hash>.css`.

Для production это означает разные требования к кешированию:

- `index.html` и SPA deep links должны revalidate-иться (`Cache-Control: no-cache`), потому что HTML ссылается на конкретные asset hash текущего deploy;
- `/app/assets/*` могут иметь долгий immutable cache, потому что содержимое привязано к hash в имени файла;
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
- контракт имен полей backend response в snake_case.
