import { PlusIcon, PencilIcon, RowsPlusBottomIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Alert, Group, Stack, Text, Tooltip } from '@mantine/core'

import type { DutySettingsArea } from '../model/types'
import { formatDutySettingsAreaTitle } from '../model/utils'
import { DutySettingsTaskCard } from './DutySettingsTaskCard'
import classes from './DutySettingsAreaSection.module.css'

type Props = {
    area: DutySettingsArea
    onCreateArea: () => void
    onEditArea: (areaId: number) => void
    onDeleteArea: (areaId: number) => void
    onCreateTask: (areaId: number) => void
    onEditTask: (taskId: string) => void
    onDeleteTask: (taskId: string) => void
}

export function DutySettingsAreaSection({
    area,
    onCreateArea,
    onEditArea,
    onDeleteArea,
    onCreateTask,
    onEditTask,
    onDeleteTask,
}: Props) {
    return (
        <Stack gap="md" className={classes.section}>
            <Group justify="space-between" align="center" gap="md" wrap="wrap">
                <Group gap="sm" wrap="nowrap" className={classes.header}>
                    <Text size="xl" fw={500}>{formatDutySettingsAreaTitle(area)}</Text>

                    <Group gap={4} wrap="nowrap" className={`${classes.headerActions} ${classes.actions}`}>
                        <Tooltip label="Добавить территорию">
                            <ActionIcon
                                variant="subtle"
                                color="gray"
                                aria-label="Добавить территорию"
                                onClick={onCreateArea}
                            >
                                <PlusIcon size={25} />
                            </ActionIcon>
                        </Tooltip>

                        <Tooltip label="Редактировать территорию">
                            <ActionIcon
                                variant="subtle"
                                color="gray"
                                aria-label="Редактировать территорию"
                                onClick={() => onEditArea(area.id)}
                            >
                                <PencilIcon size={25} />
                            </ActionIcon>
                        </Tooltip>

                        <Tooltip label="Удалить территорию">
                            <ActionIcon
                                variant="subtle"
                                color="red"
                                aria-label="Удалить территорию"
                                onClick={() => onDeleteArea(area.id)}
                            >
                                <TrashIcon size={25} />
                            </ActionIcon>
                        </Tooltip>
                    </Group>
                </Group>
            </Group>

            {area.tasks.length === 0 ? (
                <Alert color="gray">В этой территории пока нет задач.</Alert>
            ) : (
                <Stack gap={0}>
                    {area.tasks.map((task) => (
                        <DutySettingsTaskCard
                            key={task.id}
                            task={task}
                            onEdit={onEditTask}
                            onDelete={onDeleteTask}
                        />
                    ))}
                </Stack>
            )}

            <div className={`${classes.addTaskWrap} ${classes.addTaskAction}`}>
                <button
                    type="button"
                    className={classes.addTaskLink}
                    onClick={() => onCreateTask(area.id)}
                >
                    <RowsPlusBottomIcon size={25} />
                    <Text className={classes.addTaskText}>Добавить задачу</Text>
                </button>
            </div>
        </Stack>
    )
}
