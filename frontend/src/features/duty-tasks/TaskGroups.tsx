import { Alert, Stack } from '@mantine/core'

import { TaskGroupSection } from './TaskGroupSection'
import type { ResidentDutyTask } from './types'
import { groupTasksByArea } from './utils'

type Props = {
    pendingTaskId: string | null
    tasks: ResidentDutyTask[]
    onTake: (taskId: string) => void
    onReturn: (taskId: string) => void
    onComplete: (taskId: string) => void
    onOpen: (taskId: string) => void
}

export function TaskGroups({
    pendingTaskId,
    tasks,
    onTake,
    onReturn,
    onComplete,
    onOpen,
}: Props) {
    const groups = groupTasksByArea(tasks)

    if (groups.length === 0) {
        return <Alert color="gray">На текущее дежурство нет задач.</Alert>
    }

    return (
        <Stack gap="lg">
            {groups.map((group) => (
                <TaskGroupSection
                    key={group.key}
                    group={group}
                    pendingTaskId={pendingTaskId}
                    onTake={onTake}
                    onReturn={onReturn}
                    onComplete={onComplete}
                    onOpen={onOpen}
                />
            ))}
        </Stack>
    )
}
