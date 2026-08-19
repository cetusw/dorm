import { Checkbox, Paper, Text } from '@mantine/core'

import { DutyTaskStatusBadge } from '../../../entities/duty-task'
import { getTaskCardPresentation } from '../model/taskCardPresentation'
import type { TaskActionHandler } from '../model/taskActions'
import type { ResidentDutyTask } from '../model/types'
import { TaskCostBadge } from './TaskCostBadge'
import classes from './TaskMobileCard.module.css'

type Props = {
    isReadOnly?: boolean
    pending: boolean
    task: ResidentDutyTask
    onComplete: TaskActionHandler
    onOpen: TaskActionHandler
    onOpenDetails: (taskId: string) => void
}

export function TaskMobileCard({
    isReadOnly = false,
    pending,
    task,
    onComplete,
    onOpen,
    onOpenDetails,
}: Props) {
    const presentation = getTaskCardPresentation({
        isReadOnly,
        task,
    })
    const showCheckbox = presentation.showCheckbox
    const isCheckboxInteractive = !isReadOnly && !pending && task.is_mine && (task.can_complete || task.can_open)
    const isCheckboxChecked = task.status === 'completed' || task.status === 'verified'

    function handleCheckboxToggle() {
        if (!isCheckboxInteractive) {
            return
        }

        if (isCheckboxChecked) {
            void onOpen(task.id)
            return
        }

        void onComplete(task.id)
    }

    return (
        <div className={classes.cardRoot}>
            <Paper
                withBorder={false}
                radius={0}
                p={0}
                bg="transparent"
                className={classes.surface}
                onClick={() => onOpenDetails(task.id)}
            >
                <div className={classes.headerRow}>
                    {showCheckbox && (
                        <div
                            className={classes.checkboxWrap}
                            onClick={(event) => {
                                event.stopPropagation()
                                handleCheckboxToggle()
                            }}
                        >
                            <Checkbox
                                checked={isCheckboxChecked}
                                disabled={pending}
                                readOnly={!isCheckboxInteractive}
                                size="25px"
                                radius="xl"
                                iconColor="#FFFFFF"
                                styles={{
                                    input: isCheckboxChecked
                                        ? {
                                            backgroundColor: '#8C8C8C',
                                            borderColor: '#8C8C8C',
                                        }
                                        : undefined,
                                }}
                                aria-label={`${isCheckboxChecked ? 'Отменить выполнение' : 'Выполнить'} задачу ${task.title}`}
                                onClick={(event) => {
                                    event.stopPropagation()
                                }}
                                onChange={(event) => {
                                    event.stopPropagation()
                                    handleCheckboxToggle()
                                }}
                            />
                        </div>
                    )}

                    <div className={classes.content}>
                        <div className={classes.titleWrap}>
                            <Text fw={400} className={classes.title}>
                                {task.title}
                            </Text>
                        </div>

                        <div className={classes.metaRow}>
                            <TaskCostBadge cost={task.cost} />
                            {presentation.showStatus && <DutyTaskStatusBadge status={task.status} justify="flex-start" />}
                            {presentation.showAssignee && presentation.assigneeLabel && (
                                <Text className={classes.assigneeText}>
                                    {presentation.assigneeLabel}
                                </Text>
                            )}
                        </div>
                    </div>
                </div>
            </Paper>
        </div>
    )
}
