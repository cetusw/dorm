import { PencilIcon, PlusIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Group, Stack, Text, Tooltip } from '@mantine/core'

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
    onIncludeTask: (taskId: string) => void
    onExcludeTask: (taskId: string) => void
    onDeleteTask: (taskId: string) => void
}

export function DutySettingsAreaSection({
    area,
    onCreateArea,
    onEditArea,
    onDeleteArea,
    onCreateTask,
    onEditTask,
    onIncludeTask,
    onExcludeTask,
    onDeleteTask,
}: Props) {
    return (
        <Stack gap={0} className={classes.section}>
            <Group gap={15} wrap="nowrap" align="center" className={classes.header}>
                <Text size="lg" fw={700}>{formatDutySettingsAreaTitle(area)}</Text>
                <Group gap={5} wrap="nowrap" className={classes.headerActions}>
                    <Tooltip label="Добавить территорию">
                        <ActionIcon size={30} radius="md" variant="subtle" className={classes.iconButton} aria-label="Добавить территорию" onClick={onCreateArea}>
                            <PlusIcon size={20} />
                        </ActionIcon>
                    </Tooltip>
                    <Tooltip label="Редактировать территорию">
                        <ActionIcon size={30} radius="md" variant="subtle" className={classes.iconButton} aria-label="Редактировать территорию" onClick={() => onEditArea(area.id)}>
                            <PencilIcon size={20} />
                        </ActionIcon>
                    </Tooltip>
                    <Tooltip label="Удалить территорию">
                        <ActionIcon size={30} radius="md" variant="subtle" className={classes.iconButton} aria-label="Удалить территорию" onClick={() => onDeleteArea(area.id)}>
                            <TrashIcon size={20} />
                        </ActionIcon>
                    </Tooltip>
                </Group>
            </Group>

            <Stack gap={0}>
                {area.tasks.map((task) => (
                    <DutySettingsTaskCard
                        key={task.id}
                        task={task}
                        onEdit={onEditTask}
                        onInclude={onIncludeTask}
                        onExclude={onExcludeTask}
                        onDelete={onDeleteTask}
                    />
                ))}
            </Stack>

            <div className={classes.addTaskWrap}>
                <button
                    type="button"
                    className={classes.addTaskLink}
                    onClick={() => onCreateTask(area.id)}
                >
                    <PlusIcon size={25} />
                    <Text className={classes.addTaskText}>Добавить задачу</Text>
                </button>
            </div>
        </Stack>
    )
}
