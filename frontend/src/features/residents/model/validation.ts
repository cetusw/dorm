import type { ResidentFormValues } from './types'

const cyrillicPattern = /^[\p{Script=Cyrillic}\s-]+$/u
const loginPattern = /^[a-z0-9._-]+$/
const passwordPattern = /^[A-Za-z0-9]+$/

function countRunes(value: string): number {
    return [...value].length
}

function validateRequiredCyrillic(value: string, label: string): string | null {
    const trimmed = value.trim()

    if (trimmed.length === 0) {
        return `Введите ${label.toLowerCase()}`
    }

    if (countRunes(trimmed) > 255) {
        return `${label} не должна превышать 255 символов`
    }

    if (!cyrillicPattern.test(trimmed)) {
        return `${label} должна содержать только кириллицу, пробел или дефис`
    }

    return null
}

function validateOptionalCyrillic(value: string, label: string): string | null {
    const trimmed = value.trim()

    if (trimmed.length === 0) {
        return null
    }

    if (countRunes(trimmed) > 255) {
        return `${label} не должно превышать 255 символов`
    }

    if (!cyrillicPattern.test(trimmed)) {
        return `${label} должно содержать только кириллицу, пробел или дефис`
    }

    return null
}

export const residentFormValidation = {
    lastName: (value: string) => validateRequiredCyrillic(value, 'Фамилия'),
    firstName: (value: string) => validateRequiredCyrillic(value, 'Имя'),
    middleName: (value: string) => validateOptionalCyrillic(value, 'Отчество'),
    login: (value: string) => {
        const trimmed = value.trim()

        if (trimmed.length === 0) {
            return 'Введите логин'
        }

        if (countRunes(trimmed) > 255) {
            return 'Логин не должен превышать 255 символов'
        }

        if (!loginPattern.test(trimmed)) {
            return 'Логин должен содержать только латинские буквы, цифры, точку, дефис или подчеркивание'
        }

        return null
    },
    password: (value: string) => {
        const trimmed = value.trim()

        if (trimmed.length === 0) {
            return 'Введите пароль'
        }

        if (trimmed.length < 6) {
            return 'Пароль должен содержать минимум 6 символов'
        }

        if (!passwordPattern.test(trimmed)) {
            return 'Пароль должен содержать только латинские буквы и цифры'
        }

        return null
    },
    floor: (value: string) => {
        const trimmed = value.trim()
        if (trimmed.length === 0) {
            return null
        }

        if (!/^-?\d+$/.test(trimmed)) {
            return 'Этаж должен быть числом'
        }

        return null
    },
    roomNumber: (value: string) => {
        const trimmed = value.trim()
        if (trimmed.length === 0) {
            return null
        }

        if (countRunes(trimmed) > 255) {
            return 'Комната не должна превышать 255 символов'
        }

        return null
    },
} satisfies Partial<Record<keyof ResidentFormValues, (value: string) => string | null>>
