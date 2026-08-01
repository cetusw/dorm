import { EntityDeleteModal } from '../../../shared/ui/EntityDeleteModal'
import { deleteArea } from '../api/areasApi'
import type { AreaListItem } from '../model/types'

type Props = {
    opened: boolean
    area: AreaListItem | null
    dormitoryId: string
    deleteAreaRequest?: (areaId: number) => Promise<void>
    onClose: () => void
    onDeleted: () => Promise<void> | void
}

export function DeleteAreaModal({ opened, area, dormitoryId, deleteAreaRequest, onClose, onDeleted }: Props) {
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

                await (deleteAreaRequest ?? ((areaId) => deleteArea(dormitoryId, areaId)))(area.id)
                await onDeleted()
            }}
        />
    )
}
