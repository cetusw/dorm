import { Button, Card, Group, Stack, Text, Title } from '@mantine/core'
import type { ResidentDutyTask } from './types'
import { TaskStatusBadge } from './TaskStatusBadge'

type Props = {
    task: ResidentDutyTask
    onTake: (taskId: string) => void
    onReturn: (taskId: string) => void
    onComplete: (taskId: string) => void
}

export function TaskCard({ task, onTake, onReturn, onComplete }: Props) {
    return (
        <Card withBorder radius="md" padding="md">
            <Stack gap="xs">
                <Group justify="space-between" align="flex-start">
                    <div>
                        <Title order={3}>{task.title}</Title>
                        <Text size="sm" c="dimmed">
                            {task.area_floor} этаж · {task.area_name}
                        </Text>
                    </div>

                    <TaskStatusBadge status={task.status} />
                </Group>

                <Text size="sm">Стоимость: {task.cost} б.</Text>

                {task.assignee_name && (
                    <Text size="sm" c="dimmed">
                        Исполнитель: {task.assignee_name}
                    </Text>
                )}

                <Group mt="sm">
                    {task.can_take && (
                        <Button onClick={() => onTake(task.id)}>
                            Взять
                        </Button>
                    )}

                    {task.can_return && (
                        <Button variant="light" onClick={() => onReturn(task.id)}>
                            Вернуть
                        </Button>
                    )}

                    {task.can_complete && (
                        <Button onClick={() => onComplete(task.id)}>
                            Выполнено
                        </Button>
                    )}
                </Group>
            </Stack>
        </Card>
    )
}