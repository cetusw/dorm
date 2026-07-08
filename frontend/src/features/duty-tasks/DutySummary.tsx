import { Group, Paper, RingProgress, Stack, Text, Title } from '@mantine/core'

type Props = {
    title: string
    costPerResidentGoal: number
    myTakenCostSum: number
}

function buildProgressValue(currentValue: number, targetValue: number): number {
    if (targetValue <= 0) {
        return 0
    }

    return Math.min(100, Math.round((currentValue / targetValue) * 100))
}

export function DutySummary({ title, costPerResidentGoal, myTakenCostSum }: Props) {
    const progressValue = buildProgressValue(myTakenCostSum, costPerResidentGoal)

    return (
        <Paper withBorder radius="md" p="lg">
            <Title order={1}>{title}</Title>
            <Stack gap="lg">
                <Group align="center" wrap="wrap">
                    <RingProgress
                        size={112}
                        thickness={12}
                        roundCaps
                        sections={[{ value: progressValue, color: 'blue' }]}
                        label={
                            <Stack gap={0} align="center">
                                <Text fw={700} size="lg" lh={1}>
                                    {progressValue}%
                                </Text>
                            </Stack>
                        }
                    />
                    <Group gap="md" align="center">
                        <Text fw={600}>{costPerResidentGoal} / {myTakenCostSum}</Text>
                    </Group>
                </Group>
            </Stack>
        </Paper>
    )
}
