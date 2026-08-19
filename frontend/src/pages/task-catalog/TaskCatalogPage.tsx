import { useState } from 'react'

import { Alert, Center, Loader } from '@mantine/core'

import { useSelectedDormitoryId } from '../../features/dormitories/model/useDormitorySelection'
import { useTaskCatalog } from '../../features/task-catalog/model/useTaskCatalog'
import type { TaskListItem } from '../../features/task-catalog/model/types'
import { DeleteTaskModal } from '../../features/task-catalog/ui/DeleteTaskModal'
import { TaskFormModal } from '../../features/task-catalog/ui/TaskFormModal'
import { TaskCatalogTable } from '../../features/task-catalog/ui/TaskCatalogTable'
import { EmptyState } from '../../shared/ui/EmptyState'
import { ManagementPageFrame } from '../../shared/ui/ManagementPageFrame'
import { PageActionButton } from '../../shared/ui/PageActionButton'

export function TaskCatalogPage() {
    const selectedDormitoryId = useSelectedDormitoryId()
    const { tasks, loading, error, reload } = useTaskCatalog(selectedDormitoryId)
    const [formOpened, setFormOpened] = useState(false)
    const [editingTaskId, setEditingTaskId] = useState<string | null>(null)
    const [deletingTask, setDeletingTask] = useState<TaskListItem | null>(null)

    function handleCreate() {
        setEditingTaskId(null)
        setFormOpened(true)
    }

    function handleEdit(taskId: string) {
        setEditingTaskId(taskId)
        setFormOpened(true)
    }

    function handleDelete(task: TaskListItem) {
        setDeletingTask(task)
    }

    const titleActions = (
        <PageActionButton onClick={handleCreate} disabled={!selectedDormitoryId}>
            Создать задачу
        </PageActionButton>
    )

    const content = !selectedDormitoryId ? (
            <EmptyState
                title="Общежитие не выбрано"
                description="Выберите общежитие в верхней панели, чтобы посмотреть список задач."
            />
        ) : loading ? (
            <Center py="xl">
                <Loader />
            </Center>
        ) : error ? (
            <Alert color="red" title="Ошибка">
                {error}
            </Alert>
        ) : tasks.length === 0 ? (
            <EmptyState
                title="Задачи не найдены"
                description="В выбранном общежитии пока нет задач."
            />
        ) : (
            <TaskCatalogTable
                tasks={tasks}
                onEdit={handleEdit}
                onDelete={handleDelete}
            />
        )

    return (
        <ManagementPageFrame title="Задачи" titleActions={titleActions}>
            {content}

            {selectedDormitoryId && (
                <>
                    <TaskFormModal
                        opened={formOpened}
                        mode={editingTaskId === null ? 'create' : 'edit'}
                        taskId={editingTaskId}
                        dormitoryId={selectedDormitoryId}
                        onClose={() => {
                            setFormOpened(false)
                            setEditingTaskId(null)
                        }}
                        onSaved={reload}
                    />

                    <DeleteTaskModal
                        opened={deletingTask !== null}
                        task={deletingTask}
                        dormitoryId={selectedDormitoryId}
                        onClose={() => setDeletingTask(null)}
                        onDeleted={reload}
                    />
                </>
            )}
        </ManagementPageFrame>
    )
}
