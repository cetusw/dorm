import { Button, Group } from '@mantine/core'

import type { ResidentDutyTask } from './types'

type Props = {
    pending: boolean
    task: ResidentDutyTask
    onTake: (taskId: string) => void
    onReturn: (taskId: string) => void
    onComplete: (taskId: string) => void
    onOpen: (taskId: string) => void
}

export function TaskRowActions({
    pending,
    task,
    onTake,
    onReturn,
    onComplete,
    onOpen,
}: Props) {
    return (
        <Group gap="xs" wrap="wrap" justify="flex-end">
            {task.can_take && (
                <Button
                    size="xs"
                    radius="md"
                    loading={pending}
                    onClick={() => onTake(task.id)}
                >
                    Взять
                </Button>
            )}

            {task.can_return && (
                <Button
                    size="xs"
                    radius="md"
                    variant="default"
                    loading={pending}
                    onClick={() => onReturn(task.id)}
                >
                    Вернуть
                </Button>
            )}

            {task.can_complete && (
                <Button
                    size="xs"
                    radius="md"
                    color="green"
                    loading={pending}
                    onClick={() => onComplete(task.id)}
                >
                    Выполнено
                </Button>
            )}

            {task.can_open && (
                <Button
                    size="xs"
                    radius="md"
                    variant="light"
                    color="yellow"
                    loading={pending}
                    onClick={() => onOpen(task.id)}
                >
                    Отменить выполнение
                </Button>
            )}
        </Group>
    )
}
