import { useCallback, useEffect, useState } from 'react'

import {
    createPushSubscription,
    savePushSubscription,
    deletePushSubscription,
    detectPushEnvironment,
    getCurrentPushSubscription,
    getWebPushConfig,
    mapPushSubscriptionToDeleteRequest,
    mapPushSubscriptionToRequest,
    requestNotificationPermission,
} from '../../../entities/push-subscription'
import type { PushNotificationState, UsePushNotificationsResult } from './pushNotifications.types'

function toErrorMessage(error: unknown): string {
    if (error instanceof Error && error.message.trim() !== '') {
        return error.message
    }

    return 'Не удалось обновить настройки уведомлений.'
}

async function getReadyRegistration(): Promise<ServiceWorkerRegistration> {
    if (!('serviceWorker' in navigator)) {
        throw new Error('Service Worker не поддерживается')
    }

    return navigator.serviceWorker.ready
}

export function usePushNotifications(): UsePushNotificationsResult {
    const [state, setState] = useState<PushNotificationState>('loading')
    const [error, setError] = useState<string | null>(null)

    const refresh = useCallback(async () => {
        setState('loading')
        setError(null)

        try {
            const environment = detectPushEnvironment()

            if (!environment.secureContext) {
                setState('insecure-context')
                return
            }

            if (!environment.serviceWorkerSupported || !environment.notificationsSupported || !environment.pushManagerSupported) {
                setState('unsupported')
                return
            }

            const config = await getWebPushConfig()
            if (!config.enabled || !config.publicKey) {
                setState('disabled-by-server')
                return
            }

            if (environment.installRequired) {
                setState('installation-required')
                return
            }

            if (Notification.permission === 'denied') {
                setState('permission-denied')
                return
            }

            const registration = await getReadyRegistration()
            const subscription = await getCurrentPushSubscription(registration)

            if (subscription) {
                setState('subscribed')
                return
            }

            setState('unsubscribed')
        } catch (currentError) {
            setError(toErrorMessage(currentError))
            setState('error')
        }
    }, [])

    useEffect(() => {
        void refresh()
    }, [refresh])

    const enable = useCallback(async () => {
        setState('subscribing')
        setError(null)

        try {
            const environment = detectPushEnvironment()

            if (!environment.secureContext) {
                setState('insecure-context')
                return
            }

            if (!environment.serviceWorkerSupported || !environment.notificationsSupported || !environment.pushManagerSupported) {
                setState('unsupported')
                return
            }

            const config = await getWebPushConfig()
            if (!config.enabled || !config.publicKey) {
                setState('disabled-by-server')
                return
            }

            if (environment.installRequired) {
                setState('installation-required')
                return
            }

            const permission = await requestNotificationPermission()
            if (permission !== 'granted') {
                setState(permission === 'denied' ? 'permission-denied' : 'unsubscribed')
                return
            }

            const registration = await getReadyRegistration()
            const subscription = await createPushSubscription(registration, config.publicKey)
            await savePushSubscription(mapPushSubscriptionToRequest(subscription))
            setState('subscribed')
        } catch (currentError) {
            setError(toErrorMessage(currentError))
            setState('error')
        }
    }, [])

    const disable = useCallback(async () => {
        setState('unsubscribing')
        setError(null)

        try {
            const registration = await getReadyRegistration()
            const subscription = await getCurrentPushSubscription(registration)

            if (!subscription) {
                setState('unsubscribed')
                return
            }

            await deletePushSubscription(mapPushSubscriptionToDeleteRequest(subscription.endpoint))
            const unsubscribed = await subscription.unsubscribe()
            if (!unsubscribed) {
                throw new Error('Браузер не удалил push-подписку')
            }

            setState('unsubscribed')
        } catch (currentError) {
            setError(toErrorMessage(currentError))
            setState('error')
        }
    }, [])

    return {
        state,
        error,
        enable,
        disable,
        refresh,
    }
}
