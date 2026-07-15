import type { TeamFormValues } from './types'

export const teamFormValidation: {
    name: (value: TeamFormValues['name']) => string | null
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
}
