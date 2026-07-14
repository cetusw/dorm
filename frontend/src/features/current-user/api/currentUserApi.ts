import { apiRequest } from '../../../shared/api/apiClient'
import type { CurrentUser } from '../model/types'

export async function getCurrentUser(): Promise<CurrentUser> {
    const response = await apiRequest('/api/v1/auth/me')
    return response.json()
}
