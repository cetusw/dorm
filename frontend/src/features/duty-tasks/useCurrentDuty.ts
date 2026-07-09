import { useEffect, useRef, useState } from 'react'

import { ApiError, completeTask, getCurrentDuty, openTask, returnTask, takeTask, verifyTask } from './api'
import type { ResidentCurrentDuty } from './types'
import {
    preserveTaskOrder,
    sortTasksForInitialDisplay,
    toErrorMessage,
} from './utils'

function applyInitialTaskOrdering(duty: ResidentCurrentDuty): ResidentCurrentDuty {
    return {
        ...duty,
        tasks: sortTasksForInitialDisplay(duty.tasks),
    }
}

function appendUniqueTaskId(taskIds: string[], taskId: string): string[] {
    if (taskIds.includes(taskId)) {
        return taskIds
    }

    return [...taskIds, taskId]
}

export function useCurrentDuty() {
    const [duty, setDuty] = useState<ResidentCurrentDuty | null>(null)
    const [error, setError] = useState<string | null>(null)
    const [loading, setLoading] = useState(true)
    const [pendingTaskId, setPendingTaskId] = useState<string | null>(null)
    const orderedTaskIdsRef = useRef<string[]>([])
    const initialMineTaskIdsRef = useRef<string[]>([])
    const initialFreeTaskIdsRef = useRef<string[]>([])

    async function reload() {
        setLoading(true)
        setError(null)
        setDuty(null)

        try {
            const loadedDuty = applyInitialTaskOrdering(await getCurrentDuty())
            orderedTaskIdsRef.current = loadedDuty.tasks.map((task) => task.id)
            initialMineTaskIdsRef.current = loadedDuty.tasks
                .filter((task) => task.is_mine)
                .map((task) => task.id)
            initialFreeTaskIdsRef.current = loadedDuty.tasks
                .filter((task) => !task.assignee_id)
                .map((task) => task.id)
            setDuty(loadedDuty)
        } catch (currentError) {
            if (currentError instanceof ApiError && currentError.status === 404) {
                return
            }

            setError(toErrorMessage(currentError))
        } finally {
            setLoading(false)
        }
    }

    async function runTaskAction(
        taskId: string,
        action: (currentTaskId: string) => Promise<ResidentCurrentDuty>,
    ): Promise<boolean> {
        setPendingTaskId(taskId)
        setError(null)

        try {
            const updatedDuty = await action(taskId)
            setDuty({
                ...updatedDuty,
                tasks: preserveTaskOrder(updatedDuty.tasks, orderedTaskIdsRef.current),
            })
            return true
        } catch (currentError) {
            setError(toErrorMessage(currentError))
            return false
        } finally {
            setPendingTaskId(null)
        }
    }

    async function handleTake(taskId: string) {
        const success = await runTaskAction(taskId, takeTask)
        if (!success) {
            return
        }

        initialMineTaskIdsRef.current = appendUniqueTaskId(initialMineTaskIdsRef.current, taskId)
    }

    async function handleReturn(taskId: string) {
        const success = await runTaskAction(taskId, returnTask)
        if (!success) {
            return
        }

        initialFreeTaskIdsRef.current = appendUniqueTaskId(initialFreeTaskIdsRef.current, taskId)
    }

    useEffect(() => {
        void reload()
    }, [])

    return {
        duty,
        error,
        loading,
        pendingTaskId,
        handleTake,
        handleReturn,
        handleComplete: (taskId: string) => runTaskAction(taskId, completeTask),
        handleOpen: (taskId: string) => runTaskAction(taskId, openTask),
        handleVerify: (taskId: string) => runTaskAction(taskId, verifyTask),
        initialMineTaskIds: initialMineTaskIdsRef.current,
        initialFreeTaskIds: initialFreeTaskIdsRef.current,
    }
}
