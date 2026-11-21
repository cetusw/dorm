# Документация `dorm`

`dorm` — Go/React-сервис для организации дежурств и уборок в общежитии или коливинге. Этот каталог описывает текущее реализованное состояние системы.

- Архитектура: [overview](./architecture/overview.md), [backend](./architecture/backend.md), [frontend](./architecture/frontend.md), [database](./architecture/database.md), [authorization](./architecture/authorization.md)
- Бизнес-правила: [дежурства](./domain/duties.md), [задачи](./domain/tasks.md), [команды](./domain/teams.md), [жители](./domain/residents.md), [предупреждения](./domain/penalties.md), [уведомления](./domain/notifications.md)
- Resident API: [resident-api.md](./api/resident-api.md)
- Эксплуатация: [development](./operations/development.md), [migrations](./operations/migrations.md), [deployment](./operations/deployment.md), [backup-restore](./operations/backup-restore.md), [notifications](./operations/notifications.md)
- Архитектурные решения: [ADR](./decisions/README.md)
- Исторические и вспомогательные спецификации: [specifications](./specifications/README.md)

Начинать чтение стоит с [общего обзора](./architecture/overview.md). Для изменения поведения сначала нужно открыть соответствующий документ из `architecture/`, `domain/`, `api/` или `operations/`.

## Порядок знакомства

1. [Архитектурный обзор](./architecture/overview.md) и [связи полномочий/сессии](./architecture/authorization.md).
2. [Локальный запуск](./operations/development.md), затем домен дежурств и [API](./api/resident-api.md).
3. Для сопровождения — [релизы](./operations/deployment.md), [backup/restore](./operations/backup-restore.md), [миграции](./operations/migrations.md).

Документы описывают текущий checkout, но не подтверждают версию или выполнение миграций в production. ADR сохраняют контекст решений; specifications — исторические материалы. Порядок работы агента задаёт корневой `AGENTS.md`, а фактические версии toolchain — go.mod/Dockerfile/package-lock.
