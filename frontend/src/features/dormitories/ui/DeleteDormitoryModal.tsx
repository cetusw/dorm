import { deleteDormitory } from '../api/dormitoriesApi'
import type { DormitoryListItem } from '../model/types'
import { EntityDeleteModal } from '../../../shared/ui/EntityDeleteModal'

type Props = {
    opened: boolean
    dormitory: DormitoryListItem | null
    onClose: () => void
    onDeleted: () => Promise<void> | void
}

export function DeleteDormitoryModal({ opened, dormitory, onClose, onDeleted }: Props) {
    return (
        <EntityDeleteModal
            opened={opened}
            onClose={onClose}
            title="Удаление общежития"
            entityLabel="общежитие"
            entityName={dormitory?.name ?? null}
            errorMessage="Не удалось удалить общежитие"
            onConfirm={async () => {
                if (!dormitory) {
                    return
                }

                await deleteDormitory(dormitory.id)
                await onDeleted()
            }}
        />
    )
}
