import { Alert, Drawer, ScrollArea, Table, Text, Tooltip } from '@mantine/core'

import type { ResidentDutyTask } from '../model/types'
import { TaskRowActions, type TaskRowActionMode } from '../ui/TaskRowActions'
import { TaskStatusBadge } from '../ui/TaskStatusBadge'
import type { FloorPlan } from './types'
import { getFloorLabel } from './utils'

type Props = {
    actionMode: TaskRowActionMode
    floorPlan: FloorPlan
    isReadOnly?: boolean
    opened: boolean
    pendingTaskId: string | null
    tasks: ResidentDutyTask[]
    onClose: () => void
    onTake: (taskId: string) => void | Promise<unknown>
    onReturn: (taskId: string) => void | Promise<unknown>
    onComplete: (taskId: string) => void | Promise<unknown>
    onOpen: (taskId: string) => void | Promise<unknown>
    onReopen?: (taskId: string) => void | Promise<unknown>
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

export function BuildingPlanDrawer({
    actionMode,
    floorPlan,
    isReadOnly = false,
    opened,
    pendingTaskId,
    tasks,
    onClose,
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onReopen,
    onVerify,
}: Props) {
    const title = `${getFloorLabel(floorPlan.floor)} · ${tasks[0]?.area_name ?? 'Территория'}`

    return (
        <Drawer opened={opened} onClose={onClose} position="right" size={860} title={opened ? title : 'Территория'}>
            {tasks.length === 0 ? (
                <Alert color="gray">Для этой территории нет задач.</Alert>
            ) : (
                <ScrollArea>
                    <Table horizontalSpacing="lg" verticalSpacing="md" highlightOnHover miw={780}>
                        <Table.Tbody>
                            {tasks.map((task) => (
                                <Table.Tr key={task.id}>
                                    <Table.Td w={isReadOnly ? '52%' : '34%'}>
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
                                    <Table.Td w={isReadOnly ? '24%' : '18%'}>
                                        {renderAssignee(task)}
                                    </Table.Td>
                                    <Table.Td w={isReadOnly ? '24%' : '36%'}>
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
                                                onReopen={onReopen}
                                                onVerify={onVerify}
                                            />
                                        )}
                                    </Table.Td>
                                </Table.Tr>
                            ))}
                        </Table.Tbody>
                    </Table>
                </ScrollArea>
            )}
        </Drawer>
    )
}
