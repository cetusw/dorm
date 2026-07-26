import { PencilIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Box, Group, Paper, Text, Tooltip } from '@mantine/core'

import type { DutySettingsTask } from '../model/types'
import { formatLastCompletedAt } from '../model/utils'
import classes from './DutySettingsTaskCard.module.css'

type Props = {
    task: DutySettingsTask
    onEdit: (taskId: string) => void
    onDelete: (taskId: string) => void
}

export function DutySettingsTaskCard({ task, onEdit, onDelete }: Props) {
    return (
        <Paper withBorder p="md" className={classes.card}>
            <div className={classes.content}>
                <div className={classes.titleCell}>
                    <Text fw={500} className={classes.title}>{task.title}</Text>
                </div>

                <div className={classes.scoreCell}>
                    <Box className={classes.pill}>
                        <Text size="sm">{task.cost} баллов</Text>
                    </Box>
                </div>

                <div className={classes.dateCell}>
                    <Tooltip label="Дата последнего выполнения">
                        <Text fw={500}>{formatLastCompletedAt(task.last_completed_at)}</Text>
                    </Tooltip>
                </div>

                <Group gap={4} wrap="nowrap" className={classes.actions} justify="flex-end">
                        <Tooltip label="Редактировать задачу">
                            <ActionIcon
                                variant="subtle"
                                color="gray"
                                aria-label="Редактировать задачу"
                                className={classes.iconButton}
                                onClick={() => onEdit(task.id)}
                            >
                                <PencilIcon size={25} />
                            </ActionIcon>
                        </Tooltip>

                        <Tooltip label="Удалить задачу">
                            <ActionIcon
                                variant="subtle"
                                color="red"
                                aria-label="Удалить задачу"
                                className={classes.iconButton}
                                onClick={() => onDelete(task.id)}
                            >
                                <TrashIcon size={25} />
                            </ActionIcon>
                        </Tooltip>
                </Group>
            </div>
        </Paper>
    )
}
