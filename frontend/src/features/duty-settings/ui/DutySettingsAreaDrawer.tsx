import { RowsPlusBottomIcon } from '@phosphor-icons/react'
import { Alert, Drawer, FocusTrap, ScrollArea, Stack, Text } from '@mantine/core'

import type { DutySettingsArea } from '../model/types'
import { formatDutySettingsAreaTitle } from '../model/utils'
import { DutySettingsTaskCard } from './DutySettingsTaskCard'
import classes from './DutySettingsAreaDrawer.module.css'

type Props = {
    area: DutySettingsArea | null
    opened: boolean
    onClose: () => void
    onCreateTask: (areaId: number) => void
    onEditTask: (taskId: string) => void
    onDeleteTask: (taskId: string) => void
}

export function DutySettingsAreaDrawer({
    area,
    opened,
    onClose,
    onCreateTask,
    onEditTask,
    onDeleteTask,
}: Props) {
    return (
        <Drawer
            opened={opened}
            onClose={onClose}
            position="right"
            size={860}
            zIndex={100}
            title={area ? formatDutySettingsAreaTitle(area) : 'Территория'}
        >
            <FocusTrap.InitialFocus />

            {!area ? null : area.tasks.length === 0 ? (
                <Stack gap="md">
                    <Alert color="gray">В этой территории пока нет задач.</Alert>
                    <button
                        type="button"
                        className={classes.addTaskLink}
                        onClick={() => onCreateTask(area.id)}
                    >
                        <RowsPlusBottomIcon size={25} />
                        <Text className={classes.addTaskText}>Добавить задачу</Text>
                    </button>
                </Stack>
            ) : (
                <ScrollArea>
                    <Stack gap={0} pb={4}>
                        {area.tasks.map((task) => (
                            <DutySettingsTaskCard
                                key={task.id}
                                task={task}
                                onEdit={onEditTask}
                                onDelete={onDeleteTask}
                            />
                        ))}

                        <button
                            type="button"
                            className={classes.addTaskLink}
                            onClick={() => onCreateTask(area.id)}
                        >
                            <RowsPlusBottomIcon size={25} />
                            <Text className={classes.addTaskText}>Добавить задачу</Text>
                        </button>
                    </Stack>
                </ScrollArea>
            )}
        </Drawer>
    )
}
