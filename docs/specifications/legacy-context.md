# Legacy context

Этот документ фиксирует важный legacy-контекст repository, который больше не описывает текущий runtime, но полезен для чтения истории проекта.

## Устаревшие ожидания

- В корневом `AGENTS.md` всё ещё упоминаются:
  - HTML-админка;
  - Telegram state machine;
  - Google Sheets отчёты;
  - `group.nextDutyTeam`.
- В миграциях видно удаление:
  - `role`
  - `telegram_id`
  - `spreadsheet_id`
  - `next_duty_team`
  - calendar-based `task.frequency`

## Что актуально вместо этого

- frontend — React SPA в `/app`;
- backend — JSON API на Fiber;
- права — relation-based, без `role_id`;
- recurrence — sequence-based;
- rotation — `team.rotation_position`;
- notifications — browser Web Push.

## Как использовать этот документ

Если старый текст в repository противоречит актуальной документации, ориентироваться нужно на:

1. код;
2. миграции;
3. тесты;
4. документы из `architecture/`, `domain/`, `api/`, `operations/`.
