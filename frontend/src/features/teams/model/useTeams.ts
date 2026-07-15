import { useCallback, useEffect, useState } from 'react'

import { getTeams } from '../api/teamsApi'
import type { TeamListItem } from './types'

function toErrorMessage(error: unknown): string {
    if (error instanceof Error) {
        return error.message
    }

    return 'Не удалось загрузить список команд'
}

export function useTeams(groupId: string | null) {
    const [teams, setTeams] = useState<TeamListItem[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const reload = useCallback(async () => {
        if (!groupId) {
            setTeams([])
            setLoading(false)
            setError(null)
            return
        }

        setLoading(true)
        setError(null)

        try {
            const response = await getTeams(groupId)
            setTeams(response.teams)
        } catch (error) {
            setError(toErrorMessage(error))
        } finally {
            setLoading(false)
        }
    }, [groupId])

    useEffect(() => {
        void reload()
    }, [reload])

    return {
        teams,
        loading,
        error,
        reload,
    }
}
