import { useEffect, useState } from 'react'

import type { DutyTaskTab } from './types'

const ACTIVE_TAB_STORAGE_KEY = 'current-duty-active-tab'

function readStoredTab(): DutyTaskTab {
    if (typeof window === 'undefined') {
        return 'mine'
    }

    const value = window.localStorage.getItem(ACTIVE_TAB_STORAGE_KEY)
    if (value === 'mine' || value === 'free' || value === 'all' || value === 'review') {
        return value
    }

    return 'mine'
}

export function useStoredDutyTab(visibleTabs: DutyTaskTab[]) {
    const [activeTab, setActiveTab] = useState<DutyTaskTab>(readStoredTab)

    useEffect(() => {
        window.localStorage.setItem(ACTIVE_TAB_STORAGE_KEY, activeTab)
    }, [activeTab])

    useEffect(() => {
        if (visibleTabs.length === 0) {
            return
        }

        if (!visibleTabs.includes(activeTab)) {
            setActiveTab(visibleTabs[0])
        }
    }, [activeTab, visibleTabs])

    return [activeTab, setActiveTab] as const
}
