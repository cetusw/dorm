import { apiRequest } from '../../../shared/api/apiClient'
import type { ResidentListResponse } from '../model/types'

export async function getResidents(dormitoryId: string): Promise<ResidentListResponse> {
    const response = await apiRequest(`/api/v1/users?dormitory_id=${encodeURIComponent(dormitoryId)}`)
    return response.json()
}
