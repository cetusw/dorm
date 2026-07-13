import { ApiError } from './ApiError'

export async function apiRequest(path: string, options: RequestInit = {}): Promise<Response> {
    const response = await fetch(path, {
        ...options,
        credentials: 'same-origin',
        headers: options.headers,
    })

    if (response.ok) {
        return response
    }

    const body = await response.json().catch(() => null)
    const message = body?.error ?? 'Ошибка запроса'

    if (response.status === 401) {
        window.location.assign('/app/login')
    }

    throw new ApiError(message, response.status)
}
