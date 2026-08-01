import type { ReactNode } from 'react'

import { Box, Paper, Stack, Text, Tooltip } from '@mantine/core'
import { useMediaQuery } from '@mantine/hooks'

import type { MemberDutyProgressModel } from '../model/selectors'

type Props = {
    progress: MemberDutyProgressModel
    showTitle?: boolean
    withContainer?: boolean
    tooltip?: ReactNode
    barWidth?: number | string
}

function buildProgressSections(progress: MemberDutyProgressModel) {
    const total = progress.sections.reduce((sum, section) => sum + section.value, 0)

    if (total <= 0) {
        return []
    }

    return progress.sections.map((section) => ({
        ...section,
        width: (section.value / total) * 100,
    }))
}

function ProgressBar({
    sections,
    tooltip,
    width,
}: {
    sections: ReturnType<typeof buildProgressSections>
    tooltip?: ReactNode
    width?: number | string
}) {
    const bar = (
        <Box
            style={{
                display: 'flex',
                height: 18,
                width,
                maxWidth: '100%',
                backgroundColor: 'var(--app-color-analytics-empty)',
                border: '1px solid var(--app-color-border)',
                borderRadius: 999,
                overflow: 'hidden',
            }}
        >
            {sections.map((section, index) => (
                <Box
                    key={`${section.color}-${index}`}
                    style={{
                        width: `${section.width}%`,
                        backgroundColor: section.color,
                        flexShrink: 0,
                    }}
                />
            ))}
        </Box>
    )

    if (!tooltip) {
        return bar
    }

    return (
        <Tooltip label={tooltip} multiline withArrow>
            <Box>{bar}</Box>
        </Tooltip>
    )
}

export function MemberDutyProgressCard({
    progress,
    showTitle = true,
    withContainer = true,
    tooltip,
    barWidth,
}: Props) {
    const allowTooltip = useMediaQuery('(hover: hover) and (pointer: fine)')
    const sections = buildProgressSections(progress)
    const content = (
        <Stack gap="sm">
            {showTitle ? (
                <Text size="lg" fw={700} c="var(--app-color-text)">
                    {progress.title}
                </Text>
            ) : null}
            <ProgressBar
                sections={sections}
                tooltip={allowTooltip ? tooltip : undefined}
                width={barWidth}
            />
        </Stack>
    )

    if (!withContainer) {
        return content
    }

    return <Paper withBorder radius="lg" p="lg" bg="var(--app-color-surface)">{content}</Paper>
}
