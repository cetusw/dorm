import type { GroupFormValues } from './types'

export const groupFormValidation = {
    name: (value: string) => {
        const trimmed = value.trim()

        if (trimmed.length === 0) {
            return 'Введите название'
        }

        if ([...trimmed].length > 255) {
            return 'Название не должно превышать 255 символов'
        }

        return null
    },
} satisfies Partial<Record<keyof GroupFormValues, (value: string) => string | null>>
