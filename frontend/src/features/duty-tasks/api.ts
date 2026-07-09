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

export async function getCurrentDuty(): Promise<ResidentCurrentDuty> {
    const response = await request('/api/v1/resident/current-duty')
    return response.json()
}

async function requestDutyAction(taskId: string, action: string): Promise<ResidentCurrentDuty> {
    const response = await request(`/api/v1/resident/tasks/${taskId}/${action}`, {
        method: 'POST',
    })

    return response.json()
}

export async function takeTask(taskId: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'take')
}

export async function returnTask(taskId: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'return')
}

export async function completeTask(taskId: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'complete')
}

export async function openTask(taskId: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'open')
}

export async function verifyTask(taskId: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'verify')
}
