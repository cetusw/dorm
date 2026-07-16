import type { TaskFormValues } from './types'

function validateRequiredPositiveNumber(value: string, emptyMessage: string, invalidMessage: string) {
    const trimmed = value.trim()

    if (trimmed.length === 0) {
        return emptyMessage
    }

    const parsed = Number(trimmed)
    if (!Number.isInteger(parsed) || parsed <= 0) {
        return invalidMessage
    }

    return null
}

export const taskFormValidation = {
    title: (value: string) => {
        const trimmed = value.trim()

        if (trimmed.length === 0) {
            return 'Введите название'
        }

        if ([...trimmed].length > 255) {
            return 'Название не должно превышать 255 символов'
        }

        return null
    },
    cost: (value: string) => validateRequiredPositiveNumber(value, 'Введите стоимость', 'Стоимость должна быть больше нуля'),
    frequency: (value: string) => validateRequiredPositiveNumber(value, 'Введите частоту', 'Частота должна быть больше нуля'),
} satisfies Partial<Record<keyof TaskFormValues, (value: string) => string | null>>
