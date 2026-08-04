import { useEffect, useState } from 'react'

import { ApiError } from '../../../shared/api/ApiError'
import type { DutyHistoryResponse } from '../../current-duty/model/types'
import { toErrorMessage } from '../../current-duty/model/utils'
import { getDutyHistory } from '../api/dutyHistoryApi'

function syncHistoryGroupUrl(groupId: string) {
    const params = new URLSearchParams(window.location.search)
    if (params.get('group_id') === groupId) {
        return
    }
    params.set('group_id', groupId)
    window.history.replaceState(window.history.state, '', `${window.location.pathname}?${params.toString()}`)
}

export function useDutyHistory(initialGroupId?: string) {
    const [data, setData] = useState<DutyHistoryResponse | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    async function load(groupId?: string) {
        setLoading(true)
        setError(null)
        try {
            const response = await getDutyHistory(groupId)
            if (response.selected_group_id) {
                syncHistoryGroupUrl(response.selected_group_id)
            }
            setData(response)
        } catch (currentError) {
            if (currentError instanceof ApiError && currentError.status === 404) {
                setData(null)
                return
            }
            setError(toErrorMessage(currentError))
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        void load(initialGroupId)
    }, [initialGroupId])

    return {
        data,
        loading,
        error,
        selectGroup: (groupId: string) => load(groupId),
    }
}
