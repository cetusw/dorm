import { apiRequest } from '../../../shared/api/apiClient'
import type {
    CreatePenaltyRequest,
    CurrentUserPenaltiesResponse,
    PenaltyItem,
    PenaltyResidentDetailsResponse,
    PenaltyResidentOption,
    PenaltyResidentOptionsResponse,
    PenaltyResidentSummary,
    PenaltyResidentsResponse,
} from '../model/types'

function normalizePenaltyResidentSummary(item: Record<string, unknown>): PenaltyResidentSummary {
    return {
        user_id: String(item.user_id ?? ''),
        full_name: String(item.full_name ?? ''),
        total_weight: Number(item.total_weight ?? 0),
        threshold_reached: Boolean(item.threshold_reached),
    }
}

function normalizePenaltyResidentOption(item: Record<string, unknown>): PenaltyResidentOption {
    return {
        id: String(item.id ?? ''),
        name: String(item.name ?? ''),
    }
}

function normalizePenaltyItem(item: Record<string, unknown>): PenaltyItem {
    return {
        id: String(item.id ?? ''),
        reason: String(item.reason ?? ''),
        weight: Number(item.weight ?? 0),
        issued_on: String(item.issued_on ?? ''),
    }
}

export async function getPenaltyResidents(): Promise<PenaltyResidentsResponse> {
    const response = await apiRequest('/api/v1/penalties')
    const payload = await response.json() as Record<string, unknown>

    return {
        residents: Array.isArray(payload.residents)
            ? payload.residents.map((item) => normalizePenaltyResidentSummary(item as Record<string, unknown>))
            : [],
    }
}

export async function searchPenaltyResidents(
    query: string,
): Promise<PenaltyResidentOptionsResponse> {
    const params = new URLSearchParams()

    if (query.trim() !== '') {
        params.set('q', query.trim())
    }

    const path = params.toString() === ''
        ? '/api/v1/penalties/residents'
        : `/api/v1/penalties/residents?${params.toString()}`

    const response = await apiRequest(path)
    const payload = await response.json() as Record<string, unknown>

    return {
        residents: Array.isArray(payload.residents)
            ? payload.residents.map((item) => normalizePenaltyResidentOption(item as Record<string, unknown>))
            : [],
    }
}

export async function getResidentPenalties(
    userId: string,
): Promise<PenaltyResidentDetailsResponse> {
    const response = await apiRequest(`/api/v1/penalties/residents/${encodeURIComponent(userId)}`)
    const payload = await response.json() as Record<string, unknown>

    return {
        user_id: String(payload.user_id ?? ''),
        full_name: String(payload.full_name ?? ''),
        penalties: Array.isArray(payload.penalties)
            ? payload.penalties.map((item) => normalizePenaltyItem(item as Record<string, unknown>))
            : [],
    }
}

export async function getCurrentUserPenalties(): Promise<CurrentUserPenaltiesResponse> {
    const response = await apiRequest('/api/v1/penalties/me')
    const payload = await response.json() as Record<string, unknown>

    return {
        penalties: Array.isArray(payload.penalties)
            ? payload.penalties.map((item) => normalizePenaltyItem(item as Record<string, unknown>))
            : [],
    }
}

export async function createPenalty(
    request: CreatePenaltyRequest,
): Promise<PenaltyItem> {
    const response = await apiRequest('/api/v1/penalties', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizePenaltyItem(await response.json() as Record<string, unknown>)
}

export async function resolvePenalty(
    penaltyId: string,
): Promise<void> {
    await apiRequest(`/api/v1/penalties/${encodeURIComponent(penaltyId)}`, {
        method: 'DELETE',
    })
}
