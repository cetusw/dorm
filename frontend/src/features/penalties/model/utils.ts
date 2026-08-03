import dayjs from 'dayjs'

export function formatPenaltyWeight(value: number): string {
    return new Intl.NumberFormat('ru-RU', {
        minimumFractionDigits: 0,
        maximumFractionDigits: 1,
    }).format(value)
}

export function formatPenaltyDate(value: string): string {
    const parsed = dayjs(value)
    return parsed.isValid() ? parsed.format('DD.MM.YYYY') : value
}

export function countPenaltyReasonCharacters(value: string): number {
    return Array.from(value.trim()).length
}

export function getTodayPenaltyDate(): string {
    return dayjs().format('YYYY-MM-DD')
}
