import { PencilIcon, PlusIcon, RowsPlusBottomIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Group, Stack, Text } from '@mantine/core'

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
        <Stack gap="md" className={classes.section}>
            <Group gap={15} wrap="nowrap" align="center" className={classes.header}>
                <Text size="xl" fw={500}>{formatDutySettingsAreaTitle(area)}</Text>
                <Group gap={4} wrap="nowrap" className={classes.headerActions}>
                    <ActionIcon variant="subtle" color="gray" size={25} aria-label="Добавить территорию" onClick={onCreateArea}>
                        <PlusIcon size={25} />
                    </ActionIcon>
                    <ActionIcon variant="subtle" color="gray" size={25} aria-label="Редактировать территорию" onClick={() => onEditArea(area.id)}>
                        <PencilIcon size={25} />
                    </ActionIcon>
                    <ActionIcon variant="subtle" color="red" size={25} aria-label="Удалить территорию" onClick={() => onDeleteArea(area.id)}>
                        <TrashIcon size={25} />
                    </ActionIcon>
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
                    <RowsPlusBottomIcon size={25} />
                    <Text className={classes.addTaskText}>Добавить задачу</Text>
                </button>
            </div>
        </Stack>
    )
}
