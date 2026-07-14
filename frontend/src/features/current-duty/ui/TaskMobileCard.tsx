import { Badge, Button, Group, Paper, Stack, Text } from '@mantine/core'

import type { ResidentDutyTask } from '../model/types'
import { TaskRowActions, type TaskRowActionMode } from './TaskRowActions'

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

type StatusConfig = {
    color: string
    label: string
    note?: string
}

function MobileTaskStatus({ task }: { task: ResidentDutyTask }) {
    const statusConfig: StatusConfig | null = (() => {
        if (task.status === 'assigned' && !task.is_mine) {
            return {
                color: 'blue',
                label: task.assignee_name || 'Задача в работе',
                note: 'В работе',
            }
        }

        if (task.status === 'completed') {
            return {
                color: 'yellow',
                label: 'На проверке',
            }
        }

        if (task.status === 'verified') {
            return {
                color: 'green',
                label: 'Подтверждена',
            }
        }

        if (task.status === 'free') {
            return {
                color: 'gray',
                label: 'Свободна',
            }
        }

        return null
    })()

    if (!statusConfig) {
        return null
    }

    return (
        <Paper
            withBorder
            radius="md"
            px="sm"
            py={task.status === 'assigned' && !task.is_mine ? 'xs' : 10}
            bg={`var(--mantine-color-${statusConfig.color}-0)`}
            style={{
                borderColor: `var(--mantine-color-${statusConfig.color}-2)`,
            }}
        >
            <Stack gap={2} align="center">
                <Text size="sm" fw={700} ta="center">
                    {statusConfig.label}
                </Text>
                {statusConfig.note && (
                    <Text size="xs" c="dimmed" ta="center">
                        {statusConfig.note}
                    </Text>
                )}
            </Stack>
        </Paper>
    )
}

function MobileTaskActions({
    pending,
    task,
    onTake,
    onReturn,
    onComplete,
    onOpen,
}: Omit<Props, 'mode' | 'onVerify'>) {
    if (task.can_take) {
        return (
            <Button fullWidth radius="md" loading={pending} onClick={() => onTake(task.id)}>
                Взять
            </Button>
        )
    }

    if (task.can_return || task.can_complete) {
        return (
            <Group grow>
                {task.can_return && (
                    <Button
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
                        radius="md"
                        color="green"
                        loading={pending}
                        onClick={() => onComplete(task.id)}
                    >
                        Выполнить
                    </Button>
                )}
            </Group>
        )
    }

    if (task.can_open) {
        return (
            <Button
                fullWidth
                radius="md"
                variant="light"
                color="yellow"
                loading={pending}
                onClick={() => onOpen(task.id)}
            >
                Отменить выполнение
            </Button>
        )
    }

    return <MobileTaskStatus task={task} />
}

export function TaskMobileCard({
    mode = 'default',
    pending,
    task,
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onVerify,
}: Props) {
    return (
        <Paper withBorder radius="lg" p="md" bg="white">
            <Stack gap="sm">
                <Stack gap={4}>
                    <Text fw={600} size="sm">
                        {task.title}
                    </Text>
                    <Group gap="xs">
                        <Badge variant="light" radius="sm" color="gray">
                            {task.cost} баллов
                        </Badge>
                    </Group>
                </Stack>

                {mode === 'default' ? (
                    <MobileTaskActions
                        pending={pending}
                        task={task}
                        onTake={onTake}
                        onReturn={onReturn}
                        onComplete={onComplete}
                        onOpen={onOpen}
                    />
                ) : (
                    <TaskRowActions
                        mode={mode}
                        pending={pending}
                        task={task}
                        onTake={onTake}
                        onReturn={onReturn}
                        onComplete={onComplete}
                        onOpen={onOpen}
                        onVerify={onVerify}
                    />
                )}
            </Stack>
        </Paper>
    )
}
