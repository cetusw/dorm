import { apiRequest } from '../../../shared/api/apiClient'
import type { LoginResponse } from '../model/types'

export async function loginResident(login: string, password: string): Promise<LoginResponse> {
    const response = await apiRequest('/api/v1/auth/login', {
        method: 'POST',
        redirectOn401: false,
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

export async function logoutResident(): Promise<void> {
    await apiRequest('/api/v1/auth/logout', {
        method: 'POST',
    })
}
