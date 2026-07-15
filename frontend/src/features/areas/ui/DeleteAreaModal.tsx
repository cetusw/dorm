import { EntityDeleteModal } from '../../../shared/ui/EntityDeleteModal'
import { deleteArea } from '../api/areasApi'
import type { AreaListItem } from '../model/types'

type Props = {
    opened: boolean
    area: AreaListItem | null
    dormitoryId: string
    onClose: () => void
    onDeleted: () => Promise<void> | void
}

export function DeleteAreaModal({ opened, area, dormitoryId, onClose, onDeleted }: Props) {
    return (
        <EntityDeleteModal
            opened={opened}
            onClose={onClose}
            title="Удаление территории"
            entityLabel="территорию"
            entityName={area?.name ?? null}
            errorMessage="Не удалось удалить территорию"
            onConfirm={async () => {
                if (!area) {
                    return
                }

                await deleteArea(dormitoryId, area.id)
                await onDeleted()
            }}
        />
    )
}
