import { apiRequest } from '../../../shared/api/apiClient'
import type { CurrentUser } from '../model/types'

function normalizeCurrentUser(response: Record<string, unknown>): CurrentUser {
    return {
        id: String(response.id ?? ''),
        first_name: String(response.first_name ?? ''),
        last_name: String(response.last_name ?? ''),
        can_manage_dormitories: Boolean(response.can_manage_dormitories),
        can_manage_penalties: Boolean(response.can_manage_penalties),
        can_manage_warehouse: Boolean(response.can_manage_warehouse),
    }
}

export async function getCurrentUser(): Promise<CurrentUser> {
    const response = await apiRequest('/api/v1/auth/me')
    return normalizeCurrentUser(await response.json())
}
