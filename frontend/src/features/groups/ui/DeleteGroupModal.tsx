import { EntityDeleteModal } from '../../../shared/ui/EntityDeleteModal'
import { deleteGroup } from '../api/groupsApi'
import type { GroupListItem } from '../model/types'

type Props = {
    opened: boolean
    group: GroupListItem | null
    onClose: () => void
    onDeleted: () => Promise<void> | void
}

export function DeleteGroupModal({ opened, group, onClose, onDeleted }: Props) {
    return (
        <EntityDeleteModal
            opened={opened}
            onClose={onClose}
            title="Удаление группы"
            entityLabel="группу"
            entityName={group?.name ?? null}
            errorMessage="Не удалось удалить группу"
            onConfirm={async () => {
                if (!group) {
                    return
                }

                await deleteGroup(group.id)
                await onDeleted()
            }}
        />
    )
}
