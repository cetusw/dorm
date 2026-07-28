import { RowsPlusBottomIcon } from '@phosphor-icons/react'
import { Stack, Text } from '@mantine/core'

import type { DutySettingsArea } from '../model/types'
import { formatDutySettingsAreaTitle } from '../model/utils'
import { DutySettingsTaskCard } from './DutySettingsTaskCard'
import classes from './DutySettingsAreaSection.module.css'

type Props = {
    area: DutySettingsArea
    onCreateTask: (areaId: number) => void
    onEditTask: (taskId: string) => void
    onIncludeTask: (taskId: string) => void
    onExcludeTask: (taskId: string) => void
    onDeleteTask: (taskId: string) => void
}

export function DutySettingsAreaSection({
    area,
    onCreateTask,
    onEditTask,
    onIncludeTask,
    onExcludeTask,
    onDeleteTask,
}: Props) {
    return (
        <Stack gap="md" className={classes.section}>
            <Text size="xl" fw={500}>{formatDutySettingsAreaTitle(area)}</Text>

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
