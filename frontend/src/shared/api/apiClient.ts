import { replaceTo } from '../../app/navigation'
import { ApiError } from './ApiError'

type ApiRequestOptions = RequestInit & {
    redirectOn401?: boolean
}

let authenticationVersion = 0
let authenticationChanges = 0
let sessionCheck: { version: number; promise: Promise<void> } | null = null

export function getAuthenticationVersion() {
    return authenticationVersion
}

export function beginAuthenticationChange() {
    authenticationVersion++
    authenticationChanges++
    return () => {
        authenticationVersion++
        authenticationChanges--
    }
}

async function confirmSession(version: number) {
    if (version !== authenticationVersion || authenticationChanges > 0) return
    if (!sessionCheck || sessionCheck.version !== version) {
        const check = (async () => {
            try {
                const response = await fetch('/api/v1/auth/me', { credentials: 'same-origin' })
                if (response.status === 401 && version === authenticationVersion && authenticationChanges === 0) {
                    replaceTo('/app/login')
                }
            } catch {
            }
        })()
        sessionCheck = { version, promise: check }
        void check.finally(() => {
            if (sessionCheck?.promise === check) sessionCheck = null
        })
    }
    await sessionCheck.promise
}

export async function apiRequest(path: string, options: ApiRequestOptions = {}): Promise<Response> {
    const version = authenticationVersion
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

    if (response.status === 401 && redirectOn401 && path !== '/api/v1/auth/me') {
        void confirmSession(version)
    }

    throw new ApiError(message, response.status, body)
}
