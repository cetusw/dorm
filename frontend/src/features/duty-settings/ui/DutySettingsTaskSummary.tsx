import type { ReactNode } from 'react'

import { CheckCircleIcon, CoinsIcon, UserIcon } from '@phosphor-icons/react'
import { Group, Text } from '@mantine/core'

import type { DutySettingsTaskSummary as DutySettingsTaskSummaryModel } from '../model/types'

type Props = {
    summary: DutySettingsTaskSummaryModel
}

type SummaryItemProps = {
    icon: ReactNode
    label: string
}

function SummaryItem({ icon, label }: SummaryItemProps) {
    return (
        <Group gap="xs" wrap="nowrap">
            {icon}
            <Text fw={500}>{label}</Text>
        </Group>
    )
}

export function DutySettingsTaskSummary({ summary }: Props) {
    return (
        <Group gap="lg" align="center" wrap="wrap" justify="flex-start">
            <SummaryItem icon={<CheckCircleIcon size={20} />} label={`${summary.task_count} задач`} />
            <SummaryItem icon={<CoinsIcon size={20} />} label={`${summary.total_cost} баллов`} />
            <SummaryItem icon={<UserIcon size={20} />} label={`${summary.cost_per_member} баллов на участника команды`} />
        </Group>
    )
}
