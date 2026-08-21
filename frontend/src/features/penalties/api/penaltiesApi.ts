import { apiRequest } from '../../../shared/api/apiClient'
import type {
    CreatePenaltyRequest,
    CurrentUserPenaltiesResponse,
    PenaltyEntryItem,
    PenaltyResidentDetailsResponse,
    PenaltyResidentOption,
    PenaltyResidentOptionsResponse,
    PenaltyResidentSummary,
    PenaltyResidentsResponse,
    ResolvePenaltyRequest,
    UpdatePenaltyEntryRequest,
} from '../model/types'

function normalizePenaltyResidentSummary(item: Record<string, unknown>): PenaltyResidentSummary {
    return {
        user_id: String(item.user_id ?? ''),
        full_name: String(item.full_name ?? ''),
        total_weight: Number(item.total_weight ?? 0),
        individual_task_count: Number(item.individual_task_count ?? 0),
        threshold_reached: Boolean(item.threshold_reached),
    }
}

function normalizePenaltyResidentOption(item: Record<string, unknown>): PenaltyResidentOption {
    return {
        id: String(item.id ?? ''),
        name: String(item.name ?? ''),
    }
}

function normalizePenaltyEntryItem(item: Record<string, unknown>): PenaltyEntryItem {
    return {
        id: String(item.id ?? ''),
        type: item.type === 'resolve' ? 'resolve' : 'issue',
        reason: String(item.reason ?? ''),
        weight: Number(item.weight ?? 0),
        created_at: String(item.created_at ?? ''),
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
        total_weight: Number(payload.total_weight ?? 0),
        entries: Array.isArray(payload.entries)
            ? payload.entries.map((item) => normalizePenaltyEntryItem(item as Record<string, unknown>))
            : [],
    }
}

export async function getCurrentUserPenalties(): Promise<CurrentUserPenaltiesResponse> {
    const response = await apiRequest('/api/v1/penalties/me')
    const payload = await response.json() as Record<string, unknown>

    return {
        total_weight: Number(payload.total_weight ?? 0),
        entries: Array.isArray(payload.entries)
            ? payload.entries.map((item) => normalizePenaltyEntryItem(item as Record<string, unknown>))
            : [],
    }
}

export async function createPenalty(
    request: CreatePenaltyRequest,
): Promise<PenaltyEntryItem> {
    const response = await apiRequest('/api/v1/penalties', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizePenaltyEntryItem(await response.json() as Record<string, unknown>)
}

export async function resolvePenalty(
    request: ResolvePenaltyRequest,
): Promise<void> {
    await apiRequest('/api/v1/penalties/resolve', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })
}

export async function deletePenaltyEntry(entryId: string): Promise<void> {
    await apiRequest(`/api/v1/penalties/${encodeURIComponent(entryId)}`, {
        method: 'DELETE',
    })
}

export async function updatePenaltyEntry(
    entryId: string,
    request: UpdatePenaltyEntryRequest,
): Promise<PenaltyEntryItem> {
    const response = await apiRequest(`/api/v1/penalties/${encodeURIComponent(entryId)}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizePenaltyEntryItem(await response.json() as Record<string, unknown>)
}
