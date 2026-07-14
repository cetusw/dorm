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

    useEffect(() => {
        let active = true

        async function loadDormitories() {
            setLoading(true)
            setError(null)

            try {
                const response = await getDormitories()
                if (!active) {
                    return
                }

                setDormitories(response.dormitories)
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

        void loadDormitories()

        return () => {
            active = false
        }
    }, [])

    return {
        dormitories,
        loading,
        error,
    }
}
