import { useCallback, useEffect, useRef, useState } from 'react'

import { ApiError } from '../../../shared/api/ApiError'
import {
    completeTask,
    getCurrentDuty,
    openTask,
    reopenTask,
    returnTask,
    TaskAlreadyAssignedError,
    takeTask,
    verifyTask,
} from '../api/currentDutyApi'
import type { CurrentDutyNotification, ResidentCurrentDuty, ResidentDutyTask } from './types'
import { setStoredDormitoryId } from '../../dormitories/model/useDormitorySelection'
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

function syncGroupIdInUrl(groupId: string) {
    const params = new URLSearchParams(window.location.search)
    if (params.get('group_id') === groupId) {
        return
    }

    params.set('group_id', groupId)
    const suffix = params.toString()
    window.history.replaceState(window.history.state, '', `${window.location.pathname}?${suffix}`)
}

function countTeamMembersBelowGoal(duty: ResidentCurrentDuty): number {
    const takenCostByMemberId = new Map<string, number>()

    duty.tasks.forEach((task) => {
        if (!task.assignee_id) {
            return
        }

        takenCostByMemberId.set(
            task.assignee_id,
            (takenCostByMemberId.get(task.assignee_id) ?? 0) + task.cost,
        )
    })

    return duty.team_members.filter(
        (member) => (takenCostByMemberId.get(member.id) ?? 0) < duty.cost_per_resident_goal,
    ).length
}

const TASK_ASSIGNED_ERROR_MESSAGE = 'эту задачу уже взял другой пользователь'

type TaskActionResult = {
    updatedDuty: ResidentCurrentDuty | null
    currentError: unknown | null
}

function replaceDutyTask(
    currentDuty: ResidentCurrentDuty,
    updatedTask: ResidentDutyTask,
): ResidentCurrentDuty {
    return {
        ...currentDuty,
        tasks: preserveTaskOrder(
            currentDuty.tasks.map((task) => (task.id === updatedTask.id ? updatedTask : task)),
            currentDuty.tasks.map((task) => task.id),
        ),
    }
}

