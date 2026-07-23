import { Stack, Text } from '@mantine/core'

import type { TaskAreaGroup } from '../model/utils'
import { TaskMobileCard } from './TaskMobileCard'
import type { TaskRowActionMode } from './TaskRowActions'
import classes from './TaskMobileGroupSection.module.css'

type Props = {
    activeSwipeTaskId: string | null
    actionMode?: TaskRowActionMode
    group: TaskAreaGroup
    isReadOnly?: boolean
    pendingTaskId: string | null
    onTake: (taskId: string) => void | Promise<unknown>
    onReturn: (taskId: string) => void | Promise<unknown>
    onComplete: (taskId: string) => void | Promise<unknown>
    onOpen: (taskId: string) => void | Promise<unknown>
    onReopen?: (taskId: string) => void | Promise<unknown>
    onSwipeActiveChange: (taskId: string | null) => void
    onVerify?: (taskId: string) => void | Promise<unknown>
}

export function TaskMobileGroupSection({
    activeSwipeTaskId,
    actionMode = 'default',
    group,
    isReadOnly = false,
    pendingTaskId,
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onReopen,
    onSwipeActiveChange,
    onVerify,
}: Props) {
    return (
        <div className={classes.group}>
            <Text fw={700} size="lg" px="xs" pt="xs" pb={2}>
                {group.label}
            </Text>

            <Stack gap={6}>
                {group.tasks.map((task) => (
                    <TaskMobileCard
                        key={task.id}
                        isReadOnly={isReadOnly}
                        mode={actionMode}
                        pending={pendingTaskId === task.id}
                        task={task}
                        activeSwipeTaskId={activeSwipeTaskId}
                        onTake={onTake}
                        onReturn={onReturn}
                        onComplete={onComplete}
                        onOpen={onOpen}
                        onReopen={onReopen}
                        onSwipeActiveChange={onSwipeActiveChange}
                        onVerify={onVerify}
                    />
                ))}
            </Stack>
        </div>
    )
}
