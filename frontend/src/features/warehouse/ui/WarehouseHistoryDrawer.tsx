import { useCallback, useEffect, useRef, useState } from 'react'

import { Alert, Center, Drawer, FocusTrap, Loader, Stack, Text } from '@mantine/core'

import { ApiError } from '../../../shared/api/ApiError'
import { EmptyState } from '../../../shared/ui/EmptyState'
import { getWarehouseItemHistory } from '../api/warehouseApi'
import type { WarehouseItem, WarehouseItemHistoryResponse, WarehouseMovement } from '../model/types'
import { DeleteWarehouseMovementModal } from './DeleteWarehouseMovementModal'
import { WarehouseHistoryMovementModal } from './WarehouseHistoryMovementModal'
import { WarehouseHistoryTable } from './WarehouseHistoryTable'
import classes from './WarehouseHistoryDrawer.module.css'

type Props = {
    item: WarehouseItem | null
    opened: boolean
    onClose: () => void
    onChanged: (item: WarehouseItem) => Promise<void> | void
}

function toErrorMessage(error: unknown): string {
    if (error instanceof ApiError) {
        if (error.status === 500) {
            return 'Не удалось загрузить историю'
        }

        return error.message
    }

    if (error instanceof Error && error.message.trim() !== '') {
        return error.message
    }

    return 'Не удалось загрузить историю'
}

export function WarehouseHistoryDrawer({ item, opened, onClose, onChanged }: Props) {
    const [data, setData] = useState<WarehouseItemHistoryResponse | null>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [editingMovement, setEditingMovement] = useState<WarehouseMovement | null>(null)
    const [deletingMovement, setDeletingMovement] = useState<WarehouseMovement | null>(null)
    const historyRequestIDRef = useRef(0)

    const loadHistory = useCallback(async (itemId: string) => {
        const requestID = historyRequestIDRef.current + 1
        historyRequestIDRef.current = requestID

        setLoading(true)
        setError(null)
        setData(null)

        try {
            const response = await getWarehouseItemHistory(itemId)
            if (historyRequestIDRef.current !== requestID) {
                return null
            }

            setData(response)
            return response
        } catch (currentError) {
            if (historyRequestIDRef.current !== requestID) {
                return null
            }

            setError(toErrorMessage(currentError))
            return null
        } finally {
            if (historyRequestIDRef.current === requestID) {
                setLoading(false)
            }
        }
    }, [])

    const resetState = useCallback(() => {
        historyRequestIDRef.current += 1
        setData(null)
        setLoading(false)
        setError(null)
        setEditingMovement(null)
        setDeletingMovement(null)
    }, [])

    const handleClose = useCallback(() => {
        resetState()
        onClose()
    }, [onClose, resetState])

    useEffect(() => {
        if (!opened || !item) {
            return
        }

        void loadHistory(item.id)
    }, [item, loadHistory, opened])

    const handleMovementChanged = useCallback(async (updatedItem: WarehouseItem) => {
        if (!item) {
            return
        }

        await Promise.resolve(onChanged(updatedItem))
        await loadHistory(item.id)
    }, [item, loadHistory, onChanged])

    const handleMovementDeleted = useCallback(async () => {
        if (!item) {
            return
        }

        const response = await loadHistory(item.id)
        if (!response) {
            return
        }

        await Promise.resolve(onChanged(response.item))
    }, [item, loadHistory, onChanged])

    return (
        <Drawer
            opened={opened}
            onClose={handleClose}
            position="right"
            size={760}
            closeOnClickOutside
            title={data ? `История ${data.item.name}` : item ? `История ${item.name}` : 'История'}
        >
            <FocusTrap.InitialFocus />

            <Stack gap="md" className={classes.content}>
                {error ? <Alert color="red">{error}</Alert> : null}

                {data ? (
                    <Text className={classes.balance}>Остаток: {data.item.quantity} шт.</Text>
                ) : null}

                {loading ? (
                    <Center py="xl">
                        <Loader />
                    </Center>
                ) : data && data.movements.length === 0 ? (
                    <EmptyState
                        title="Нет истории движений"
                        description="Этот элемент учёта пока что не добовляли и не списывали"
                    />
                ) : data ? (
                    <WarehouseHistoryTable
                        movements={data.movements}
                        onEdit={setEditingMovement}
                        onDelete={setDeletingMovement}
                    />
                ) : null}
            </Stack>

            <WarehouseHistoryMovementModal
                opened={editingMovement !== null}
                item={data?.item ?? item}
                movement={editingMovement}
                onClose={() => setEditingMovement(null)}
                onSaved={handleMovementChanged}
            />

            <DeleteWarehouseMovementModal
                opened={deletingMovement !== null}
                item={data?.item ?? item}
                movement={deletingMovement}
                onClose={() => setDeletingMovement(null)}
                onDeleted={handleMovementDeleted}
            />
        </Drawer>
    )
}
