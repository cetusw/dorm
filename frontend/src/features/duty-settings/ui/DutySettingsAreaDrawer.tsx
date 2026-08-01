import { PencilIcon, PlusIcon, RowsPlusBottomIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Alert, Drawer, FocusTrap, Group, ScrollArea, Stack, Text } from '@mantine/core'

import type { DutySettingsArea } from '../model/types'
import { formatDutySettingsAreaTitle } from '../model/utils'
import { DutySettingsTaskCard } from './DutySettingsTaskCard'
import classes from './DutySettingsAreaDrawer.module.css'

type Props = {
    area: DutySettingsArea | null
    opened: boolean
    hasNestedModalOpen?: boolean
    onClose: () => void
    onCreateArea: () => void
    onEditArea: (areaId: number) => void
    onDeleteArea: (areaId: number) => void
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
        <Drawer
            opened={opened}
            onClose={onClose}
            closeOnEscape={!hasNestedModalOpen}
            closeOnClickOutside={!hasNestedModalOpen}
            position="right"
            size="50vw"
            zIndex={100}
            title={area ? (
                <Group gap={15} wrap="nowrap" align="center">
                    <Text size="xl" fw={500}>{formatDutySettingsAreaTitle(area)}</Text>
                    <Group gap={4} wrap="nowrap">
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
            ) : 'Территория'}
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
