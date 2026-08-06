import { ApiError } from './ApiError'

type ApiRequestOptions = RequestInit & {
    redirectOn401?: boolean
}

export async function apiRequest(path: string, options: ApiRequestOptions = {}): Promise<Response> {
    const { redirectOn401 = true, ...requestInit } = options

    const response = await fetch(path, {
        ...requestInit,
        credentials: 'same-origin',
        headers: requestInit.headers,
    })

    if (response.ok) {
        return response
    }

    const body = await response.json().catch(() => null)
    const message = body?.error ?? 'Ошибка запроса'

    if (response.status === 401 && redirectOn401) {
        window.location.assign('/app/login')
    }

    throw new ApiError(message, response.status, body)
}
