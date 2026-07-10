import { Button, Group, Text } from '@mantine/core'

import type { ResidentDutyTask } from './types'
import { TaskStatusBadge } from './TaskStatusBadge'

export type TaskRowActionMode = 'default' | 'review'

type Props = {
    mode?: TaskRowActionMode
    pending: boolean
    task: ResidentDutyTask
    onTake: (taskId: string) => void | Promise<unknown>
    onReturn: (taskId: string) => void | Promise<unknown>
    onComplete: (taskId: string) => void | Promise<unknown>
    onOpen: (taskId: string) => void | Promise<unknown>
    onVerify?: (taskId: string) => void | Promise<unknown>
}

export function TaskRowActions({
    mode = 'default',
    pending,
    task,
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onVerify,
}: Props) {
    const hasDefaultActions =
        task.can_take || task.can_return || task.can_complete || task.can_open

    if (mode === 'review') {
        return (
            <Group gap="xs" wrap="nowrap" justify="flex-end">
                {task.status === 'completed' && task.can_review_open && (
                    <Button
                        size="xs"
                        radius="md"
                        variant="default"
                        loading={pending}
                        onClick={() => onOpen(task.id)}
                    >
                        Отменить
                    </Button>
                )}

                {task.status === 'completed' && task.can_verify && onVerify && (
                    <Button
                        size="xs"
                        radius="md"
                        color="green"
                        loading={pending}
                        onClick={() => onVerify(task.id)}
                    >
                        Подтвердить
                    </Button>
                )}

                {task.status === 'assigned' && (
                    <Button
                        size="xs"
                        radius="md"
                        variant="light"
                        color="blue"
                        loading={pending}
                        onClick={() => onComplete(task.id)}
                    >
                        Вернуть на проверку
                    </Button>
                )}
            </Group>
        )
    }

    if (!hasDefaultActions) {
        if (task.status === 'verified' && task.is_mine) {
            return (
                <Group gap="xs" wrap="nowrap" justify="flex-end">
                    <Text size="sm" fw={600} c="green">
                        Подтверждено
                    </Text>
                </Group>
            )
        }

        return <TaskStatusBadge status={task.status} />
    }

    if (task.status === 'verified' && task.is_mine) {
        return (
            <Group gap="xs" wrap="nowrap" justify="flex-end">
                <Text size="sm" fw={600} c="green">
                    Подтверждено
                </Text>
            </Group>
        )
    }

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
