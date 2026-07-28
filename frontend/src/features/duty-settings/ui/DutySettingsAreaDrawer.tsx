import { RowsPlusBottomIcon } from '@phosphor-icons/react'
import { Alert, Drawer, FocusTrap, ScrollArea, Stack, Text } from '@mantine/core'

import type { DutySettingsArea } from '../model/types'
import { formatDutySettingsAreaTitle } from '../model/utils'
import { DutySettingsTaskCard } from './DutySettingsTaskCard'
import classes from './DutySettingsAreaDrawer.module.css'

type Props = {
    area: DutySettingsArea | null
    opened: boolean
    hasNestedModalOpen?: boolean
    onClose: () => void
    onCreateTask: (areaId: number) => void
    onEditTask: (taskId: string) => void
    onIncludeTask: (taskId: string) => void
    onExcludeTask: (taskId: string) => void
    onDeleteTask: (taskId: string) => void
}

export function DutySettingsAreaDrawer({
    area,
    opened,
    hasNestedModalOpen = false,
    onClose,
    onCreateTask,
    onEditTask,
    onIncludeTask,
    onExcludeTask,
    onDeleteTask,
}: Props) {
    return (
        <Drawer
            opened={opened}
            onClose={onClose}
            closeOnEscape={!hasNestedModalOpen}
            closeOnClickOutside={!hasNestedModalOpen}
            position="right"
            size="50vw"
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
                                onInclude={onIncludeTask}
                                onExclude={onExcludeTask}
                                onDelete={onDeleteTask}
                            />
                        ))}

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
                </ScrollArea>
            )}
        </Drawer>
    )
}
