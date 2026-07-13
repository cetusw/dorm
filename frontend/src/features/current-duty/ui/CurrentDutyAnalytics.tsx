import { SimpleGrid } from '@mantine/core'

import type { DutyAnalytics } from '../model/selectors'
import { DutyAnalyticsCard } from './DutyAnalyticsCard'

type Props = {
    analytics: DutyAnalytics
    isReadOnly: boolean
    targetValue: number
}

export function CurrentDutyAnalytics({
    analytics,
    isReadOnly,
    targetValue,
}: Props) {
    if (isReadOnly) {
        return (
            <SimpleGrid cols={3} spacing="lg">
                <DutyAnalyticsCard
                    label="Всего взято"
                    currentValue={analytics.totalTakenTasksCount}
                    targetValue={analytics.totalTasksCount}
                    unitLabel="задач"
                    progressColor="blue"
                />
                <DutyAnalyticsCard
                    label="Всего выполнено"
                    currentValue={analytics.totalCompletedTasksCount}
                    targetValue={analytics.totalTasksCount}
                    unitLabel="задач"
                    progressColor="yellow"
                />
                <DutyAnalyticsCard
                    label="Всего проверено"
                    currentValue={analytics.totalVerifiedTasksCount}
                    targetValue={analytics.totalTasksCount}
                    unitLabel="задач"
                    progressColor="green"
                />
            </SimpleGrid>
        )
    }

    return (
        <SimpleGrid cols={{ base: 1, md: 3 }} spacing="lg">
            <DutyAnalyticsCard
                label="Взято"
                currentValue={analytics.takenCostSum}
                targetValue={targetValue}
                unitLabel="баллов"
                progressColor="blue"
            />
            <DutyAnalyticsCard
                label="Выполнено"
                currentValue={analytics.completedTasksCount}
                targetValue={analytics.takenTasksCount}
                unitLabel="задач"
                progressColor="green"
            />
            <DutyAnalyticsCard
                label="Проверено"
                currentValue={analytics.myVerifiedTasksCount}
                targetValue={analytics.takenTasksCount}
                unitLabel="задач"
                progressColor="teal"
            />
        </SimpleGrid>
    )
}
