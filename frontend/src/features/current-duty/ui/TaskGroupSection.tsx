import { Divider, Paper, ScrollArea, Table, Text, Tooltip } from '@mantine/core'

import type { ResidentDutyTask } from '../model/types'
import type { TaskAreaGroup } from '../model/utils'
import { TaskRowActions, type TaskRowActionMode } from './TaskRowActions'
import { TaskStatusBadge } from './TaskStatusBadge'

type Props = {
    actionMode?: TaskRowActionMode
    group: TaskAreaGroup
    isReadOnly?: boolean
    pendingTaskId: string | null
    showAssigneeColumn: boolean
    onTake: (taskId: string) => void | Promise<unknown>
    onReturn: (taskId: string) => void | Promise<unknown>
    onComplete: (taskId: string) => void | Promise<unknown>
    onOpen: (taskId: string) => void | Promise<unknown>
    onVerify?: (taskId: string) => void | Promise<unknown>
}

function renderAssignee(task: ResidentDutyTask) {
    if (!task.assignee_name) {
        return (
            <Tooltip label="Исполнитель">
                <Text size="sm" c="dimmed">
                    Не назначена
                </Text>
            </Tooltip>
        )
    }

    if (task.is_mine) {
        return (
            <Tooltip label="Исполнитель">
                <Text size="sm" fw={600} c="blue">
                    Вы
                </Text>
            </Tooltip>
        )
    }

    return (
        <Tooltip label="Исполнитель">
            <Text size="sm" fw={500}>
                {task.assignee_name}
            </Text>
        </Tooltip>
    )
}

export function TaskGroupSection({
    actionMode = 'default',
    group,
    isReadOnly = false,
    pendingTaskId,
    showAssigneeColumn,
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onVerify,
}: Props) {
    return (
        <Paper withBorder radius="lg" bg="white">
            <Text fw={700} size="lg" px="lg" py="md">
                {group.label}
            </Text>

            <Divider />

            <ScrollArea>
                <Table horizontalSpacing="lg" verticalSpacing="md" highlightOnHover miw={780}>
                    <Table.Tbody>
                        {group.tasks.map((task) => (
                            <Table.Tr key={task.id}>
                                <Table.Td w={isReadOnly ? '52%' : showAssigneeColumn ? '34%' : '52%'}>
                                    <Tooltip label="Задача">
                                        <Text fw={600}>{task.title}</Text>
                                    </Tooltip>
                                </Table.Td>
                                {!isReadOnly && (
                                    <Table.Td w="12%">
                                        <Tooltip label="Стоимость">
                                            <Text fw={600}>{task.cost}</Text>
                                        </Tooltip>
                                    </Table.Td>
                                )}
                                {showAssigneeColumn && (
                                    <Table.Td w={isReadOnly ? '24%' : '18%'}>{renderAssignee(task)}</Table.Td>
                                )}
                                <Table.Td w={isReadOnly ? '24%' : showAssigneeColumn ? '20%' : '36%'}>
                                    {isReadOnly ? (
                                        <TaskStatusBadge status={task.status} />
                                    ) : (
                                        <TaskRowActions
                                            mode={actionMode}
                                            pending={pendingTaskId === task.id}
                                            task={task}
                                            onTake={onTake}
                                            onReturn={onReturn}
                                            onComplete={onComplete}
                                            onOpen={onOpen}
                                            onVerify={onVerify}
                                        />
                                    )}
                                </Table.Td>
                            </Table.Tr>
                        ))}
                    </Table.Tbody>
                </Table>
            </ScrollArea>
        </Paper>
    )
}
