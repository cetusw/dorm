import { Badge, ScrollArea, Stack, Table, Text, Tooltip } from '@mantine/core'

import type { ResidentDutyTask } from '../model/types'
import { TaskStatusBadge } from './TaskStatusBadge'

type Props = {
    tasks: ResidentDutyTask[]
}

export function TeamMemberTaskList({ tasks }: Props) {
    return (
        <>
            <Stack hiddenFrom="md" gap="sm">
                {tasks.map((task) => (
                    <Stack
                        key={task.id}
                        gap="xs"
                        py="sm"
                        style={{ borderTop: '1px solid var(--app-color-border)' }}
                    >
                        <Tooltip label="Задача">
                            <Text fw={600}>{task.title}</Text>
                        </Tooltip>
                        <Tooltip label="Территория">
                            <Text size="sm" c="dimmed">
                                {task.area_floor} этаж · {task.area_name}
                            </Text>
                        </Tooltip>
                        <Tooltip label="Балл">
                            <Badge variant="light" radius="sm" color="gray" w="fit-content">
                                {task.cost} баллов
                            </Badge>
                        </Tooltip>
                        <Tooltip label="Статус">
                            <div>
                                <TaskStatusBadge task={task} justify="flex-start" />
                            </div>
                        </Tooltip>
                    </Stack>
                ))}
            </Stack>

            <ScrollArea visibleFrom="md">
                <Table horizontalSpacing="lg" verticalSpacing="md" highlightOnHover miw={760}>
                    <Table.Tbody>
                        {tasks.map((task) => (
                            <Table.Tr key={task.id}>
                                <Table.Td w="24%">
                                    <Tooltip label="Территория">
                                        <Text>{task.area_floor} этаж · {task.area_name}</Text>
                                    </Tooltip>
                                </Table.Td>
                                <Table.Td w="42%">
                                    <Tooltip label="Задача">
                                        <Text fw={600}>{task.title}</Text>
                                    </Tooltip>
                                </Table.Td>
                                <Table.Td w="12%">
                                    <Tooltip label="Балл">
                                        <Text fw={600}>{task.cost}</Text>
                                    </Tooltip>
                                </Table.Td>
                                <Table.Td w="22%">
                                    <Tooltip label="Статус">
                                        <div>
                                            <TaskStatusBadge task={task} justify="flex-start" />
                                        </div>
                                    </Tooltip>
                                </Table.Td>
                            </Table.Tr>
                        ))}
                    </Table.Tbody>
                </Table>
            </ScrollArea>
        </>
    )
}
