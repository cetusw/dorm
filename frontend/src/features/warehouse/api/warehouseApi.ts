import { apiRequest } from '../../../shared/api/apiClient'

import type {
    CreateWarehouseItemRequest,
    UpdateWarehouseItemRequest,
    UpdateWarehouseMovementRequest,
    WarehouseItem,
    WarehouseItemHistoryResponse,
    WarehouseMovementRequest,
    WarehouseResponse,
} from '../model/types'

function normalizeWarehouseResponse(response: Record<string, unknown>): WarehouseResponse {
    const items = Array.isArray(response.items) ? response.items : []

    return {
        items: items.map((item) => {
            const record = item as Record<string, unknown>

            return {
                id: String(record.id ?? ''),
                name: String(record.name ?? ''),
                quantity: Number(record.quantity ?? 0),
            }
        }),
    }
}

export async function getWarehouse(): Promise<WarehouseResponse> {
    const response = await apiRequest('/api/v1/warehouse')
    return normalizeWarehouseResponse(await response.json())
}

function normalizeWarehouseItem(response: Record<string, unknown>): WarehouseItem {
    return {
        id: String(response.id ?? ''),
        name: String(response.name ?? ''),
        quantity: Number(response.quantity ?? 0),
    }
}

function normalizeWarehouseItemHistory(response: Record<string, unknown>): WarehouseItemHistoryResponse {
    const item = normalizeWarehouseItem((response.item ?? {}) as Record<string, unknown>)
    const movements = Array.isArray(response.movements) ? response.movements : []

    return {
        item,
        movements: movements.map((movement) => {
            const record = movement as Record<string, unknown>
            const actor = (record.created_by ?? {}) as Record<string, unknown>

            return {
                id: String(record.id ?? ''),
                type: record.type === 'write-off' ? 'write-off' : 'add',
                quantity: Number(record.quantity ?? 0),
                comment: record.comment == null ? null : String(record.comment),
                created_by: {
                    id: String(actor.id ?? ''),
                    full_name: String(actor.full_name ?? ''),
                },
                created_at: String(record.created_at ?? ''),
                balance_after: Number(record.balance_after ?? 0),
            }
        }),
    }
}

export async function createWarehouseItem(request: CreateWarehouseItemRequest): Promise<WarehouseItem> {
    const response = await apiRequest('/api/v1/warehouse/items', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeWarehouseItem(await response.json())
}

export async function getWarehouseItemHistory(itemId: string): Promise<WarehouseItemHistoryResponse> {
    const response = await apiRequest(`/api/v1/warehouse/items/${encodeURIComponent(itemId)}/history`)
    return normalizeWarehouseItemHistory(await response.json())
}

export async function updateWarehouseItem(itemId: string, request: UpdateWarehouseItemRequest): Promise<WarehouseItem> {
    const response = await apiRequest(`/api/v1/warehouse/items/${encodeURIComponent(itemId)}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeWarehouseItem(await response.json())
}

export async function deleteWarehouseItem(itemId: string): Promise<void> {
    await apiRequest(`/api/v1/warehouse/items/${encodeURIComponent(itemId)}`, {
        method: 'DELETE',
    })
}

export async function addWarehouseItemQuantity(itemId: string, request: WarehouseMovementRequest): Promise<WarehouseItem> {
    const response = await apiRequest(`/api/v1/warehouse/items/${encodeURIComponent(itemId)}/add`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeWarehouseItem(await response.json())
}

export async function writeOffWarehouseItemQuantity(itemId: string, request: WarehouseMovementRequest): Promise<WarehouseItem> {
    const response = await apiRequest(`/api/v1/warehouse/items/${encodeURIComponent(itemId)}/write-off`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeWarehouseItem(await response.json())
}

export async function updateWarehouseMovement(
    itemId: string,
    movementId: string,
    request: UpdateWarehouseMovementRequest,
): Promise<WarehouseItem> {
    const response = await apiRequest(
        `/api/v1/warehouse/items/${encodeURIComponent(itemId)}/movements/${encodeURIComponent(movementId)}`,
        {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(request),
        },
    )

    return normalizeWarehouseItem(await response.json())
}

export async function deleteWarehouseMovement(itemId: string, movementId: string): Promise<void> {
    await apiRequest(`/api/v1/warehouse/items/${encodeURIComponent(itemId)}/movements/${encodeURIComponent(movementId)}`, {
        method: 'DELETE',
    })
}
