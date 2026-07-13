import { Group, Paper, Progress, Stack, Text } from '@mantine/core'

type Props = {
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

export function DutyAnalyticsCard({
    label,
    currentValue,
    targetValue,
    unitLabel,
    progressColor,
}: Props) {
    const progressValue = buildProgressValue(currentValue, targetValue)

    return (
        <Paper withBorder radius="lg" p="lg" bg="white">
            <Stack gap="md">
                <Text size="sm" c="dimmed" fw={600}>
                    {label}
                </Text>
                <Group justify="space-between" align="flex-end">
                    <Text size="xl" fw={700}>
                        {currentValue}
                    </Text>
                    <Text size="sm" c="dimmed">
                        из {targetValue} {unitLabel}
                    </Text>
                </Group>
                <Progress value={progressValue} color={progressColor} radius="xl" size="lg" />
            </Stack>
        </Paper>
    )
}
