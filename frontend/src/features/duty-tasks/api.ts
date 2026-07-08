import type { ResidentCurrentDuty } from './types'

const DEV_USER_ID = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'

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
        throw new Error(message)
    }

    return response
}

export async function getCurrentDuty(): Promise<ResidentCurrentDuty> {
    const response = await request('/api/v1/resident/current-duty')
    return response.json()
}

export async function takeTask(taskId: string): Promise<void> {
    await request(`/api/v1/resident/tasks/${taskId}/take`, {
        method: 'POST',
    })
}

export async function returnTask(taskId: string): Promise<void> {
    await request(`/api/v1/resident/tasks/${taskId}/return`, {
        method: 'POST',
    })
}

export async function completeTask(taskId: string): Promise<void> {
    await request(`/api/v1/resident/tasks/${taskId}/complete`, {
        method: 'POST',
    })
}