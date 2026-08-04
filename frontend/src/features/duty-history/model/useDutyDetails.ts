import { useEffect, useState } from 'react'

import type { ResidentDutyDetails } from '../../current-duty/model/types'
import { sortTasksForInitialDisplay, toErrorMessage } from '../../current-duty/model/utils'
import { getDutyDetails } from '../api/dutyHistoryApi'

export function useDutyDetails(dutyId: string) {
    const [duty, setDuty] = useState<ResidentDutyDetails | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    function applyLoadedDuty(loadedDuty: ResidentDutyDetails) {
        const normalizedDuty = {
            ...loadedDuty,
            tasks: sortTasksForInitialDisplay(loadedDuty.tasks),
        }
        setDuty(normalizedDuty)
    }

    async function reload() {
        setLoading(true)
        setError(null)
        try {
            applyLoadedDuty(await getDutyDetails(dutyId))
        } catch (currentError) {
            setError(toErrorMessage(currentError))
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        void reload()
    }, [dutyId])

    return {
        duty,
        loading,
        error,
        reload,
    }
}
