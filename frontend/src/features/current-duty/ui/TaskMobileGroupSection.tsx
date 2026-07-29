import { Stack, Text } from '@mantine/core'

import type { TaskActionHandler, TaskRowActionMode } from '../model/taskActions'
import type { TaskAreaGroup } from '../model/utils'
import { TaskMobileCard } from './TaskMobileCard'
import classes from './TaskMobileGroupSection.module.css'

type Props = {
    activeSwipeTaskId: string | null
    actionMode?: TaskRowActionMode
    group: TaskAreaGroup
    isReadOnly?: boolean
    pendingTaskId: string | null
    onTake: TaskActionHandler
    onReturn: TaskActionHandler
    onComplete: TaskActionHandler
    onOpen: TaskActionHandler
    onOpenDetails: (taskId: string) => void
    onReopen?: TaskActionHandler
    onSwipeActiveChange: (taskId: string | null) => void
    onVerify?: TaskActionHandler
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
    onOpenDetails,
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
                        onOpenDetails={onOpenDetails}
                        onReopen={onReopen}
                        onSwipeActiveChange={onSwipeActiveChange}
                        onVerify={onVerify}
                    />
                ))}
            </Stack>
        </div>
    )
}
