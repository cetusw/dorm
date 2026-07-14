import { useEffect, useState } from 'react'

import type { DutyTaskSelect } from './types'

const ACTIVE_SELECT_STORAGE_KEY = 'current-duty-active-select'

function readStoredSelect(): DutyTaskSelect {
    if (typeof window === 'undefined') {
        return 'mine'
    }

    const value = window.localStorage.getItem(ACTIVE_SELECT_STORAGE_KEY)
    if (value === 'mine' || value === 'free' || value === 'all' || value === 'review') {
        return value
    }

    return 'mine'
}

export function useStoredDutySelect(visibleSelects: DutyTaskSelect[]) {
    const [activeSelect, setActiveSelect] = useState<DutyTaskSelect>(readStoredSelect)

    useEffect(() => {
        window.localStorage.setItem(ACTIVE_SELECT_STORAGE_KEY, activeSelect)
    }, [activeSelect])

    useEffect(() => {
        if (visibleSelects.length === 0) {
            return
        }

        if (!visibleSelects.includes(activeSelect)) {
            setActiveSelect(visibleSelects[0])
        }
    }, [activeSelect, visibleSelects])

    return [activeSelect, setActiveSelect] as const
}
