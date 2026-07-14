import { useEffect, useState } from 'react'

import { getCurrentUser } from '../api/currentUserApi'

type CurrentUserState = Awaited<ReturnType<typeof getCurrentUser>>

function toErrorMessage(error: unknown): string {
    if (error instanceof Error && error.message.trim().length > 0) {
        return error.message
    }

    return 'Не удалось загрузить текущего пользователя'
}

export function useCurrentUserState() {
    const [currentUser, setCurrentUser] = useState<CurrentUserState | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        let active = true

        async function loadCurrentUser() {
            setLoading(true)
            setError(null)

            try {
                const user = await getCurrentUser()
                if (!active) {
                    return
                }

                setCurrentUser(user)
            } catch (currentError) {
                if (!active) {
                    return
                }

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
    }, [])

    return {
        currentUser,
        loading,
        error,
    }
}
