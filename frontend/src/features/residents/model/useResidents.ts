import { useCallback, useEffect, useState } from 'react'

import { getResidents } from '../api/residentsApi'
import type { ResidentListItem } from './types'

function toErrorMessage(error: unknown): string {
    if (error instanceof Error) {
        return error.message
    }

    return 'Не удалось загрузить список жителей'
}

export function useResidents(selectedDormitoryId: string | null) {
    const [residents, setResidents] = useState<ResidentListItem[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const reload = useCallback(async () => {
        if (!selectedDormitoryId) {
            setResidents([])
            setError(null)
            setLoading(false)
            return
        }

        setLoading(true)
        setError(null)

        try {
            const response = await getResidents(selectedDormitoryId)
            setResidents(response.users)
        } catch (error) {
            setError(toErrorMessage(error))
        } finally {
            setLoading(false)
        }
    }, [selectedDormitoryId])

    useEffect(() => {
        void reload()
    }, [reload])

    return {
        residents,
        loading,
        error,
        reload,
    }
}
