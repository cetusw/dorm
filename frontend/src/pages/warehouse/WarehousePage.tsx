import { useState } from 'react'

import { Alert, Center, Loader } from '@mantine/core'

import type { CurrentUser } from '../../features/current-user/model/types'
import { useWarehouse } from '../../features/warehouse/model/useWarehouse'
import { WarehouseCreateModal } from '../../features/warehouse/ui/WarehouseCreateModal'
import { DeleteWarehouseItemModal } from '../../features/warehouse/ui/DeleteWarehouseItemModal'
import { WarehouseHistoryDrawer } from '../../features/warehouse/ui/WarehouseHistoryDrawer'
import { WarehouseItemEditModal } from '../../features/warehouse/ui/WarehouseItemEditModal'
import { WarehouseMovementModal } from '../../features/warehouse/ui/WarehouseMovementModal'
import { WarehouseTable } from '../../features/warehouse/ui/WarehouseTable'
import { EmptyState } from '../../shared/ui/EmptyState'
import { ManagementPageFrame } from '../../shared/ui/ManagementPageFrame'
import { PageActionButton } from '../../shared/ui/PageActionButton'
import type { WarehouseItem } from '../../features/warehouse/model/types'

type Props = {
    currentUser: CurrentUser | null
}

export function WarehousePage({ currentUser }: Props) {
    const canManageWarehouse = currentUser?.can_manage_warehouse === true
    const { items, loading, error, syncInBackground, upsertItem, removeItem } = useWarehouse(canManageWarehouse)
    const [createModalOpened, setCreateModalOpened] = useState(false)
    const [historyItem, setHistoryItem] = useState<WarehouseItem | null>(null)
    const [addItem, setAddItem] = useState<WarehouseItem | null>(null)
    const [deleteItem, setDeleteItem] = useState<WarehouseItem | null>(null)
    const [editItem, setEditItem] = useState<WarehouseItem | null>(null)
    const [writeOffItem, setWriteOffItem] = useState<WarehouseItem | null>(null)

    async function handleWarehouseItemChanged(item: WarehouseItem) {
        upsertItem(item)
        setHistoryItem((currentItem) => (currentItem?.id === item.id ? item : currentItem))
        setAddItem((currentItem) => (currentItem?.id === item.id ? item : currentItem))
        setWriteOffItem((currentItem) => (currentItem?.id === item.id ? item : currentItem))
        setEditItem((currentItem) => (currentItem?.id === item.id ? item : currentItem))
        setDeleteItem((currentItem) => (currentItem?.id === item.id ? item : currentItem))

        await syncInBackground()
    }

    async function handleWarehouseItemDeleted(item: WarehouseItem) {
        removeItem(item.id)
        setHistoryItem((currentItem) => (currentItem?.id === item.id ? null : currentItem))
        setAddItem((currentItem) => (currentItem?.id === item.id ? null : currentItem))
        setWriteOffItem((currentItem) => (currentItem?.id === item.id ? null : currentItem))
        setEditItem((currentItem) => (currentItem?.id === item.id ? null : currentItem))
        setDeleteItem((currentItem) => (currentItem?.id === item.id ? null : currentItem))

        await syncInBackground()
    }

    const titleActions = (
        <PageActionButton
            disabled={!canManageWarehouse}
            onClick={() => setCreateModalOpened(true)}
        >
            Добавить на склад
        </PageActionButton>
    )

    const content = loading ? (
            <Center py="xl">
                <Loader />
            </Center>
        ) : error ? (
            <Alert color="red" title="Ошибка">
                {error}
            </Alert>
        ) : items.length === 0 ? (
            <EmptyState
                title="На складе пусто"
                description="Добавьте инвентарь на склад"
            />
        ) : (
            <WarehouseTable
                items={items}
                onOpenHistory={setHistoryItem}
                onAdd={setAddItem}
                onEdit={setEditItem}
                onDelete={setDeleteItem}
                onWriteOff={setWriteOffItem}
            />
        )

    return (
        <ManagementPageFrame title="Склад" titleActions={titleActions}>
            {content}

            <WarehouseCreateModal
                opened={createModalOpened}
                onClose={() => setCreateModalOpened(false)}
                onSaved={handleWarehouseItemChanged}
            />

            <WarehouseHistoryDrawer
                item={historyItem}
                opened={historyItem !== null}
                onClose={() => setHistoryItem(null)}
                onChanged={handleWarehouseItemChanged}
            />

            <WarehouseItemEditModal
                opened={editItem !== null}
                item={editItem}
                onClose={() => setEditItem(null)}
                onSaved={handleWarehouseItemChanged}
            />

            <DeleteWarehouseItemModal
                opened={deleteItem !== null}
                item={deleteItem}
                onClose={() => setDeleteItem(null)}
                onDeleted={handleWarehouseItemDeleted}
            />

            <WarehouseMovementModal
                mode="add"
                opened={addItem !== null}
                item={addItem}
                onClose={() => setAddItem(null)}
                onSaved={handleWarehouseItemChanged}
            />

            <WarehouseMovementModal
                mode="write-off"
                opened={writeOffItem !== null}
                item={writeOffItem}
                onClose={() => setWriteOffItem(null)}
                onSaved={handleWarehouseItemChanged}
            />
        </ManagementPageFrame>
    )
}
