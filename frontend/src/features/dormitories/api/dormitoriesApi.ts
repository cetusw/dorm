import { apiRequest } from '../../../shared/api/apiClient'
import type {
    CreateDormitoryRequest,
    DormitoryDetails,
    DormitoryListResponse,
    UpdateDormitoryRequest,
    UserOptionsResponse,
} from '../model/types'

function normalizeDormitoryList(response: DormitoryListResponse): DormitoryListResponse {
    return {
        dormitories: Array.isArray(response.dormitories) ? response.dormitories : [],
    }
}

export async function getDormitories(): Promise<DormitoryListResponse> {
    const response = await apiRequest('/api/v1/dormitories')
    return normalizeDormitoryList(await response.json())
}

function normalizeDormitoryDetails(dormitory: Record<string, unknown>): DormitoryDetails {
    return {
        id: Number(dormitory.id),
        name: String(dormitory.name ?? ''),
        city: String(dormitory.city ?? ''),
        streetType: String(dormitory.street_type ?? ''),
        streetName: String(dormitory.street_name ?? ''),
        houseNumber: String(dormitory.house_number ?? ''),
        leader: dormitory.leader && typeof dormitory.leader === 'object'
            ? {
                id: String((dormitory.leader as Record<string, unknown>).id ?? ''),
                name: String((dormitory.leader as Record<string, unknown>).name ?? ''),
            }
            : null,
    }
}

function normalizeUserOptions(response: UserOptionsResponse): UserOptionsResponse {
    return {
        users: Array.isArray(response.users) ? response.users : [],
    }
}

export async function getDormitory(dormitoryId: number): Promise<DormitoryDetails> {
    const response = await apiRequest(`/api/v1/dormitories/${dormitoryId}`)
    return normalizeDormitoryDetails(await response.json())
}

export async function getUserOptions(): Promise<UserOptionsResponse> {
    const response = await apiRequest('/api/v1/users/options')
    return normalizeUserOptions(await response.json())
}

export async function createDormitory(request: CreateDormitoryRequest): Promise<DormitoryDetails> {
    const response = await apiRequest('/api/v1/dormitories', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeDormitoryDetails(await response.json())
}

export async function updateDormitory(
    dormitoryId: number,
    request: UpdateDormitoryRequest,
): Promise<DormitoryDetails> {
    const response = await apiRequest(`/api/v1/dormitories/${dormitoryId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeDormitoryDetails(await response.json())
}

export async function deleteDormitory(dormitoryId: number): Promise<void> {
    await apiRequest(`/api/v1/dormitories/${dormitoryId}`, {
        method: 'DELETE',
    })
}
