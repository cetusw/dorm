import { apiRequest } from '../../../shared/api/apiClient'
import type { LoginResponse } from '../model/types'

export async function loginResident(login: string, password: string): Promise<LoginResponse> {
    const response = await apiRequest('/api/v1/auth/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            login,
            password,
        }),
    })

    return response.json()
}
