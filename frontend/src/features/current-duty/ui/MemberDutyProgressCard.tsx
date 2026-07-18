import { Box, Paper, Stack, Text } from '@mantine/core'

import type { MemberDutyProgressModel } from '../model/selectors'

type Props = {
    progress: MemberDutyProgressModel
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

export function MemberDutyProgressCard({ progress }: Props) {
    const sections = buildProgressSections(progress)

    return (
        <Paper withBorder radius="lg" p="lg" bg="var(--app-color-surface)">
            <Stack gap="sm">
                <Text size="lg" fw={700} c="var(--app-color-text)">
                    {progress.title}
                </Text>
                <Box
                    style={{
                        display: 'flex',
                        height: 18,
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
            </Stack>
        </Paper>
    )
}
