import type { UpdateWarehouseMovementRequest, WarehouseMovementRequest } from './types'

export type WarehouseMovementFormValues = {
    quantity: string
    comment: string
}

export const initialWarehouseMovementFormValues: WarehouseMovementFormValues = {
    quantity: '',
    comment: '',
}

function toNullableString(value: string): string | null {
    const trimmed = value.trim()
    return trimmed.length === 0 ? null : trimmed
}

export function toWarehouseMovementRequest(values: WarehouseMovementFormValues): WarehouseMovementRequest {
    return {
        quantity: Number(values.quantity.trim()),
        comment: toNullableString(values.comment),
    }
}

export function toWarehouseMovementUpdateRequest(values: WarehouseMovementFormValues): UpdateWarehouseMovementRequest {
    return {
        quantity: Number(values.quantity.trim()),
        comment: toNullableString(values.comment),
    }
}

export const warehouseMovementValidation = {
    quantity: (value: string) => {
        const trimmed = value.trim()

        if (trimmed.length === 0) {
            return 'Введите количество'
        }

        if (!/^\d+$/.test(trimmed)) {
            return 'Количество должно быть целым числом'
        }

        if (Number(trimmed) <= 0) {
            return 'Количество должно быть больше нуля'
        }

        return null
    },
    comment: (value: string) => {
        if ([...value.trim()].length > 256) {
            return 'Комментарий не должен превышать 256 символов'
        }

        return null
    },
}
