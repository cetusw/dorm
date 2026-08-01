import { useCallback, useEffect, useState } from 'react'

import { getAreas } from '../api/areasApi'
import type { AreaListItem } from './types'

function toErrorMessage(error: unknown): string {
    if (error instanceof Error) {
        return error.message
    }

    return 'Не удалось загрузить список территорий'
}

export function useAreas(selectedDormitoryId: string | null) {
    const [areas, setAreas] = useState<AreaListItem[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const reload = useCallback(async () => {
        if (!selectedDormitoryId) {
            setAreas([])
            setLoading(false)
            setError(null)
            return
        }

        setLoading(true)
        setError(null)

        try {
            const response = await getAreas(selectedDormitoryId)
            setAreas(response.areas)
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
        areas,
        loading,
        error,
        reload,
    }
}
