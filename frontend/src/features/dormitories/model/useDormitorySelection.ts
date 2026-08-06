import { useCallback, useEffect, useState } from 'react'

import { getDormitories } from '../api/dormitoriesApi'
import type { DormitoryListItem } from './types'

const SELECTED_DORMITORY_STORAGE_KEY = 'selected-dormitory-id'
const DORMITORIES_CHANGED_EVENT = 'dormitories:changed'
const SELECTED_DORMITORY_CHANGED_EVENT = 'selected-dormitory:changed'

export function readStoredDormitoryId(): string | null {
    return window.localStorage.getItem(SELECTED_DORMITORY_STORAGE_KEY)
}

function storeDormitoryId(dormitoryId: string | null) {
    if (dormitoryId === null) {
        window.localStorage.removeItem(SELECTED_DORMITORY_STORAGE_KEY)
        return
    }

    window.localStorage.setItem(SELECTED_DORMITORY_STORAGE_KEY, dormitoryId)
}

export function setStoredDormitoryId(dormitoryId: string | null) {
    storeDormitoryId(dormitoryId)
    notifySelectedDormitoryChanged()
}

export function notifyDormitoriesChanged() {
    window.dispatchEvent(new CustomEvent(DORMITORIES_CHANGED_EVENT))
}

function notifySelectedDormitoryChanged() {
    window.dispatchEvent(new CustomEvent(SELECTED_DORMITORY_CHANGED_EVENT))
}

export function useSelectedDormitoryId() {
    const [selectedDormitoryId, setSelectedDormitoryId] = useState<string | null>(() =>
        readStoredDormitoryId(),
    )

    useEffect(() => {
        function syncSelectedDormitory() {
            setSelectedDormitoryId(readStoredDormitoryId())
        }

        window.addEventListener(SELECTED_DORMITORY_CHANGED_EVENT, syncSelectedDormitory)
        window.addEventListener(DORMITORIES_CHANGED_EVENT, syncSelectedDormitory)
        window.addEventListener('storage', syncSelectedDormitory)

        return () => {
            window.removeEventListener(SELECTED_DORMITORY_CHANGED_EVENT, syncSelectedDormitory)
            window.removeEventListener(DORMITORIES_CHANGED_EVENT, syncSelectedDormitory)
            window.removeEventListener('storage', syncSelectedDormitory)
        }
    }, [])

    return selectedDormitoryId
}

export function useDormitorySelection(enabled: boolean) {
    const [dormitories, setDormitories] = useState<DormitoryListItem[]>([])
    const [selectedDormitoryId, setSelectedDormitoryIdState] = useState<string | null>(null)
    const [loading, setLoading] = useState(false)

    const setSelectedDormitoryId = useCallback((dormitoryId: string | null) => {
        setSelectedDormitoryIdState(dormitoryId)
        setStoredDormitoryId(dormitoryId)
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
