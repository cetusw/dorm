import { apiRequest, beginAuthenticationChange } from '../../../shared/api/apiClient'
import type { LoginResponse } from '../model/types'

export async function loginResident(login: string, password: string): Promise<LoginResponse> {
    const finish = beginAuthenticationChange()
    try {
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

        return await response.json()
    } finally {
        finish()
    }
}

export async function logoutResident(): Promise<void> {
    const finish = beginAuthenticationChange()
    try {
        await apiRequest('/api/v1/auth/logout', {
            redirectOn401: false,
            method: 'POST',
        })
    } finally {
        finish()
    }
}
