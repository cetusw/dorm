export type WarehouseItem = {
    id: string
    name: string
    quantity: number
}

export type WarehouseResponse = {
    items: WarehouseItem[]
}

export type CreateWarehouseItemRequest = {
    name: string
    quantity?: number
    comment: string | null
}

export type UpdateWarehouseItemRequest = {
    name: string
}

export type WarehouseMovementRequest = {
    quantity: number
    comment: string | null
}

export type UpdateWarehouseMovementRequest = {
    quantity: number
    comment: string | null
}

export type WarehouseMovementActor = {
    id: string
    full_name: string
}

export type WarehouseMovement = {
    id: string
    type: 'add' | 'write-off'
    quantity: number
    comment: string | null
    created_by: WarehouseMovementActor
    created_at: string
    balance_after: number
}

export type WarehouseItemHistoryResponse = {
    item: WarehouseItem
    movements: WarehouseMovement[]
}
