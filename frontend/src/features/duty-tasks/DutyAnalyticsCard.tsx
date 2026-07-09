import { Group, Paper, RingProgress, Stack, Text } from '@mantine/core'

type Props = {
    currentValue: number
    label: string
    progressColor: string
    targetValue: number
    unitLabel: string
}

function buildProgressValue(currentValue: number, targetValue: number): number {
    if (targetValue <= 0) {
        return 0
    }

    return Math.min(100, Math.round((currentValue / targetValue) * 100))
}

export function DutyAnalyticsCard({
    currentValue,
    label,
    progressColor,
    targetValue,
    unitLabel,
}: Props) {
    const progressValue = buildProgressValue(currentValue, targetValue)

    return (
        <Paper withBorder radius="md" p="lg">
            <Group gap="md" wrap="nowrap" align="center">
                <RingProgress
                    size={92}
                    thickness={10}
                    roundCaps
                    sections={[{ value: progressValue, color: progressColor }]}
                />

                <Stack gap={4}>
                    <Text size="sm" c="dimmed" fw={600}>
                        {label}
                    </Text>
                    <Text fw={700} size="lg">
                        {currentValue} из {targetValue} {unitLabel}
                    </Text>
                </Stack>
            </Group>
        </Paper>
    )
}
