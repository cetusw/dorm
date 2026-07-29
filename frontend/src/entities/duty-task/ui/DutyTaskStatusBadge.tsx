import { Badge, Group } from '@mantine/core'

import type { DutyTaskStatus } from '../model/status'
import { getDutyTaskStatusConfig } from '../model/status'

type Props = {
    status: DutyTaskStatus
    justify?: 'flex-start' | 'flex-end'
}

export function DutyTaskStatusBadge({ status, justify = 'flex-end' }: Props) {
    const config = getDutyTaskStatusConfig(status)

    return (
        <Group justify={justify}>
            <Badge
                variant="filled"
                radius="sm"
                styles={{
                    root: {
                        backgroundColor: config.backgroundColor,
                        color: config.textColor,
                        fontWeight: 500,
                        border: 'none',
                        textTransform: 'none',
                    },
                    label: {
                        textTransform: 'none',
                    },
                }}
            >
                {config.label}
            </Badge>
        </Group>
    )
}
