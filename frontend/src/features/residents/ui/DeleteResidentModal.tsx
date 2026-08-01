import type { ResidentListItem } from '../model/types'
import { deleteResident } from '../api/residentsApi'
import { EntityDeleteModal } from '../../../shared/ui/EntityDeleteModal'

type Props = {
    opened: boolean
    resident: ResidentListItem | null
    onClose: () => void
    onDeleted: () => Promise<void> | void
}

export function DeleteResidentModal({ opened, resident, onClose, onDeleted }: Props) {
    return (
        <EntityDeleteModal
            opened={opened}
            onClose={onClose}
            title="Удаление жителя"
            entityLabel="жителя"
            entityName={resident?.name ?? null}
            errorMessage="Не удалось удалить жителя"
            onConfirm={async () => {
                if (!resident) {
                    return
                }

                await deleteResident(resident.id)
                await onDeleted()
            }}
        />
    )
}
