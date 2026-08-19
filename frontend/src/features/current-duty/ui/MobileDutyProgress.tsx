import { useMemo, useState } from 'react'

import { Box, Group, Popover, Stack, Text, UnstyledButton } from '@mantine/core'
import { useElementSize } from '@mantine/hooks'

import type { DutyAnalytics } from '../model/selectors'
import {
    buildMemberDutyProgressModel,
    buildReadonlyDutyProgressModel,
} from '../model/selectors'

type Props = {
    analytics: DutyAnalytics
    isReadOnly: boolean
    targetValue: number
}

type LegendItem = {
    color: string
    label: string
}

function buildLegendItems(
    analytics: DutyAnalytics,
    isReadOnly: boolean,
    targetValue: number,
): LegendItem[] {
    if (isReadOnly) {
        return [
            {
                color: 'var(--app-color-analytics-taken)',
                label: `Взято ${analytics.totalTakenTasksCount} из ${analytics.totalTasksCount} задач`,
            },
            {
                color: 'var(--app-color-analytics-completed)',
                label: `Выполнено ${analytics.totalCompletedTasksCount} из ${analytics.totalTasksCount} задач`,
            },
            {
                color: 'var(--app-color-analytics-verified)',
                label: `Проверено ${analytics.totalVerifiedTasksCount} из ${analytics.totalTasksCount} задач`,
            },
        ]
    }

    return [
        {
            color: 'var(--app-color-analytics-taken)',
            label: `Взято ${analytics.takenCostSum} из ${targetValue} баллов`,
        },
        {
            color: 'var(--app-color-analytics-completed)',
            label: `Выполнено ${analytics.completedTasksCount} из ${analytics.takenTasksCount} задач`,
        },
        {
            color: 'var(--app-color-analytics-verified)',
            label: `Проверено ${analytics.myVerifiedTasksCount} из ${analytics.takenTasksCount} задач`,
        },
    ]
}

export function MobileDutyProgress({
    analytics,
    isReadOnly,
    targetValue,
}: Props) {
    const [opened, setOpened] = useState(false)
    const { ref } = useElementSize<HTMLButtonElement>()
    const progress = isReadOnly
        ? buildReadonlyDutyProgressModel(analytics)
        : buildMemberDutyProgressModel(analytics, targetValue)
    const total = progress.sections.reduce((sum, section) => sum + section.value, 0)
    const legendItems = buildLegendItems(analytics, isReadOnly, targetValue)
    const dropdownWidth = useMemo(
        () => `${Math.max(...legendItems.map((item) => item.label.length), 0) + 6}ch`,
        [legendItems],
    )

    return (
        <Popover
            opened={opened}
            onChange={setOpened}
            position="bottom-end"
            offset={8}
            shadow="none"
            withArrow={false}
            withinPortal={false}
        >
            <Popover.Target>
                <UnstyledButton
                    ref={ref}
                    aria-label="Открыть детали прогресса"
                    onClick={() => setOpened((value) => !value)}
                    style={{
                        display: 'block',
                        width: '100%',
                        minWidth: 0,
                    }}
                >
                    <Box
                        style={{
                            display: 'flex',
                            height: 20,
                            width: '100%',
                            backgroundColor: 'var(--app-color-analytics-empty)',
                            border: '1px solid var(--app-color-border)',
                            borderRadius: 999,
                            overflow: 'hidden',
                        }}
                    >
                        {total > 0 ? progress.sections.map((section, index) => (
                            <Box
                                key={`${section.color}-${index}`}
                                style={{
                                    width: `${(section.value / total) * 100}%`,
                                    backgroundColor: section.color,
                                    flexShrink: 0,
                                }}
                            />
                        )) : null}
                    </Box>
                </UnstyledButton>
            </Popover.Target>

            <Popover.Dropdown
                p={10}
                style={{
                    width: dropdownWidth,
                    maxWidth: 'calc(100vw - 32px)',
                    border: '1px solid var(--app-color-border)',
                    borderRadius: 15,
                    backgroundColor: 'var(--app-color-surface)',
                }}
            >
                <Stack gap={10}>
                    <Text
                        size="14px"
                        c="var(--app-color-text-muted)"
                        fw={400}
                    >
                        Прогресс
                    </Text>

                    <Stack gap={8}>
                        {legendItems.map((item) => (
                            <Group key={item.label} gap={5} wrap="nowrap" align="center">
                                <Box
                                    style={{
                                        width: 12,
                                        minWidth: 12,
                                        height: 12,
                                        borderRadius: '50%',
                                        backgroundColor: item.color,
                                    }}
                                />
                                <Text
                                    size="14px"
                                    c="var(--app-color-text)"
                                    lh={1.3}
                                    style={{ whiteSpace: 'nowrap' }}
                                >
                                    {item.label}
                                </Text>
                            </Group>
                        ))}
                    </Stack>
                </Stack>
            </Popover.Dropdown>
        </Popover>
    )
}
