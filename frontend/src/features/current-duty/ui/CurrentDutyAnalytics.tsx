import { SimpleGrid } from '@mantine/core'

import { DutyAnalyticsCard } from './DutyAnalyticsCard'

type Props = {
    completedTasksCount: number
    takenCostSum: number
    takenTasksCount: number
    targetValue: number
}

export function CurrentDutyAnalytics({
    completedTasksCount,
    takenCostSum,
    takenTasksCount,
    targetValue,
}: Props) {
    return (
        <SimpleGrid cols={{ base: 1, md: 2 }} spacing="lg">
            <DutyAnalyticsCard
                label="Взято задач"
                currentValue={takenCostSum}
                targetValue={targetValue}
                unitLabel="баллов"
                progressColor="blue"
            />
            <DutyAnalyticsCard
                label="Выполнено"
                currentValue={completedTasksCount}
                targetValue={takenTasksCount}
                unitLabel="задач"
                progressColor="green"
            />
        </SimpleGrid>
    )
}
