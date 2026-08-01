import { apiRequest } from '../../../shared/api/apiClient'
import type {
    DeletePushSubscriptionRequest,
    SavePushSubscriptionRequest,
    WebPushConfig,
} from '../model/pushSubscription.types'

function normalizeWebPushConfig(payload: Record<string, unknown>): WebPushConfig {
    return {
        enabled: Boolean(payload.enabled),
        publicKey: typeof payload.publicKey === 'string' && payload.publicKey.trim() !== ''
            ? payload.publicKey
            : null,
    }
}

export async function getWebPushConfig(): Promise<WebPushConfig> {
    const response = await apiRequest('/api/v1/notifications/config')
    return normalizeWebPushConfig(await response.json())
}

export async function savePushSubscription(request: SavePushSubscriptionRequest): Promise<void> {
    await apiRequest('/api/v1/notifications/subscriptions', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })
}

export async function deletePushSubscription(request: DeletePushSubscriptionRequest): Promise<void> {
    await apiRequest('/api/v1/notifications/subscriptions', {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })
}
