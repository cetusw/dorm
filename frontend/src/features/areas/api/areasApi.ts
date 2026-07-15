import { apiRequest } from '../../../shared/api/apiClient'
import type {
    AreaDetails,
    AreaListResponse,
    CreateAreaRequest,
    UpdateAreaRequest,
} from '../model/types'

function normalizeAreaListResponse(response: AreaListResponse): AreaListResponse {
    return {
        areas: Array.isArray(response.areas) ? response.areas : [],
    }
}

function normalizeAreaDetails(area: Record<string, unknown>): AreaDetails {
    return {
        id: Number(area.id),
        name: String(area.name ?? ''),
        floor: area.floor == null ? null : Number(area.floor),
        group: area.group && typeof area.group === 'object'
            ? {
                id: String((area.group as Record<string, unknown>).id ?? ''),
                name: String((area.group as Record<string, unknown>).name ?? ''),
            }
            : null,
    }
}

export async function getAreas(dormitoryId: string): Promise<AreaListResponse> {
    const response = await apiRequest(`/api/v1/areas?dormitory_id=${encodeURIComponent(dormitoryId)}`)
    return normalizeAreaListResponse(await response.json())
}

export async function getArea(areaId: number): Promise<AreaDetails> {
    const response = await apiRequest(`/api/v1/areas/${areaId}`)
    return normalizeAreaDetails(await response.json())
}

export async function createArea(dormitoryId: string, request: CreateAreaRequest): Promise<AreaDetails> {
    const response = await apiRequest(`/api/v1/areas?dormitory_id=${encodeURIComponent(dormitoryId)}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeAreaDetails(await response.json())
}

export async function updateArea(
    dormitoryId: string,
    areaId: number,
    request: UpdateAreaRequest,
): Promise<AreaDetails> {
    const response = await apiRequest(`/api/v1/areas/${areaId}?dormitory_id=${encodeURIComponent(dormitoryId)}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeAreaDetails(await response.json())
}

export async function deleteArea(dormitoryId: string, areaId: number): Promise<void> {
    await apiRequest(`/api/v1/areas/${areaId}?dormitory_id=${encodeURIComponent(dormitoryId)}`, {
        method: 'DELETE',
    })
}
