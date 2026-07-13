import {Badge, Group} from '@mantine/core'

import type {DutyTaskStatus} from '../model/types'

type Props = {
    status: DutyTaskStatus
}

const statusMap: Record<DutyTaskStatus, { color: string; label: string }> = {
    free: {color: 'gray', label: 'Свободна'},
    assigned: {color: 'blue', label: 'Взята'},
    completed: {color: 'yellow', label: 'На проверке'},
    verified: {color: 'green', label: 'Проверена'},
}

export function TaskStatusBadge({status}: Props) {
    const config = statusMap[status]

    return (
        <Group justify="flex-end">
            <Badge color={config.color} variant="light" radius="sm">
                {config.label}
            </Badge>
        </Group>
    )
}
