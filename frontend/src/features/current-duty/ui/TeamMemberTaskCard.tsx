import { Text } from '@mantine/core'

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
        <div className={classes.taskCard}>
            <div className={classes.taskCardContent}>
                <div className={classes.taskTitleCell}>
                    <Text className={classes.taskTitle}>{task.title}</Text>
                </div>

                <div className={classes.taskMetaRow}>
                    <TaskCostBadge cost={task.cost} />
                    <div className={classes.taskMetaRight}>
                        <DutyTaskStatusBadge
                            status={resolveTaskStatus(task)}
                            justify="flex-start"
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}
