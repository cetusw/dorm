import { ApiError } from '../../../shared/api/ApiError'
import { apiRequest } from '../../../shared/api/apiClient'
import type { ResidentCurrentDuty, ResidentDutyTask } from '../model/types'

type TaskAssignedConflictPayload = {
    task?: ResidentDutyTask
}

export class TaskAlreadyAssignedError extends ApiError {
    readonly task: ResidentDutyTask

    constructor(message: string, status: number, task: ResidentDutyTask, details?: unknown) {
        super(message, status, details)
        this.name = 'TaskAlreadyAssignedError'
        this.task = task
    }
}

function normalizeDutyTask(task: ResidentDutyTask): ResidentDutyTask {
    return {
        ...task,
        is_mine: Boolean(task.is_mine),
        can_take: Boolean(task.can_take),
        can_return: Boolean(task.can_return),
        can_complete: Boolean(task.can_complete),
        can_open: Boolean(task.can_open),
        can_verify: Boolean(task.can_verify),
        can_review_open: Boolean(task.can_review_open),
    }
}

function extractTaskAssignedConflictTask(currentError: ApiError): ResidentDutyTask | null {
    const payload = currentError.details as TaskAssignedConflictPayload | null
    if (!payload?.task || typeof payload.task !== 'object') {
        return null
    }

    return normalizeDutyTask(payload.task)
}

function normalizeCurrentDuty(duty: ResidentCurrentDuty): ResidentCurrentDuty {
    return {
        ...duty,
        can_manage_duty_settings: Boolean(duty.can_manage_duty_settings),
        notice_tone: duty.notice_tone === 'warning' || duty.notice_tone === 'info' ? duty.notice_tone : '',
        visible_tabs: Array.isArray(duty.visible_tabs) ? duty.visible_tabs : [],
        groups: Array.isArray(duty.groups) ? duty.groups : [],
        team_members: Array.isArray(duty.team_members) ? duty.team_members : [],
        tasks: Array.isArray(duty.tasks) ? duty.tasks : [],
    }
}

function buildCurrentDutyPath(groupId?: string, dormitoryId?: string): string {
    const params = new URLSearchParams()
    if (groupId) {
        params.set('group_id', groupId)
    }
    if (dormitoryId) {
        params.set('dormitory_id', dormitoryId)
    }

    const suffix = params.toString()
    return suffix === '' ? '/api/v1/resident/current-duty' : `/api/v1/resident/current-duty?${suffix}`
}

export async function getCurrentDuty(groupId?: string, dormitoryId?: string): Promise<ResidentCurrentDuty> {
    const response = await apiRequest(buildCurrentDutyPath(groupId, dormitoryId))
    return normalizeCurrentDuty(await response.json())
}

async function requestDutyAction(
    taskId: string,
    action: string,
    groupId?: string,
    dormitoryId?: string,
): Promise<ResidentCurrentDuty> {
    const params = new URLSearchParams()
    if (groupId) {
        params.set('group_id', groupId)
    }
    if (dormitoryId) {
        params.set('dormitory_id', dormitoryId)
    }
    const suffix = params.toString()
    const path = suffix === ''
        ? `/api/v1/resident/tasks/${taskId}/${action}`
        : `/api/v1/resident/tasks/${taskId}/${action}?${suffix}`

    const response = await apiRequest(path, {
        method: 'POST',
    })

    return normalizeCurrentDuty(await response.json())
}

export async function takeTask(taskId: string, groupId?: string, dormitoryId?: string): Promise<ResidentCurrentDuty> {
    try {
        return await requestDutyAction(taskId, 'take', groupId, dormitoryId)
    } catch (currentError) {
        if (!(currentError instanceof ApiError) || currentError.status !== 409) {
            throw currentError
        }

        const task = extractTaskAssignedConflictTask(currentError)
        if (!task) {
            throw currentError
        }

        throw new TaskAlreadyAssignedError(currentError.message, currentError.status, task, currentError.details)
    }
}

export async function returnTask(taskId: string, groupId?: string, dormitoryId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'return', groupId, dormitoryId)
}

export async function completeTask(taskId: string, groupId?: string, dormitoryId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'complete', groupId, dormitoryId)
}

export async function openTask(taskId: string, groupId?: string, dormitoryId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'open', groupId, dormitoryId)
}

export async function reopenTask(taskId: string, groupId?: string, dormitoryId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'reopen', groupId, dormitoryId)
}

export async function verifyTask(taskId: string, groupId?: string, dormitoryId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'verify', groupId, dormitoryId)
}

export async function createGroupDuty(
    groupId: string,
    startDate: string,
    endDate: string,
): Promise<void> {
    await apiRequest(`/api/v1/groups/${groupId}/duties`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            start_date: startDate,
            end_date: endDate,
        }),
    })
}
