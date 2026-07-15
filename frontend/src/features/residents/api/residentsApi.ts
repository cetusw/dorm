import { apiRequest } from '../../../shared/api/apiClient'
import type {
    CreateResidentRequest,
    ResidentDetails,
    ResidentListResponse,
    UpdateResidentRequest,
} from '../model/types'

function normalizeResidentListResponse(response: ResidentListResponse): ResidentListResponse {
    return {
        users: Array.isArray(response.users) ? response.users : [],
    }
}

function normalizeResidentDetails(resident: Record<string, unknown>): ResidentDetails {
    return {
        id: String(resident.id ?? ''),
        first_name: String(resident.first_name ?? ''),
        last_name: String(resident.last_name ?? ''),
        middle_name: resident.middle_name == null ? null : String(resident.middle_name),
        login: String(resident.login ?? ''),
        floor: resident.floor == null ? null : Number(resident.floor),
        room_number: resident.room_number == null ? null : String(resident.room_number),
    }
}

export async function getResidents(dormitoryId: string): Promise<ResidentListResponse> {
    const response = await apiRequest(`/api/v1/users?dormitory_id=${encodeURIComponent(dormitoryId)}`)
    return normalizeResidentListResponse(await response.json())
}

export async function getResident(residentId: string): Promise<ResidentDetails> {
    const response = await apiRequest(`/api/v1/users/${residentId}`)
    return normalizeResidentDetails(await response.json())
}

export async function createResident(request: CreateResidentRequest): Promise<ResidentDetails> {
    const response = await apiRequest('/api/v1/users', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeResidentDetails(await response.json())
}

export async function updateResident(
    residentId: string,
    request: UpdateResidentRequest,
): Promise<ResidentDetails> {
    const response = await apiRequest(`/api/v1/users/${residentId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeResidentDetails(await response.json())
}

export async function deleteResident(residentId: string): Promise<void> {
    await apiRequest(`/api/v1/users/${residentId}`, {
        method: 'DELETE',
    })
}
