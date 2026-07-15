import type { AreaFormValues } from './types'

export const areaFormValidation: {
    name: (value: AreaFormValues['name']) => string | null
    floor: (value: AreaFormValues['floor']) => string | null
} = {
    name: (value) => {
        const trimmed = value.trim()

        if (trimmed.length === 0) {
            return 'Введите название'
        }

        if ([...trimmed].length > 255) {
            return 'Название не должно превышать 255 символов'
        }

        return null
    },
    floor: (value) => {
        const trimmed = value.trim()
        if (trimmed.length === 0) {
            return null
        }

        if (!/^-?\d+$/.test(trimmed)) {
            return 'Этаж должен быть числом'
        }

        return null
    },
}
