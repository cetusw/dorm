import { useEffect, useState } from 'react'

import { ApiError } from '../../../shared/api/ApiError'
import { getAuthenticationVersion } from '../../../shared/api/apiClient'
import { getCurrentUser } from '../api/currentUserApi'

type CurrentUserState = Awaited<ReturnType<typeof getCurrentUser>>

function toErrorMessage(error: unknown): string {
    if (error instanceof ApiError && error.message.trim().length > 0) {
        return error.message
    }

    return 'Не удалось проверить сессию. Проверьте соединение и повторите попытку.'
}

export function useCurrentUserState(enabled = true) {
    const [currentUser, setCurrentUser] = useState<CurrentUserState | null>(null)
    const [loading, setLoading] = useState(enabled)
    const [error, setError] = useState<string | null>(null)

    const [resultVersion, setResultVersion] = useState(-1)
    const [attempt, setAttempt] = useState(0)
    const [unauthenticated, setUnauthenticated] = useState(false)

    useEffect(() => {
        if (!enabled) {
            return
        }

        let active = true
        const version = getAuthenticationVersion()

        async function loadCurrentUser() {
            setLoading(true)
            setError(null)
            setUnauthenticated(false)
            setCurrentUser(null)

            try {
                const user = await getCurrentUser()
                if (!active || version !== getAuthenticationVersion()) {
                    return
                }

                setResultVersion(version)
                setCurrentUser(user)
            } catch (currentError) {
                if (!active || version !== getAuthenticationVersion()) {
                    return
                }

                setResultVersion(version)
                setUnauthenticated(currentError instanceof ApiError && currentError.status === 401)
                setError(toErrorMessage(currentError))
            } finally {
                if (active) {
                    setLoading(false)
                }
            }
        }

        void loadCurrentUser()

        return () => {
            active = false
        }
    }, [enabled, attempt])

    const fresh = resultVersion === getAuthenticationVersion()
    return {
        currentUser: enabled && fresh ? currentUser : null,
        loading: enabled && (loading || !fresh),
        error: fresh ? error : null,
        unauthenticated: enabled && fresh && unauthenticated,
        retry: () => setAttempt(value => value + 1),
    }
}
