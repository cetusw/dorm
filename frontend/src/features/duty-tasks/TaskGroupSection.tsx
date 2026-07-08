import { Divider, Paper, ScrollArea, Table, Text, Tooltip } from '@mantine/core'

import { TaskRowActions } from './TaskRowActions'
import type { ResidentDutyTask } from './types'
import type { TaskAreaGroup } from './utils'

type Props = {
    group: TaskAreaGroup
    pendingTaskId: string | null
    onTake: (taskId: string) => void
    onReturn: (taskId: string) => void
    onComplete: (taskId: string) => void
    onOpen: (taskId: string) => void
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
    group,
    pendingTaskId,
    onTake,
    onReturn,
    onComplete,
    onOpen,
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
                                <Table.Td w="34%">
                                    <Tooltip label="Задача">
                                        <Text fw={600}>{task.title}</Text>
                                    </Tooltip>
                                </Table.Td>
                                <Table.Td w="12%">
                                    <Tooltip label="Стоимость">
                                        <Text fw={600}>{task.cost}</Text>
                                    </Tooltip>
                                </Table.Td>
                                <Table.Td w="18%">{renderAssignee(task)}</Table.Td>
                                <Table.Td w="20%">
                                    <TaskRowActions
                                        pending={pendingTaskId === task.id}
                                        task={task}
                                        onTake={onTake}
                                        onReturn={onReturn}
                                        onComplete={onComplete}
                                        onOpen={onOpen}
                                    />
                                </Table.Td>
                            </Table.Tr>
                        ))}
                    </Table.Tbody>
                </Table>
            </ScrollArea>
        </Paper>
    )
}
