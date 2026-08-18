import { ConfirmActionModal } from '../../../shared/ui/ConfirmActionModal'
import { deleteWarehouseItem } from '../api/warehouseApi'
import type { WarehouseItem } from '../model/types'

type Props = {
    opened: boolean
    item: WarehouseItem | null
    onClose: () => void
    onDeleted: (item: WarehouseItem) => Promise<void> | void
}

export function DeleteWarehouseItemModal({ opened, item, onClose, onDeleted }: Props) {
    return (
        <ConfirmActionModal
            opened={opened}
            onClose={onClose}
            title="Удаление элемента учёта"
            description={item
                ? `Вы уверены, что хотите удалить ${item.name}. Это приведёт к полной потере истории элемента учёта`
                : 'Вы уверены, что хотите удалить элемент учёта?'}
            confirmLabel="Удалить"
            confirmColor="red"
            errorMessage="Не удалось удалить элемент учёта"
            onConfirm={async () => {
                if (!item) {
                    return
                }

                await deleteWarehouseItem(item.id)
                await onDeleted(item)
            }}
        />
    )
}