export function useCurrentDuty(initialGroupId?: string, selectedDormitoryId?: string | null) {
    const [duty, setDuty] = useState<ResidentCurrentDuty | null>(null)
    const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null)
    const [error, setError] = useState<string | null>(null)
    const [loading, setLoading] = useState(true)
    const [pendingTaskId, setPendingTaskId] = useState<string | null>(null)
    const [notification, setNotification] = useState<CurrentDutyNotification | null>(null)
    const orderedTaskIdsRef = useRef<string[]>([])
    const visibleMineTaskIdsRef = useRef<string[]>([])
    const visibleFreeTaskIdsRef = useRef<string[]>([])
    const notificationIdRef = useRef(0)
    const pendingDormitoryChangeRef = useRef<{ dormitoryId: string, groupId?: string } | null>(null)

    function showNotification(payload: Omit<CurrentDutyNotification, 'id'>) {
        notificationIdRef.current += 1
        setNotification({
            id: notificationIdRef.current,
            ...payload,
        })
    }

    const clearNotification = useCallback(() => {
        setNotification(null)
    }, [])

    function applyLoadedDuty(loadedDuty: ResidentCurrentDuty) {
        orderedTaskIdsRef.current = loadedDuty.tasks.map((task) => task.id)
        visibleMineTaskIdsRef.current = loadedDuty.tasks
            .filter((task) => task.is_mine)
            .map((task) => task.id)
        visibleFreeTaskIdsRef.current = loadedDuty.tasks
            .filter((task) => !task.assignee_id)
            .map((task) => task.id)
        setSelectedGroupId(loadedDuty.selected_group_id)
        if (loadedDuty.selected_group_id) {
            syncGroupIdInUrl(loadedDuty.selected_group_id)
        }
        setDuty(loadedDuty)
    }

    async function reload(groupId?: string, dormitoryId: string | null = selectedDormitoryId ?? null) {
        setLoading(true)
        setError(null)
        setDuty(null)

        try {
            const loadedDuty = applyInitialTaskOrdering(await getCurrentDuty(groupId, dormitoryId ?? undefined))
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
        action: (currentTaskId: string, groupId?: string, dormitoryId?: string) => Promise<ResidentCurrentDuty>,
    ): Promise<TaskActionResult> {
        setPendingTaskId(taskId)
        setError(null)

        try {
            const updatedDuty = await action(taskId, selectedGroupId ?? undefined, selectedDormitoryId ?? undefined)
            applyLoadedDuty({
                ...updatedDuty,
                tasks: preserveTaskOrder(updatedDuty.tasks, orderedTaskIdsRef.current),
            })
            return { updatedDuty, currentError: null }
        } catch (currentError) {
            if (
                currentError instanceof ApiError &&
                currentError.status === 409 &&
                currentError.message !== TASK_ASSIGNED_ERROR_MESSAGE
            ) {
                await reload(selectedGroupId ?? undefined)
            }

            return { updatedDuty: null, currentError }
        } finally {
            setPendingTaskId(null)
        }
    }

    const handleTake: TaskActionHandler = async (taskId) => {
        const previousDuty = duty
        const { updatedDuty, currentError } = await runTaskAction(taskId, takeTask)
        if (!updatedDuty) {
            if (
                currentError instanceof TaskAlreadyAssignedError &&
                currentError.message === TASK_ASSIGNED_ERROR_MESSAGE
            ) {
                setDuty((currentDuty) => (
                    currentDuty
                        ? replaceDutyTask(currentDuty, currentError.task)
                        : currentDuty
                ))
                visibleFreeTaskIdsRef.current = visibleFreeTaskIdsRef.current.filter(
                    (visibleTaskId) => visibleTaskId !== currentError.task.id,
                )
                showNotification({
                    color: '#991B1B',
                    title: 'Задача занята',
                    message: 'Эту задачу уже взял другой участник команды.',
                })
                return false
            }

            setError(toErrorMessage(currentError))
            return false
        }

        if (
            previousDuty &&
            previousDuty.my_taken_cost_sum < previousDuty.cost_per_resident_goal &&
            updatedDuty.my_taken_cost_sum >= updatedDuty.cost_per_resident_goal &&
            countTeamMembersBelowGoal(previousDuty) > 1
        ) {
            showNotification({
                color: '#166534',
                title: 'Взято достаточно задач',
                message: 'Вы взяли задач на достаточное количество баллов, но можете продолжить брать задачи',
            })
        }

        visibleMineTaskIdsRef.current = appendUniqueTaskId(visibleMineTaskIdsRef.current, taskId)
        return true
    }

    const handleReturn: TaskActionHandler = async (taskId) => {
        const { updatedDuty, currentError } = await runTaskAction(taskId, returnTask)
        if (!updatedDuty) {
            setError(toErrorMessage(currentError))
            return false
        }

        visibleFreeTaskIdsRef.current = appendUniqueTaskId(visibleFreeTaskIdsRef.current, taskId)
        return true
    }

    useEffect(() => {
        const pendingDormitoryChange = pendingDormitoryChangeRef.current
        const nextGroupId = pendingDormitoryChange && pendingDormitoryChange.dormitoryId === selectedDormitoryId
            ? pendingDormitoryChange.groupId
            : initialGroupId

        if (pendingDormitoryChange && pendingDormitoryChange.dormitoryId === selectedDormitoryId) {
            pendingDormitoryChangeRef.current = null
        }

        void reload(nextGroupId)
    }, [initialGroupId, selectedDormitoryId])

    return {
        selectedGroupId,
        duty,
        error,
        loading,
        pendingTaskId,
        selectGroup: (groupId: string) => reload(groupId),
        selectGroupForDormitory: (groupId: string, dormitoryId: string) => {
            if (selectedDormitoryId === dormitoryId) {
                void reload(groupId, dormitoryId)
                return
            }

            pendingDormitoryChangeRef.current = { dormitoryId, groupId }
            setStoredDormitoryId(dormitoryId)
        },
        handleTake,
        handleReturn,
        handleComplete: (async (taskId: string) => {
            const { updatedDuty, currentError } = await runTaskAction(taskId, completeTask)
            if (!updatedDuty) {
                setError(toErrorMessage(currentError))
            }
            return Boolean(updatedDuty)
        }) satisfies TaskActionHandler,
        handleOpen: (async (taskId: string) => {
            const { updatedDuty, currentError } = await runTaskAction(taskId, openTask)
            if (!updatedDuty) {
                setError(toErrorMessage(currentError))
            }
            return Boolean(updatedDuty)
        }) satisfies TaskActionHandler,
        handleReopen: (async (taskId: string) => {
            const { updatedDuty, currentError } = await runTaskAction(taskId, reopenTask)
            if (!updatedDuty) {
                setError(toErrorMessage(currentError))
            }
            return Boolean(updatedDuty)
        }) satisfies TaskActionHandler,
        handleVerify: (async (taskId: string) => {
            const { updatedDuty, currentError } = await runTaskAction(taskId, verifyTask)
            if (!updatedDuty) {
                setError(toErrorMessage(currentError))
            }
            return Boolean(updatedDuty)
        }) satisfies TaskActionHandler,
        reloadCurrentDuty: () => reload(selectedGroupId ?? undefined),
        notification,
        clearNotification,
        visibleMineTaskIds: visibleMineTaskIdsRef.current,
        visibleFreeTaskIds: visibleFreeTaskIdsRef.current,
    }
}
