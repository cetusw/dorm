import { Paper, Text } from '@mantine/core'

import { DutyTaskStatusBadge, type DutyTaskStatus } from '../../../entities/duty-task'
import type { ResidentDutyTask } from '../model/types'
import { TaskCostBadge } from './TaskCostBadge'
import classes from './TeamMemberTaskGroups.module.css'

type Props = {
    task: ResidentDutyTask
}

function resolveTaskStatus(task: ResidentDutyTask): DutyTaskStatus {
    if (task.needs_revision) {
        return 'assigned'
    }

    return task.status
}

export function TeamMemberTaskCard({ task }: Props) {
    return (
        <Paper
            withBorder
            radius="lg"
            p={0}
            bg="var(--app-color-surface)"
            className={classes.taskCard}
        >
            <div className={classes.taskCardContent}>
                <Text className={classes.taskTitle}>{task.title}</Text>

                <div className={classes.taskMetaRow}>
                    <div className={classes.taskScoreCell}>
                        <TaskCostBadge cost={task.cost} />
                    </div>

                    <div className={classes.taskStatusCell}>
                        <DutyTaskStatusBadge
                            status={resolveTaskStatus(task)}
                            justify="flex-start"
                        />
                    </div>
                </div>
            </div>
        </Paper>
    )
}
