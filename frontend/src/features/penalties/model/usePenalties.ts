import { useCallback, useEffect, useState } from 'react'

import { getPenaltyResidents } from '../api/penaltiesApi'
import type { PenaltyResidentSummary } from './types'

function toErrorMessage(error: unknown): string {
    if (error instanceof Error && error.message.trim() !== '') {
        return error.message
    }

    return 'Не удалось выполнить запрос'
}

function isSameResident(
    left: PenaltyResidentSummary,
    right: PenaltyResidentSummary,
): boolean {
    return left.user_id === right.user_id
        && left.full_name === right.full_name
        && left.total_weight === right.total_weight
		&& left.individual_task_count === right.individual_task_count
        && left.threshold_reached === right.threshold_reached
}

function reconcileResidents(
    currentResidents: PenaltyResidentSummary[],
    nextResidents: PenaltyResidentSummary[],
): PenaltyResidentSummary[] {
    const currentByID = new Map(currentResidents.map((resident) => [resident.user_id, resident]))
    const reconciledResidents = nextResidents.map((resident) => {
        const currentResident = currentByID.get(resident.user_id)

        return currentResident && isSameResident(currentResident, resident)
            ? currentResident
            : resident
    })

    const unchanged = currentResidents.length === reconciledResidents.length
        && currentResidents.every((resident, index) => resident === reconciledResidents[index])

    return unchanged ? currentResidents : reconciledResidents
}

export function usePenalties() {
    const [residents, setResidents] = useState<PenaltyResidentSummary[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    const synchronizeResidents = useCallback(async () => {
        try {
            const response = await getPenaltyResidents()
            setResidents((currentResidents) => reconcileResidents(currentResidents, response.residents))
            setError(null)
        } catch {
            // The current list remains usable if background synchronization fails.
        }
    }, [])

    const loadInitialResidents = useCallback(async () => {
        setLoading(true)
        setError(null)

        try {
            const response = await getPenaltyResidents()
            setResidents(response.residents)
        } catch (currentError) {
            setError(toErrorMessage(currentError))
            setResidents([])
        } finally {
            setLoading(false)
        }
    }, [])

    useEffect(() => {
        const timeoutId = window.setTimeout(() => {
            void loadInitialResidents()
        }, 0)

        return () => {
            window.clearTimeout(timeoutId)
        }
    }, [loadInitialResidents])

    return {
        residents,
        loading,
        error,
        synchronizeResidents,
    }
}
