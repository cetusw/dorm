export const RECURRENCE_OPTIONS = [
    { value: '0', label: 'Одноразовая' },
    { value: '1', label: 'Еженедельная' },
    { value: '2', label: 'Каждые две недели' },
    { value: '4', label: 'Ежемесячная' },
    { value: '12', label: 'Сезонная' },
] as const

export const RECURRENCE_LABELS: Record<number, string> = {
    0: 'Одноразовая',
    1: 'Еженедельная',
    2: 'Каждые две недели',
    4: 'Ежемесячная',
    12: 'Сезонная',
}

export function formatRecurrenceInterval(value: number): string {
    return RECURRENCE_LABELS[value] ?? 'Неизвестно'
}
