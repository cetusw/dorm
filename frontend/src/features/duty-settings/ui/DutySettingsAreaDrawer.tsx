import { PencilIcon, PlusIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Drawer, FocusTrap, Group, ScrollArea, Stack, Text, Tooltip } from '@mantine/core'

import type { DutySettingsArea } from '../model/types'
import { formatDutySettingsAreaTitle } from '../model/utils'
import { EmptyState } from '../../../shared/ui/EmptyState'
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
            title={area ? (
                <Group gap={15} wrap="nowrap" align="center">
                    <Text size="lg" fw={700}>{formatDutySettingsAreaTitle(area)}</Text>
                    <Group gap={5} wrap="nowrap">
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
            ) : 'Территория'}
        >
            <FocusTrap.InitialFocus />

            {!area ? null : area.tasks.length === 0 ? (
                <Stack gap="md">
                    <EmptyState
                        title="Задачи не найдены"
                        description="В этой территории пока нет задач."
                    />
                    <button
                        type="button"
                        className={classes.addTaskLink}
                        onClick={() => onCreateTask(area.id)}
                    >
                        <PlusIcon size={25} />
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
                                <PlusIcon size={25} />
                                <Text className={classes.addTaskText}>Добавить задачу</Text>
                            </button>
                        </div>
                    </Stack>
                </ScrollArea>
            )}
        </Drawer>
    )
}
