import type { ResidentCurrentDuty } from './types'

export class ApiError extends Error {
    status: number

    constructor(message: string, status: number) {
        super(message)
        this.name = 'ApiError'
        this.status = status
    }
}

async function request(path: string, options: RequestInit = {}) {
    const response = await fetch(path, {
        ...options,
        credentials: 'same-origin',
        headers: options.headers,
    })

    if (!response.ok) {
        const body = await response.json().catch(() => null)
        const message = body?.error ?? 'Ошибка запроса'
        if (response.status === 401) {
            window.location.assign('/app/login')
        }
        throw new ApiError(message, response.status)
    }

    return response
}

function buildCurrentDutyPath(groupId?: string): string {
    if (!groupId) {
        return '/api/v1/resident/current-duty'
    }

    return `/api/v1/resident/current-duty?group_id=${encodeURIComponent(groupId)}`
}

export async function getCurrentDuty(groupId?: string): Promise<ResidentCurrentDuty> {
    const response = await request(buildCurrentDutyPath(groupId))
    return response.json()
}

async function requestDutyAction(
    taskId: string,
    action: string,
    groupId?: string,
): Promise<ResidentCurrentDuty> {
    const path = groupId
        ? `/api/v1/resident/tasks/${taskId}/${action}?group_id=${encodeURIComponent(groupId)}`
        : `/api/v1/resident/tasks/${taskId}/${action}`

    const response = await request(path, {
        method: 'POST',
    })

    return response.json()
}

export async function takeTask(taskId: string, groupId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'take', groupId)
}

export async function returnTask(taskId: string, groupId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'return', groupId)
}

export async function completeTask(taskId: string, groupId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'complete', groupId)
}

export async function openTask(taskId: string, groupId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'open', groupId)
}

export async function verifyTask(taskId: string, groupId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'verify', groupId)
}
