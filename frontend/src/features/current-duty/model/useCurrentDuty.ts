import { useEffect, useRef, useState } from 'react'

import { ApiError } from '../../../shared/api/ApiError'
import {
    completeTask,
    getCurrentDuty,
    openTask,
    returnTask,
    takeTask,
    verifyTask,
} from '../api/currentDutyApi'
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
    const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null)
    const [error, setError] = useState<string | null>(null)
    const [loading, setLoading] = useState(true)
    const [pendingTaskId, setPendingTaskId] = useState<string | null>(null)
    const orderedTaskIdsRef = useRef<string[]>([])
    const visibleMineTaskIdsRef = useRef<string[]>([])
    const visibleFreeTaskIdsRef = useRef<string[]>([])

    function applyLoadedDuty(loadedDuty: ResidentCurrentDuty) {
        orderedTaskIdsRef.current = loadedDuty.tasks.map((task) => task.id)
        visibleMineTaskIdsRef.current = loadedDuty.tasks
            .filter((task) => task.is_mine)
            .map((task) => task.id)
        visibleFreeTaskIdsRef.current = loadedDuty.tasks
            .filter((task) => !task.assignee_id)
            .map((task) => task.id)
        setSelectedGroupId(loadedDuty.selected_group_id)
        setDuty(loadedDuty)
    }

    async function reload(groupId?: string) {
        setLoading(true)
        setError(null)
        setDuty(null)

        try {
            const loadedDuty = applyInitialTaskOrdering(await getCurrentDuty(groupId))
            applyLoadedDuty(loadedDuty)
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
        action: (currentTaskId: string, groupId?: string) => Promise<ResidentCurrentDuty>,
    ): Promise<boolean> {
        setPendingTaskId(taskId)
        setError(null)

        try {
            const updatedDuty = await action(taskId, selectedGroupId ?? undefined)
            applyLoadedDuty({
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

        visibleMineTaskIdsRef.current = appendUniqueTaskId(visibleMineTaskIdsRef.current, taskId)
    }

    async function handleReturn(taskId: string) {
        const success = await runTaskAction(taskId, returnTask)
        if (!success) {
            return
        }

        visibleFreeTaskIdsRef.current = appendUniqueTaskId(visibleFreeTaskIdsRef.current, taskId)
    }

    useEffect(() => {
        void reload()
    }, [])

    return {
        selectedGroupId,
        duty,
        error,
        loading,
        pendingTaskId,
        selectGroup: (groupId: string) => reload(groupId),
        handleTake,
        handleReturn,
        handleComplete: (taskId: string) => runTaskAction(taskId, completeTask),
        handleOpen: (taskId: string) => runTaskAction(taskId, openTask),
        handleVerify: (taskId: string) => runTaskAction(taskId, verifyTask),
        visibleMineTaskIds: visibleMineTaskIdsRef.current,
        visibleFreeTaskIds: visibleFreeTaskIdsRef.current,
    }
}
