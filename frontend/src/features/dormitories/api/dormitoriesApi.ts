import { apiRequest } from '../../../shared/api/apiClient'
import type { DormitoryListResponse } from '../model/types'

function normalizeDormitoryList(response: DormitoryListResponse): DormitoryListResponse {
    return {
        dormitories: Array.isArray(response.dormitories) ? response.dormitories : [],
    }
}

export async function getDormitories(): Promise<DormitoryListResponse> {
    const response = await apiRequest('/api/v1/dormitories')
    return normalizeDormitoryList(await response.json())
}
