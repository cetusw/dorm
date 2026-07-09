import type { ResidentCurrentDuty } from './types'

const DEV_USER_ID = '0x2E614C967FB9461AA03FA54E0DBA1641'

function hexToUuid(hex: string): string {
    const value = hex.replace(/^0x/i, '').toLowerCase();

    if (!/^[0-9a-f]{32}$/.test(value)) {
        throw new Error('Invalid UUID hex string');
    }

    return [
        value.slice(0, 8),
        value.slice(8, 12),
        value.slice(12, 16),
        value.slice(16, 20),
        value.slice(20),
    ].join('-');
}

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
        headers: {
            'X-User-ID': hexToUuid(DEV_USER_ID),
            ...options.headers,
        },
    })

    if (!response.ok) {
        const body = await response.json().catch(() => null)
        const message = body?.error ?? 'Ошибка запроса'
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
