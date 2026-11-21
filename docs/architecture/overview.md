# Обзор системы

## Назначение

`dorm` автоматизирует недельные дежурства:

- хранит общежития, группы, команды, жителей, территории и каталог задач;
- формирует дежурства по группам и ротации команд;
- позволяет жителям брать, выполнять и подтверждать задачи через web-приложение;
- даёт руководителям группы доступ к истории дежурств, настройкам состава команд и каталога задач;
- даёт главе общежития доступ к управлению структурой и жителями;
- ведёт предупреждения (`penalty`), склад с историей движений и Web Push-уведомления.

## Главные сценарии

1. Житель логинится в `/app/login`, получает cookie-сессию и открывает текущее дежурство.
2. Участник дежурной команды берёт задачу, завершает её, а глава команды подтверждает или возвращает в работу.
3. Глава группы вручную создаёт дежурство для своей группы или настраивает команды, территории и задачи.
4. Планировщик запускает новое недельное дежурство и напоминания.
5. Сервис сохраняет уведомление в БД и, если Web Push включён, отправляет push-пакет в браузер.

## Основные части

- Backend: Fiber HTTP API, use-case слой и MySQL-репозитории.
- Frontend: React/Vite PWA под `/app`.
- База данных: MySQL 8, миграции через `golang-migrate`.
- Scheduler: cron-задачи для запуска дежурств и reminder-уведомлений.
- Notifications: доменные уведомления + Web Push инфраструктура.

## Архитектурные границы

```mermaid
flowchart LR
    Browser["Resident SPA / PWA"]
    SW["Service Worker"]
    HTTP["Fiber HTTP adapters"]
    UC["Use cases"]
    Domain["Domain models"]
    Repo["MySQL repositories / query services"]
    Bus["In-memory event bus"]
    Scheduler["Cron scheduler"]
    Push["Web Push sender"]
    DB[("MySQL")]

    Browser --> HTTP
    Browser --> SW
    HTTP --> UC
    UC --> Domain
    UC --> Repo
    UC --> Bus
    Repo --> DB
    Scheduler --> UC
    Bus --> UC
    UC --> Push
    Push --> SW
```

## Внешние границы

- HTTP boundary: `/api/v1/**` и статическая выдача `/app/**`.
- Storage boundary: MySQL, где хранятся и текущие сущности, и история дежурств, и уведомления.
- Browser boundary: Service Worker и Push API.
- Deployment boundary: reverse proxy и TLS находятся вне repository; приложение внутри контейнера слушает `:8080`.

## Аутентификация и сопровождение

[Авторизация](./authorization.md) использует Argon2id и серверные v2-сессии в MySQL. Отдельных Telegram/Google Sheets интеграций и таблицы ролей в runtime нет.

[Release manager](../operations/deployment.md) доставляет versioned Docker-образы, сохраняет предыдущие релизы и статические ресурсы, делает backup и проверяет readiness.
