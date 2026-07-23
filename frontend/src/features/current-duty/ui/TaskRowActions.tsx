import { Button, Group } from '@mantine/core'

import type { ResidentDutyTask } from '../model/types'

export type TaskRowActionMode = 'default' | 'review'

type Props = {
    mode?: TaskRowActionMode
    pending: boolean
    task: ResidentDutyTask
    onTake: (taskId: string) => void | Promise<unknown>
    onReturn: (taskId: string) => void | Promise<unknown>
    onComplete: (taskId: string) => void | Promise<unknown>
    onOpen: (taskId: string) => void | Promise<unknown>
    onReopen?: (taskId: string) => void | Promise<unknown>
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
    onReopen,
    onVerify,
}: Props) {
    if (mode === 'review' && task.status === 'completed' && task.can_verify && onVerify) {
        if (task.can_review_open && onReopen) {
            return (
                <Group grow>
                    <Button
                        radius="md"
                        variant="default"
                        loading={pending}
                        styles={{
                            root: {
                                backgroundColor: '#FEE2E2',
                                borderColor: '#991B1B',
                                color: 'var(--app-color-text)',
                                fontWeight: 500,
                            },
                        }}
                        onClick={() => onReopen(task.id)}
                    >
                        Переоткрыть
                    </Button>
                    <Button fullWidth radius="md" variant="default" loading={pending} onClick={() => onVerify(task.id)}>
                        Подтвердить
                    </Button>
                </Group>
            )
        }

        return (
            <Button fullWidth radius="md" variant="default" loading={pending} onClick={() => onVerify(task.id)}>
                Подтвердить
            </Button>
        )
    }

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
                    <Button radius="md" variant="default" loading={pending} onClick={() => onReturn(task.id)}>
                        Вернуть
                    </Button>
                )}
                {task.can_complete && (
                    <Button radius="md" color="brand" loading={pending} onClick={() => onComplete(task.id)}>
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
                color="brand"
                loading={pending}
                onClick={() => onOpen(task.id)}
            >
                Отменить выполнение
            </Button>
        )
    }

    return null
}
