import { ApiError } from '../duty-tasks/api'

type LoginResponse = {
    redirect_url: string
}

export async function loginResident(login: string, password: string): Promise<LoginResponse> {
    const response = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        credentials: 'same-origin',
        body: JSON.stringify({
            login,
            password,
        }),
    })

    if (!response.ok) {
        const body = await response.json().catch(() => null)
        const message = body?.error ?? 'Не удалось выполнить вход'
        throw new ApiError(message, response.status)
    }

    return response.json()
}
