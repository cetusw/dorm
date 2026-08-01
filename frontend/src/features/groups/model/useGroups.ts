import { useCallback, useEffect, useState } from 'react'

import { getGroups } from '../api/groupsApi'
import type { GroupListItem } from './types'

function toErrorMessage(error: unknown): string {
    if (error instanceof Error) {
        return error.message
    }

    return 'Не удалось загрузить список групп'
}

export function useGroups(selectedDormitoryId: string | null) {
    const [groups, setGroups] = useState<GroupListItem[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const reload = useCallback(async () => {
        if (!selectedDormitoryId) {
            setGroups([])
            setError(null)
            setLoading(false)
            return
        }

        setLoading(true)
        setError(null)

        try {
            const response = await getGroups(selectedDormitoryId)
            setGroups(response.groups)
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
        groups,
        loading,
        error,
        reload,
    }
}
