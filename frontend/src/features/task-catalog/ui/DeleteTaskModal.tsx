import { EntityDeleteModal } from '../../../shared/ui/EntityDeleteModal'
import { deleteTaskDefinition } from '../api/taskCatalogApi'
import type { TaskListItem } from '../model/types'

type Props = {
    opened: boolean
    task: TaskListItem | null
    dormitoryId: string
    onClose: () => void
    onDeleted: () => Promise<void> | void
}

export function DeleteTaskModal({
    opened,
    task,
    dormitoryId,
    onClose,
    onDeleted,
}: Props) {
    return (
        <EntityDeleteModal
            opened={opened}
            title="Удаление задачи"
            entityLabel="задачу"
            entityName={task?.title ?? null}
            submitLabel="Удалить"
            onClose={onClose}
            onConfirm={async () => {
                if (!task) {
                    return
                }

                await deleteTaskDefinition(dormitoryId, task.id)
                await onDeleted()
            }}
            errorMessage="Не удалось удалить задачу"
        />
    )
}
