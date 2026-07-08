import { Badge } from '@mantine/core'
import type { DutyTaskStatus } from './types'

type Props = {
    status: DutyTaskStatus
}

export function TaskStatusBadge({ status }: Props) {
    switch (status) {
        case 'free':
            return <Badge color="gray">Свободна</Badge>
        case 'assigned':
            return <Badge color="blue">Взята</Badge>
        case 'completed':
            return <Badge color="yellow">На проверке</Badge>
        case 'verified':
            return <Badge color="green">Проверена</Badge>
    }
}