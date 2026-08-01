import { useState } from 'react'

import { Alert, Button, Center, Loader } from '@mantine/core'

import { useSelectedDormitoryId } from '../../features/dormitories/model/useDormitorySelection'
import { useTaskCatalog } from '../../features/task-catalog/model/useTaskCatalog'
import type { TaskListItem } from '../../features/task-catalog/model/types'
import { DeleteTaskModal } from '../../features/task-catalog/ui/DeleteTaskModal'
import { TaskFormModal } from '../../features/task-catalog/ui/TaskFormModal'
import { TaskCatalogTable } from '../../features/task-catalog/ui/TaskCatalogTable'
import { ManagementPageFrame } from '../../shared/ui/ManagementPageFrame'

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
        <Button radius="md" onClick={handleCreate} disabled={!selectedDormitoryId}>
            + Создать задачу
        </Button>
    )

    let content = null

    if (!selectedDormitoryId) {
        content = (
            <Alert color="gray">
                Выберите общежитие в верхней панели.
            </Alert>
        )
    } else if (loading) {
        content = (
            <Center py="xl">
                <Loader />
            </Center>
        )
    } else if (error) {
        content = (
            <Alert color="red" title="Ошибка">
                {error}
            </Alert>
        )
    } else if (tasks.length === 0) {
        content = (
            <Alert color="gray">
                В выбранном общежитии пока нет задач.
            </Alert>
        )
    } else {
        content = (
            <TaskCatalogTable
                tasks={tasks}
                onEdit={handleEdit}
                onDelete={handleDelete}
            />
        )
    }

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
