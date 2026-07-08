import { useEffect, useState } from 'react'
import { Alert, Container, Loader, Stack, Text, Title } from '@mantine/core'

import {
    completeTask,
    getCurrentDuty,
    returnTask,
    takeTask,
} from '../features/duty-tasks/api'
import { TaskCard } from '../features/duty-tasks/TaskCard'
import type { ResidentCurrentDuty } from '../features/duty-tasks/types'

export function CurrentDutyTasksPage() {
    const [duty, setDuty] = useState<ResidentCurrentDuty | null>(null)
    const [error, setError] = useState<string | null>(null)
    const [loading, setLoading] = useState(true)

    async function loadDuty() {
        setLoading(true)
        setError(null)

        try {
            const data = await getCurrentDuty()
            setDuty(data)
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Неизвестная ошибка')
        } finally {
            setLoading(false)
        }
    }

    async function runAction(action: () => Promise<void>) {
        try {
            await action()
            await loadDuty()
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Не удалось выполнить действие')
        }
    }

    useEffect(() => {
        void loadDuty()
    }, [])

    if (loading) {
        return (
            <Container py="xl">
                <Loader />
            </Container>
        )
    }

    if (error) {
        return (
            <Container py="xl">
                <Alert color="red" title="Ошибка">
                    {error}
                </Alert>
            </Container>
        )
    }

    if (!duty) {
        return (
            <Container py="xl">
                <Text>Текущее дежурство не найдено.</Text>
            </Container>
        )
    }

    return (
        <Container size="sm" py="xl">
            <Stack gap="md">
                <div>
                    <Title order={1}>Задачи текущего дежурства</Title>
                    <Text c="dimmed">
                        {duty.group} · {duty.team}
                    </Text>
                    <Text size="sm" c="dimmed">
                        {duty.start_date} — {duty.end_date}
                    </Text>
                </div>

                {duty.tasks.length === 0 ? (
                    <Alert color="gray">На текущее дежурство нет задач.</Alert>
                ) : (
                    duty.tasks.map((task) => (
                        <TaskCard
                            key={task.id}
                            task={task}
                            onTake={(taskId) => runAction(() => takeTask(taskId))}
                            onReturn={(taskId) => runAction(() => returnTask(taskId))}
                            onComplete={(taskId) => runAction(() => completeTask(taskId))}
                        />
                    ))
                )}
            </Stack>
        </Container>
    )
}