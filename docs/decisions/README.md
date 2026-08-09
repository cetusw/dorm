# Architecture Decision Records

ADR фиксируют не текущее состояние кода само по себе, а причину важных решений.

Шаблон:

```md
# ADR-NNNN: Название

## Статус

Accepted | Superseded | Deprecated

## Контекст

...

## Решение

...

## Последствия

...
```

Текущий набор:

- [ADR-0001: Hexagonal backend](./0001-hexagonal-backend.md)
- [ADR-0002: Duty sequence](./0002-duty-sequence.md)
- [ADR-0003: Web Push](./0003-web-push.md)

Если решение меняется, старый ADR остаётся историческим документом, а новый создаётся отдельным файлом и при необходимости помечает старый как `Superseded`.
