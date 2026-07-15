import { useCallback, useEffect, useState } from 'react'

import { getDormitories } from '../api/dormitoriesApi'
import type { DormitoryListItem } from './types'

const SELECTED_DORMITORY_STORAGE_KEY = 'selected-dormitory-id'
const DORMITORIES_CHANGED_EVENT = 'dormitories:changed'

function readStoredDormitoryId(): string | null {
    return window.localStorage.getItem(SELECTED_DORMITORY_STORAGE_KEY)
}

function storeDormitoryId(dormitoryId: string | null) {
    if (dormitoryId === null) {
        window.localStorage.removeItem(SELECTED_DORMITORY_STORAGE_KEY)
        return
    }

    window.localStorage.setItem(SELECTED_DORMITORY_STORAGE_KEY, dormitoryId)
}

export function notifyDormitoriesChanged() {
    window.dispatchEvent(new CustomEvent(DORMITORIES_CHANGED_EVENT))
}

export function useDormitorySelection(enabled: boolean) {
    const [dormitories, setDormitories] = useState<DormitoryListItem[]>([])
    const [selectedDormitoryId, setSelectedDormitoryIdState] = useState<string | null>(null)
    const [loading, setLoading] = useState(false)

    const setSelectedDormitoryId = useCallback((dormitoryId: string | null) => {
        setSelectedDormitoryIdState(dormitoryId)
        storeDormitoryId(dormitoryId)
    }, [])

    const reload = useCallback(async () => {
        if (!enabled) {
            setDormitories([])
            setSelectedDormitoryId(null)
            setLoading(false)
            return
        }

        setLoading(true)

        try {
            const response = await getDormitories()
            const loadedDormitories = response.dormitories
            const storedDormitoryId = readStoredDormitoryId()
            const resolvedDormitoryId = loadedDormitories.some(
                (dormitory) => String(dormitory.id) === storedDormitoryId,
            )
                ? storedDormitoryId
                : loadedDormitories[0]
                    ? String(loadedDormitories[0].id)
                    : null

            setDormitories(loadedDormitories)
            setSelectedDormitoryId(resolvedDormitoryId)
        } finally {
            setLoading(false)
        }
    }, [enabled, setSelectedDormitoryId])

    useEffect(() => {
        void reload()
    }, [reload])

    useEffect(() => {
        if (!enabled) {
            return
        }

        function handleDormitoriesChanged() {
            void reload()
        }

        window.addEventListener(DORMITORIES_CHANGED_EVENT, handleDormitoriesChanged)

        return () => {
            window.removeEventListener(DORMITORIES_CHANGED_EVENT, handleDormitoriesChanged)
        }
    }, [enabled, reload])

    return {
        dormitories,
        loading,
        selectedDormitoryId,
        setSelectedDormitoryId,
    }
}
