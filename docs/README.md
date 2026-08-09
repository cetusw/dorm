# Документация `dorm`

`dorm` — Go/React-сервис для организации дежурств и уборок в общежитии или коливинге. Этот каталог описывает текущее реализованное состояние системы, а не исторические планы.

- Архитектура: [overview](./architecture/overview.md), [backend](./architecture/backend.md), [frontend](./architecture/frontend.md), [database](./architecture/database.md), [authorization](./architecture/authorization.md)
- Бизнес-правила: [дежурства](./domain/duties.md), [задачи](./domain/tasks.md), [команды](./domain/teams.md), [жители](./domain/residents.md), [предупреждения](./domain/penalties.md), [уведомления](./domain/notifications.md)
- Resident API: [resident-api.md](./api/resident-api.md)
- Эксплуатация: [development](./operations/development.md), [migrations](./operations/migrations.md), [deployment](./operations/deployment.md), [backup-restore](./operations/backup-restore.md), [notifications](./operations/notifications.md)
- Архитектурные решения: [ADR index](./decisions/README.md)
- Исторические и вспомогательные спецификации: [specifications](./specifications/README.md)

Начинать чтение стоит с [общего обзора](./architecture/overview.md). Для изменения поведения сначала нужно открыть соответствующий документ из `architecture/`, `domain/`, `api/` или `operations/`.
