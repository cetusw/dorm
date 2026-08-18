import { ConfirmActionModal } from '../../../shared/ui/ConfirmActionModal'
import { deleteWarehouseMovement } from '../api/warehouseApi'
import type { WarehouseItem, WarehouseMovement } from '../model/types'

type Props = {
    opened: boolean
    item: WarehouseItem | null
    movement: WarehouseMovement | null
    onClose: () => void
    onDeleted: () => Promise<void> | void
}

export function DeleteWarehouseMovementModal({ opened, item, movement, onClose, onDeleted }: Props) {
    const movementLabel = movement?.type === 'write-off' ? 'списание' : 'поступление'

    return (
        <ConfirmActionModal
            opened={opened}
            onClose={onClose}
            title="Удаление движения"
            description={`Вы уверены, что хотите удалить ${movementLabel}?`}
            confirmLabel="Удалить"
            confirmColor="red"
            errorMessage="Не удалось удалить движение склада"
            onConfirm={async () => {
                if (!item || !movement) {
                    return
                }

                await deleteWarehouseMovement(item.id, movement.id)
                await onDeleted()
            }}
        />
    )
}
