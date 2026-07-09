import { Alert, Stack } from '@mantine/core'

import { TaskGroupSection } from './TaskGroupSection'
import type { TaskRowActionMode } from './TaskRowActions'
import type { ResidentDutyTask } from './types'
import { groupTasksByArea } from './utils'

type Props = {
    actionMode?: TaskRowActionMode
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
        <Stack gap="lg">
            {groups.map((group) => (
                <TaskGroupSection
                    actionMode={actionMode}
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
    )
}
