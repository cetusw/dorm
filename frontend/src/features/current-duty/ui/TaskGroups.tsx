import { useEffect, useMemo, useState } from 'react'

import { Stack } from '@mantine/core'

import type { TaskActionHandler, TaskRowActionMode } from '../model/taskActions'
import type { ResidentDutyTask } from '../model/types'
import { EmptyState } from '../../../shared/ui/EmptyState'
import { groupTasksByArea } from '../model/utils'
import { TaskGroupSection } from './TaskGroupSection'
import { TaskMobileGroupSection } from './TaskMobileGroupSection'
import { TaskMobileDetailsDrawer } from './TaskMobileDetailsDrawer'

type Props = {
    actionMode?: TaskRowActionMode
    isReadOnly?: boolean
    pendingTaskId: string | null
    tasks: ResidentDutyTask[]
    emptyMessage?: string
    onTake: TaskActionHandler
    onReturn: TaskActionHandler
    onComplete: TaskActionHandler
    onOpen: TaskActionHandler
    onReopen?: TaskActionHandler
    onVerify?: TaskActionHandler
}

export function TaskGroups({
    actionMode = 'default',
    isReadOnly = false,
    pendingTaskId,
    tasks,
    emptyMessage = 'В этом разделе нет задач.',
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onReopen,
    onVerify,
}: Props) {
    const groups = useMemo(() => groupTasksByArea(tasks), [tasks])
    const [selectedMobileTaskId, setSelectedMobileTaskId] = useState<string | null>(null)
    const selectedMobileTask = useMemo(
        () => (selectedMobileTaskId ? tasks.find((task) => task.id === selectedMobileTaskId) ?? null : null),
        [selectedMobileTaskId, tasks],
    )

    useEffect(() => {
        if (selectedMobileTaskId && !selectedMobileTask) {
            setSelectedMobileTaskId(null)
        }
    }, [selectedMobileTask, selectedMobileTaskId])

    function handleOpenMobileDetails(taskId: string) {
        setSelectedMobileTaskId(taskId)
    }

    if (groups.length === 0) {
        return (
            <EmptyState
                title="Задачи не найдены"
                description={emptyMessage}
            />
        )
    }

    return (
        <>
            <Stack hiddenFrom="md" gap="sm">
                {groups.map((group) => (
                    <TaskMobileGroupSection
                        isReadOnly={isReadOnly}
                        key={group.key}
                        group={group}
                        pendingTaskId={pendingTaskId}
                        onComplete={onComplete}
                        onOpen={onOpen}
                        onOpenDetails={handleOpenMobileDetails}
                    />
                ))}
            </Stack>

            <TaskMobileDetailsDrawer
                isReadOnly={isReadOnly}
                mode={actionMode}
                opened={selectedMobileTask !== null}
                pendingTaskId={pendingTaskId}
                task={selectedMobileTask}
                onClose={() => setSelectedMobileTaskId(null)}
                onTake={onTake}
                onReturn={onReturn}
                onReopen={onReopen ?? (async () => false)}
                onVerify={onVerify ?? (async () => false)}
            />

            <Stack visibleFrom="md" gap={30}>
                {groups.map((group) => (
                    <TaskGroupSection
                        actionMode={actionMode}
                        isReadOnly={isReadOnly}
                        key={group.key}
                        group={group}
                        pendingTaskId={pendingTaskId}
                        onTake={onTake}
                        onReturn={onReturn}
                        onComplete={onComplete}
                        onOpen={onOpen}
                        onReopen={onReopen}
                        onVerify={onVerify}
                    />
                ))}
            </Stack>
        </>
    )
}
