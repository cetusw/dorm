import type { ResidentCurrentDuty } from './types'

const DEV_USER_ID = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'

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
            'X-User-ID': DEV_USER_ID,
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
