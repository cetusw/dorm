import { Text } from '@mantine/core'

import type { TaskActionHandler } from '../model/taskActions'
import type { TaskAreaGroup } from '../model/utils'
import { TaskMobileCard } from './TaskMobileCard'
import classes from './TaskMobileGroupSection.module.css'

type Props = {
    group: TaskAreaGroup
    isReadOnly?: boolean
    pendingTaskId: string | null
    onComplete: TaskActionHandler
    onOpen: TaskActionHandler
    onOpenDetails: (taskId: string) => void
}

export function TaskMobileGroupSection({
    group,
    isReadOnly = false,
    pendingTaskId,
    onComplete,
    onOpen,
    onOpenDetails,
}: Props) {
    return (
        <div className={classes.group}>
            <Text fw={700} size="lg" px={0} pt={0} pb={0}>
                {group.label}
            </Text>

            {group.tasks.map((task) => (
                <TaskMobileCard
                    key={task.id}
                    isReadOnly={isReadOnly}
                    pending={pendingTaskId === task.id}
                    task={task}
                    onComplete={onComplete}
                    onOpen={onOpen}
                    onOpenDetails={onOpenDetails}
                />
            ))}
        </div>
    )
}
