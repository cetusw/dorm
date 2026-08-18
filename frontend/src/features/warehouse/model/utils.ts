import dayjs from 'dayjs'

import type { WarehouseMovement } from './types'

export function formatWarehouseMovementDate(value: string): string {
    const parsed = dayjs(value)
    return parsed.isValid() ? parsed.format('DD.MM.YYYY') : value
}

export function formatWarehouseMovementQuantity(movement: WarehouseMovement): string {
    const prefix = movement.type === 'add' ? '+' : '-'
    return `${prefix}${movement.quantity}`
}
