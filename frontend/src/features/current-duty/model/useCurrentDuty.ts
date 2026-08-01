import { useEffect, useRef, useState } from 'react'

import { ApiError } from '../../../shared/api/ApiError'
import {
    completeTask,
    getCurrentDuty,
    openTask,
    reopenTask,
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
import type { TaskActionHandler } from './taskActions'

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
            if (currentError instanceof ApiError && currentError.status === 409) {
                await reload(selectedGroupId ?? undefined)
            }

            setError(toErrorMessage(currentError))
            return false
        } finally {
            setPendingTaskId(null)
        }
    }

    const handleTake: TaskActionHandler = async (taskId) => {
        const success = await runTaskAction(taskId, takeTask)
        if (!success) {
            return false
        }

        visibleMineTaskIdsRef.current = appendUniqueTaskId(visibleMineTaskIdsRef.current, taskId)
        return true
    }

    const handleReturn: TaskActionHandler = async (taskId) => {
        const success = await runTaskAction(taskId, returnTask)
        if (!success) {
            return false
        }

        visibleFreeTaskIdsRef.current = appendUniqueTaskId(visibleFreeTaskIdsRef.current, taskId)
        return true
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
        handleComplete: ((taskId: string) => runTaskAction(taskId, completeTask)) satisfies TaskActionHandler,
        handleOpen: ((taskId: string) => runTaskAction(taskId, openTask)) satisfies TaskActionHandler,
        handleReopen: ((taskId: string) => runTaskAction(taskId, reopenTask)) satisfies TaskActionHandler,
        handleVerify: ((taskId: string) => runTaskAction(taskId, verifyTask)) satisfies TaskActionHandler,
        reloadCurrentDuty: () => reload(selectedGroupId ?? undefined),
        visibleMineTaskIds: visibleMineTaskIdsRef.current,
        visibleFreeTaskIds: visibleFreeTaskIdsRef.current,
    }
}
