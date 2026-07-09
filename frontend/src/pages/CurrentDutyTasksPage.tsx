import { useEffect, useState } from 'react'

import { Alert, Box, Center, Loader, SimpleGrid, Stack, Title } from '@mantine/core'

import { DutyAnalyticsCard } from '../features/duty-tasks/DutyAnalyticsCard'
import { DutyTaskTabs, type DutyTaskTab } from '../features/duty-tasks/DutyTaskTabs'
import { TaskGroups } from '../features/duty-tasks/TaskGroups'
import { useCurrentDuty } from '../features/duty-tasks/useCurrentDuty'
import {
    countTasksCompletedByMe,
    countTasksTakenByMe,
    formatDutyPeriod,
    sumCostTakenByMe,
} from '../features/duty-tasks/utils'

const ACTIVE_TAB_STORAGE_KEY = 'current-duty-active-tab'

function readStoredTab(): DutyTaskTab {
    if (typeof window === 'undefined') {
        return 'mine'
    }

    const value = window.localStorage.getItem(ACTIVE_TAB_STORAGE_KEY)
    if (value === 'mine' || value === 'free' || value === 'all' || value === 'review') {
        return value
    }

    return 'mine'
}

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
        handleVerify,
        initialMineTaskIds,
        initialFreeTaskIds,
    } = useCurrentDuty()
    const [activeTab, setActiveTab] = useState<DutyTaskTab>(readStoredTab)
    const [reviewVisibleTaskIds, setReviewVisibleTaskIds] = useState<string[]>([])

    useEffect(() => {
        window.localStorage.setItem(ACTIVE_TAB_STORAGE_KEY, activeTab)
    }, [activeTab])

    useEffect(() => {
        if (!duty || activeTab !== 'review') {
            return
        }

        setReviewVisibleTaskIds(
            duty.tasks
                .filter((task) => task.status === 'completed')
                .map((task) => task.id),
        )
    }, [activeTab])

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
                <Alert color="gray">Ваша команда не дежурит на этой неделе</Alert>
            </Box>
        )
    }

    if (duty.tasks.length === 0) {
        return (
            <Box px={{ base: 'md', md: 'xl' }} py="xl">
                <Stack gap="lg" maw={1240} mx="auto">
                    <Title order={1}>
                        Текущее дежурство · {formatDutyPeriod(duty.start_date, duty.end_date)}
                    </Title>
                    <Alert color="gray">
                        На текущее дежурство не заведены задачи, сообщите об этом главе вашей
                        команды
                    </Alert>
                </Stack>
            </Box>
        )
    }

    const myTasks = duty.tasks.filter((task) => initialMineTaskIds.includes(task.id))
    const freeTasks = duty.tasks.filter((task) => initialFreeTaskIds.includes(task.id))
    const tasksByTab = {
        mine: myTasks,
        free: freeTasks,
        all: duty.tasks,
        review: duty.tasks.filter((task) => reviewVisibleTaskIds.includes(task.id)),
    } satisfies Record<DutyTaskTab, typeof duty.tasks>

    const takenCostSum = sumCostTakenByMe(duty.tasks)
    const takenTasksCount = countTasksTakenByMe(duty.tasks)
    const completedTasksCount = countTasksCompletedByMe(duty.tasks)

    async function handleReviewVerify(taskId: string) {
        const success = await handleVerify(taskId)
        if (!success) {
            return
        }

        setReviewVisibleTaskIds((current) => current.filter((currentTaskId) => currentTaskId !== taskId))
    }

    return (
        <Box px={{ base: 'md', md: 'xl' }} py="xl">
            <Stack gap="lg" maw={1240} mx="auto">
                {error && (
                    <Alert color="red" title="Ошибка">
                        {error}
                    </Alert>
                )}

                <Title order={1}>
                    Текущее дежурство · {formatDutyPeriod(duty.start_date, duty.end_date)}
                </Title>

                <SimpleGrid cols={{ base: 1, md: 2 }} spacing="lg">
                    <DutyAnalyticsCard
                        label="Взято задач"
                        currentValue={takenCostSum}
                        targetValue={duty.cost_per_resident_goal}
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

                <DutyTaskTabs activeTab={activeTab} onChange={setActiveTab} />

                <TaskGroups
                    actionMode={activeTab === 'review' ? 'review' : 'default'}
                    pendingTaskId={pendingTaskId}
                    tasks={tasksByTab[activeTab]}
                    emptyMessage={activeTab === 'review' ? 'Нет задач на проверке' : 'В этом разделе нет задач.'}
                    showAssigneeColumn={activeTab === 'all' || activeTab === 'review'}
                    onTake={handleTake}
                    onReturn={handleReturn}
                    onComplete={handleComplete}
                    onOpen={handleOpen}
                    onVerify={handleReviewVerify}
                />
            </Stack>
        </Box>
    )
}
