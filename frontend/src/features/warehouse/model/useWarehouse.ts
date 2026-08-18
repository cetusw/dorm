import { useCallback, useEffect, useRef, useState } from 'react'

import { getWarehouse } from '../api/warehouseApi'
import type { WarehouseItem } from './types'

function toErrorMessage(error: unknown): string {
    if (error instanceof Error) {
        return error.message
    }

    return 'Не удалось загрузить склад'
}

export function useWarehouse(enabled: boolean) {
    const [items, setItems] = useState<WarehouseItem[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const requestIDRef = useRef(0)

    const mergeItems = useCallback((currentItems: WarehouseItem[], nextItems: WarehouseItem[]) => {
        const currentByID = new Map(currentItems.map((item) => [item.id, item]))

        return nextItems.map((item) => {
            const currentItem = currentByID.get(item.id)

            if (
                currentItem &&
                currentItem.name === item.name &&
                currentItem.quantity === item.quantity
            ) {
                return currentItem
            }

            return item
        })
    }, [])

    const replaceItems = useCallback((nextItems: WarehouseItem[]) => {
        setItems((currentItems) => mergeItems(currentItems, nextItems))
    }, [mergeItems])

    const upsertItem = useCallback((nextItem: WarehouseItem) => {
        setItems((currentItems) => {
            const existingItem = currentItems.find((item) => item.id === nextItem.id)

            if (!existingItem) {
                return [...currentItems, nextItem]
            }

            if (
                existingItem.name === nextItem.name &&
                existingItem.quantity === nextItem.quantity
            ) {
                return currentItems
            }

            return currentItems.map((item) => (item.id === nextItem.id ? nextItem : item))
        })
    }, [])

    const removeItem = useCallback((itemID: string) => {
        setItems((currentItems) => currentItems.filter((item) => item.id !== itemID))
    }, [])

    const sync = useCallback(async (showLoading: boolean) => {
        requestIDRef.current += 1
        const requestID = requestIDRef.current

        if (!enabled) {
            setItems([])
            setError(null)
            setLoading(false)
            return
        }

        if (showLoading) {
            setLoading(true)
            setError(null)
        }

        try {
            const response = await getWarehouse()
            if (requestIDRef.current !== requestID) {
                return
            }

            replaceItems(response.items)
            setError(null)
        } catch (error) {
            if (requestIDRef.current !== requestID) {
                return
            }

            setError(toErrorMessage(error))
        } finally {
            if (showLoading && requestIDRef.current === requestID) {
                setLoading(false)
            }
        }
    }, [enabled, replaceItems])

    const syncInBackground = useCallback(async () => {
        await sync(false)
    }, [sync])

    useEffect(() => {
        void sync(true)
    }, [sync])

    return {
        items,
        loading,
        error,
        upsertItem,
        removeItem,
        syncInBackground,
    }
}
