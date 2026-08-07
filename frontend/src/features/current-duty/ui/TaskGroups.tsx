import { useEffect, useState } from 'react'

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
    const groups = groupTasksByArea(tasks)
    const [activeMobileSwipeTaskId, setActiveMobileSwipeTaskId] = useState<string | null>(null)
    const [selectedMobileTaskId, setSelectedMobileTaskId] = useState<string | null>(null)
    const selectedMobileTask = selectedMobileTaskId
        ? tasks.find((task) => task.id === selectedMobileTaskId) ?? null
        : null

    useEffect(() => {
        if (!activeMobileSwipeTaskId) {
            return
        }

        function handleScroll() {
            setActiveMobileSwipeTaskId(null)
        }

        window.addEventListener('scroll', handleScroll, { passive: true })

        return () => {
            window.removeEventListener('scroll', handleScroll)
        }
    }, [activeMobileSwipeTaskId])

    useEffect(() => {
        if (selectedMobileTaskId && !selectedMobileTask) {
            setSelectedMobileTaskId(null)
        }
    }, [selectedMobileTask, selectedMobileTaskId])

    function handleOpenMobileDetails(taskId: string) {
        setActiveMobileSwipeTaskId(null)
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
                        actionMode={actionMode}
                        isReadOnly={isReadOnly}
                        key={group.key}
                        group={group}
                        activeSwipeTaskId={activeMobileSwipeTaskId}
                        pendingTaskId={pendingTaskId}
                        onTake={onTake}
                        onReturn={onReturn}
                        onComplete={onComplete}
                        onOpen={onOpen}
                        onOpenDetails={handleOpenMobileDetails}
                        onReopen={onReopen}
                        onVerify={onVerify}
                        onSwipeActiveChange={setActiveMobileSwipeTaskId}
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

            <Stack visibleFrom="md" gap="lg">
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
