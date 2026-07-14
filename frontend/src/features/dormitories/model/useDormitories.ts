import { useEffect, useState } from 'react'

import { getDormitories } from '../api/dormitoriesApi'
import type { DormitoryListItem } from './types'

function toErrorMessage(error: unknown): string {
    if (error instanceof Error && error.message.trim().length > 0) {
        return error.message
    }

    return 'Не удалось загрузить список общежитий'
}

export function useDormitories() {
    const [dormitories, setDormitories] = useState<DormitoryListItem[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    async function reload(): Promise<void> {
        setLoading(true)
        setError(null)

        try {
            const response = await getDormitories()
            setDormitories(response.dormitories)
        } catch (currentError) {
            setError(toErrorMessage(currentError))
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        void reload()
    }, [])

    return {
        dormitories,
        loading,
        error,
        reload,
    }
}
