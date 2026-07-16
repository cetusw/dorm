import { useCallback, useEffect, useState } from 'react'

import { getTaskDefinitions } from '../api/taskCatalogApi'
import type { TaskListItem } from './types'

function toErrorMessage(error: unknown): string {
    if (error instanceof Error) {
        return error.message
    }

    return 'Не удалось загрузить список задач'
}

export function useTaskCatalog(selectedDormitoryId: string | null) {
    const [tasks, setTasks] = useState<TaskListItem[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const reload = useCallback(async () => {
        if (!selectedDormitoryId) {
            setTasks([])
            setLoading(false)
            setError(null)
            return
        }

        setLoading(true)
        setError(null)

        try {
            const response = await getTaskDefinitions(selectedDormitoryId)
            setTasks(response.tasks)
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
        tasks,
        loading,
        error,
        reload,
    }
}
