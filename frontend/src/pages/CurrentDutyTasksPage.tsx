import { Alert, Box, Center, Loader, Stack } from '@mantine/core'

import { DutySummary } from '../features/duty-tasks/DutySummary'
import { TaskGroups } from '../features/duty-tasks/TaskGroups'
import { useCurrentDuty } from '../features/duty-tasks/useCurrentDuty'
import { formatDutyPeriod } from '../features/duty-tasks/utils'

export function CurrentDutyTasksPage() {
    const {
        duty,
        error,
        loading,
        pendingTaskId,
        handleTake,
        handleReturn,
        handleComplete,
        handleOpen,
    } = useCurrentDuty()

    if (loading) {
        return (
            <Center py="xl">
                <Loader />
            </Center>
        )
    }

    if (error && !duty) {
        return (
            <Box px={{ base: 'md', md: 'xl' }} py="xl">
                <Alert color="red" title="Ошибка">
                    {error}
                </Alert>
            </Box>
        )
    }

    if (!duty) {
        return (
            <Box px={{ base: 'md', md: 'xl' }} py="xl">
                <Alert color="gray">Текущее дежурство не найдено.</Alert>
            </Box>
        )
    }

    return (
        <Box px={{ base: 'md', md: 'xl' }} py="xl">
            <Stack gap="lg" maw={1240} mx="auto">
                {error && (
                    <Alert color="red" title="Ошибка">
                        {error}
                    </Alert>
                )}

                <DutySummary
                    title={`Текущее дежурство · ${formatDutyPeriod(duty.start_date, duty.end_date)}`}
                    costPerResidentGoal={duty.cost_per_resident_goal}
                    myTakenCostSum={duty.my_taken_cost_sum}
                />

                <TaskGroups
                    pendingTaskId={pendingTaskId}
                    tasks={duty.tasks}
                    onTake={handleTake}
                    onReturn={handleReturn}
                    onComplete={handleComplete}
                    onOpen={handleOpen}
                />
            </Stack>
        </Box>
    )
}
