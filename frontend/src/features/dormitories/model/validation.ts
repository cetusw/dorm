import type { DormitoryFormValues } from './types'

function validateRequired(message: string, maxLength: number, maxLengthMessage: string) {
    return (value: string) => {
        const trimmedValue = value.trim()

        if (trimmedValue.length === 0) {
            return message
        }

        if ([...trimmedValue].length > maxLength) {
            return maxLengthMessage
        }

        return null
    }
}

export const dormitoryFormValidation = {
    name: validateRequired('Введите название', 255, 'Название не должно превышать 255 символов'),
    city: validateRequired('Введите город', 255, 'Город не должен превышать 255 символов'),
    streetType: validateRequired('Введите тип улицы', 100, 'Тип улицы не должен превышать 100 символов'),
    streetName: validateRequired('Введите название улицы', 100, 'Название улицы не должно превышать 100 символов'),
    houseNumber: validateRequired('Введите номер дома', 50, 'Номер дома не должен превышать 50 символов'),
} satisfies Partial<Record<keyof DormitoryFormValues, (value: string) => string | null>>
