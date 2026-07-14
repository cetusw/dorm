import { Group, Paper, Progress, SimpleGrid, Stack, Text } from '@mantine/core'

import type { DutyAnalytics } from '../model/selectors'
import { DutyAnalyticsCard } from './DutyAnalyticsCard'

type Props = {
    analytics: DutyAnalytics
    isReadOnly: boolean
    targetValue: number
}

type AnalyticsItem = {
    label: string
    currentValue: number
    targetValue: number
    unitLabel: string
    progressColor: string
}

function buildProgressValue(currentValue: number, targetValue: number): number {
    if (targetValue <= 0) {
        return 0
    }

    return Math.min(100, Math.round((currentValue / targetValue) * 100))
}

function MobileAnalytics({ items }: { items: AnalyticsItem[] }) {
    return (
        <Paper hiddenFrom="md" withBorder radius="lg" p="sm" bg="white">
            <Stack gap="xs">
                {items.map((item) => (
                    <Group key={item.label} gap="sm" wrap="nowrap" align="center">
                        <Progress
                            value={buildProgressValue(item.currentValue, item.targetValue)}
                            color={item.progressColor}
                            radius="xl"
                            size="sm"
                            style={{ flex: 1, minWidth: 0 }}
                        />
                        <Text size="xs" fw={600} ta="right" miw={88}>
                            {item.label} {item.currentValue}/{item.targetValue}
                        </Text>
                    </Group>
                ))}
            </Stack>
        </Paper>
    )
}

export function CurrentDutyAnalytics({
    analytics,
    isReadOnly,
    targetValue,
}: Props) {
    const items: AnalyticsItem[] = isReadOnly
        ? [
            {
                label: 'Всего взято',
                currentValue: analytics.totalTakenTasksCount,
                targetValue: analytics.totalTasksCount,
                unitLabel: 'задач',
                progressColor: 'blue',
            },
            {
                label: 'Всего выполнено',
                currentValue: analytics.totalCompletedTasksCount,
                targetValue: analytics.totalTasksCount,
                unitLabel: 'задач',
                progressColor: 'yellow',
            },
            {
                label: 'Всего проверено',
                currentValue: analytics.totalVerifiedTasksCount,
                targetValue: analytics.totalTasksCount,
                unitLabel: 'задач',
                progressColor: 'green',
            },
        ]
        : [
            {
                label: 'Взято',
                currentValue: analytics.takenCostSum,
                targetValue,
                unitLabel: 'баллов',
                progressColor: 'blue',
            },
            {
                label: 'Выполнено',
                currentValue: analytics.completedTasksCount,
                targetValue: analytics.takenTasksCount,
                unitLabel: 'задач',
                progressColor: 'green',
            },
            {
                label: 'Проверено',
                currentValue: analytics.myVerifiedTasksCount,
                targetValue: analytics.takenTasksCount,
                unitLabel: 'задач',
                progressColor: 'teal',
            },
        ]

    if (isReadOnly) {
        return (
            <>
                <MobileAnalytics items={items} />
                <SimpleGrid visibleFrom="md" cols={3} spacing="lg">
                    {items.map((item) => (
                        <DutyAnalyticsCard
                            key={item.label}
                            label={item.label}
                            currentValue={item.currentValue}
                            targetValue={item.targetValue}
                            unitLabel={item.unitLabel}
                            progressColor={item.progressColor}
                        />
                    ))}
                </SimpleGrid>
            </>
        )
    }

    return (
        <>
            <MobileAnalytics items={items} />
            <SimpleGrid visibleFrom="md" cols={3} spacing="lg">
                {items.map((item) => (
                    <DutyAnalyticsCard
                        key={item.label}
                        label={item.label}
                        currentValue={item.currentValue}
                        targetValue={item.targetValue}
                        unitLabel={item.unitLabel}
                        progressColor={item.progressColor}
                    />
                ))}
            </SimpleGrid>
        </>
    )
}
