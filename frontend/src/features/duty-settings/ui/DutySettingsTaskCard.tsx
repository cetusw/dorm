import {
    BackspaceIcon,
    CheckIcon,
    PencilIcon,
    TrashIcon,
} from '@phosphor-icons/react'
import { ActionIcon, Text, Tooltip } from '@mantine/core'

import { DutyTaskStatusBadge, type DutyTaskStatus } from '../../../entities/duty-task'
import { SettingsBadge } from '../../../shared/ui/SettingsBadge'
import { TaskCostBadge } from '../../current-duty/ui/TaskCostBadge'
import { formatDutySettingsRecurrence } from '../model/utils'
import type { DutySettingsTask } from '../model/types'
import classes from './DutySettingsTaskCard.module.css'

type Props = {
    task: DutySettingsTask
    onEdit: (taskId: string) => void
    onInclude: (taskId: string) => void
    onExclude: (taskId: string) => void
    onDelete: (taskId: string) => void
}

export function DutySettingsTaskCard({ task, onEdit, onInclude, onExclude, onDelete }: Props) {
    const badgeStatus: DutyTaskStatus = task.status === '' ? 'free' : task.status
    const canExclude = task.is_included && (task.status === 'free' || task.status === 'assigned' || task.status === '')
    const shouldShowStatus = task.is_included && (task.status === 'completed' || task.status === 'verified')
    const shouldShowAssignee = task.is_included && Boolean(task.assignee_name)

    return (
        <article className={classes.cardRoot} data-muted={!task.is_included ? 'true' : undefined}>
            <div className={classes.content}>
                <div className={classes.titleCell}>
                    <Tooltip label={task.title}>
                        <Text fw={500} className={classes.title}>{task.title}</Text>
                    </Tooltip>

                    <div className={classes.actions}>
                        {task.is_included ? (
                            canExclude ? (
                                <Tooltip label="Исключить из дежурства">
                                    <ActionIcon
                                        size={30}
                                        radius="md"
                                        variant="subtle"
                                        aria-label={`Исключить задачу ${task.title} из дежурства`}
                                        className={classes.iconButton}
                                        onClick={() => onExclude(task.id)}
                                    >
                                        <BackspaceIcon size={20} />
                                    </ActionIcon>
                                </Tooltip>
                            ) : null
                        ) : (
                            <Tooltip label="Включить в дежурство">
                                <ActionIcon
                                    size={30}
                                    radius="md"
                                    variant="subtle"
                                    aria-label={`Включить задачу ${task.title} в дежурство`}
                                    className={classes.iconButton}
                                    onClick={() => onInclude(task.id)}
                                >
                                    <CheckIcon size={20} />
                                </ActionIcon>
                            </Tooltip>
                        )}

                        <Tooltip label="Редактировать">
                            <ActionIcon
                                size={30}
                                radius="md"
                                variant="subtle"
                                aria-label={`Редактировать задачу ${task.title}`}
                                className={classes.iconButton}
                                onClick={() => onEdit(task.id)}
                            >
                                <PencilIcon size={20} />
                            </ActionIcon>
                        </Tooltip>

                        <Tooltip label="Удалить">
                            <ActionIcon
                                size={30}
                                radius="md"
                                variant="subtle"
                                aria-label={`Удалить задачу ${task.title}`}
                                className={classes.iconButton}
                                onClick={() => onDelete(task.id)}
                            >
                                <TrashIcon size={20} />
                            </ActionIcon>
                        </Tooltip>
                    </div>
                </div>

                <div className={classes.metaRow}>
                    <TaskCostBadge cost={task.cost} />
                    <SettingsBadge color="taskType">{formatDutySettingsRecurrence(task)}</SettingsBadge>
                    {shouldShowStatus ? <DutyTaskStatusBadge status={badgeStatus} justify="flex-start" /> : null}
                    {shouldShowAssignee ? <Text className={classes.assigneeText}>{task.assignee_name}</Text> : null}
                </div>
            </div>
        </article>
    )
}
