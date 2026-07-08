import { useEffect, useRef, useState } from 'react'

import { completeTask, getCurrentDuty, openTask, returnTask, takeTask } from './api'
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

export function useCurrentDuty() {
    const [duty, setDuty] = useState<ResidentCurrentDuty | null>(null)
    const [error, setError] = useState<string | null>(null)
    const [loading, setLoading] = useState(true)
    const [pendingTaskId, setPendingTaskId] = useState<string | null>(null)
    const orderedTaskIdsRef = useRef<string[]>([])

    async function reload() {
        setLoading(true)
        setError(null)

        try {
            const loadedDuty = applyInitialTaskOrdering(await getCurrentDuty())
            orderedTaskIdsRef.current = loadedDuty.tasks.map((task) => task.id)
            setDuty(loadedDuty)
        } catch (currentError) {
            setError(toErrorMessage(currentError))
        } finally {
            setLoading(false)
        }
    }

    async function runTaskAction(
        taskId: string,
        action: (currentTaskId: string) => Promise<ResidentCurrentDuty>,
    ) {
        setPendingTaskId(taskId)
        setError(null)

        try {
            const updatedDuty = await action(taskId)
            setDuty({
                ...updatedDuty,
                tasks: preserveTaskOrder(updatedDuty.tasks, orderedTaskIdsRef.current),
            })
        } catch (currentError) {
            setError(toErrorMessage(currentError))
        } finally {
            setPendingTaskId(null)
        }
    }

    useEffect(() => {
        void reload()
    }, [])

    return {
        duty,
        error,
        loading,
        pendingTaskId,
        handleTake: (taskId: string) => runTaskAction(taskId, takeTask),
        handleReturn: (taskId: string) => runTaskAction(taskId, returnTask),
        handleComplete: (taskId: string) => runTaskAction(taskId, completeTask),
        handleOpen: (taskId: string) => runTaskAction(taskId, openTask),
    }
}
