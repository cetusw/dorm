import { useCallback, useEffect, useState } from 'react'

import { getPenaltyResidents } from '../api/penaltiesApi'
import type { PenaltyResidentSummary } from './types'

function toErrorMessage(error: unknown): string {
    if (error instanceof Error && error.message.trim() !== '') {
        return error.message
    }

    return 'Не удалось выполнить запрос'
}

export function usePenalties() {
    const [residents, setResidents] = useState<PenaltyResidentSummary[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    const reload = useCallback(async () => {
        setLoading(true)
        setError(null)

        try {
            const response = await getPenaltyResidents()
            setResidents(response.residents)
        } catch (currentError) {
            setError(toErrorMessage(currentError))
            setResidents([])
        } finally {
            setLoading(false)
        }
    }, [])

    useEffect(() => {
        const timeoutId = window.setTimeout(() => {
            void reload()
        }, 0)

        return () => {
            window.clearTimeout(timeoutId)
        }
    }, [reload])

    return {
        residents,
        loading,
        error,
        reload,
    }
}
