import { Alert, Stack } from '@mantine/core'

import type { ResidentDutyTask } from '../model/types'
import { groupTasksByArea } from '../model/utils'
import { TaskGroupSection } from './TaskGroupSection'
import { TaskMobileGroupSection } from './TaskMobileGroupSection'
import type { TaskRowActionMode } from './TaskRowActions'

type Props = {
    actionMode?: TaskRowActionMode
    isReadOnly?: boolean
    pendingTaskId: string | null
    tasks: ResidentDutyTask[]
    emptyMessage?: string
    showAssigneeColumn?: boolean
    onTake: (taskId: string) => void | Promise<unknown>
    onReturn: (taskId: string) => void | Promise<unknown>
    onComplete: (taskId: string) => void | Promise<unknown>
    onOpen: (taskId: string) => void | Promise<unknown>
    onVerify?: (taskId: string) => void | Promise<unknown>
}

export function TaskGroups({
    actionMode = 'default',
    isReadOnly = false,
    pendingTaskId,
    tasks,
    emptyMessage = 'В этом разделе нет задач.',
    showAssigneeColumn = true,
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onVerify,
}: Props) {
    const groups = groupTasksByArea(tasks)

    if (groups.length === 0) {
        return <Alert color="gray">{emptyMessage}</Alert>
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
                        pendingTaskId={pendingTaskId}
                        onTake={onTake}
                        onReturn={onReturn}
                        onComplete={onComplete}
                        onOpen={onOpen}
                        onVerify={onVerify}
                    />
                ))}
            </Stack>

            <Stack visibleFrom="md" gap="lg">
                {groups.map((group) => (
                    <TaskGroupSection
                        actionMode={actionMode}
                        isReadOnly={isReadOnly}
                        key={group.key}
                        group={group}
                        pendingTaskId={pendingTaskId}
                        showAssigneeColumn={showAssigneeColumn}
                        onTake={onTake}
                        onReturn={onReturn}
                        onComplete={onComplete}
                        onOpen={onOpen}
                        onVerify={onVerify}
                    />
                ))}
            </Stack>
        </>
    )
}
