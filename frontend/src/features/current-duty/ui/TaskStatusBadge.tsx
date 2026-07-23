import { Badge, Group } from '@mantine/core'

import type { ResidentDutyTask } from '../model/types'

type Props = {
    task: ResidentDutyTask
    justify?: 'flex-start' | 'flex-end'
}

type BadgeConfig = {
    backgroundColor: string
    label: string
}

export function getStatusBadgeConfig(task: ResidentDutyTask): BadgeConfig | null {
    if (task.status === 'assigned') {
        return {
            backgroundColor: '#E0F2FE',
            label: 'ВЗЯТО',
        }
    }

    if (task.status === 'completed') {
        return {
            backgroundColor: '#FEF3C7',
            label: 'На проверке',
        }
    }

    if (task.status === 'verified') {
        return {
            backgroundColor: '#DCFCE7',
            label: 'Проверено',
        }
    }

    return null
}

export function TaskStatusBadge({ task, justify = 'flex-end' }: Props) {
    const config = getStatusBadgeConfig(task)

    if (!config) {
        return null
    }

    return (
        <Group justify={justify}>
            <Badge
                variant="filled"
                radius="sm"
                styles={{
                    root: {
                        backgroundColor: config.backgroundColor,
                        color: 'var(--app-color-text)',
                        fontWeight: 500,
                    },
                }}
            >
                {config.label}
            </Badge>
        </Group>
    )
}
