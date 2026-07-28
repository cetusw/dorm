import type {
    DeletePushSubscriptionRequest,
    PushEnvironment,
    PushSupport,
    SavePushSubscriptionRequest,
} from './pushSubscription.types'

type NavigatorWithStandalone = Navigator & {
    standalone?: boolean
}

export function detectPushSupport(): PushSupport {
    const serviceWorkerSupported = 'serviceWorker' in navigator
    const notificationsSupported = 'Notification' in window
    const pushManagerSupported = 'PushManager' in window
    const secureContext = window.isSecureContext
    const standaloneByDisplayMode = window.matchMedia('(display-mode: standalone)').matches
    const navigatorWithStandalone = navigator as NavigatorWithStandalone
    const standaloneByIOS = navigatorWithStandalone.standalone === true
    const standalone = standaloneByDisplayMode || standaloneByIOS

    return {
        serviceWorkerSupported,
        notificationsSupported,
        pushManagerSupported,
        secureContext,
        standalone,
        supported:
            serviceWorkerSupported &&
            notificationsSupported &&
            pushManagerSupported &&
            secureContext,
    }
}

export function isIOSLikeDevice(): boolean {
    const classicIOS = /iPad|iPhone|iPod/.test(navigator.userAgent)
    const iPadDesktopMode = navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1

    return classicIOS || iPadDesktopMode
}

export function detectPushEnvironment(): PushEnvironment {
    const support = detectPushSupport()

    return {
        ...support,
        installRequired: isIOSLikeDevice() && !support.standalone,
    }
}

export function base64UrlToUint8Array(value: string): Uint8Array<ArrayBuffer> {
    const padding = '='.repeat((4 - (value.length % 4)) % 4)
    const normalized = (value + padding)
        .replace(/-/g, '+')
        .replace(/_/g, '/')

    const raw = window.atob(normalized)
    const result = new Uint8Array(raw.length)

    for (let index = 0; index < raw.length; index += 1) {
        result[index] = raw.charCodeAt(index)
    }

    return result
}

export async function getCurrentPushSubscription(
    registration: ServiceWorkerRegistration,
): Promise<PushSubscription | null> {
    return registration.pushManager.getSubscription()
}

export async function requestNotificationPermission(): Promise<NotificationPermission> {
    if (!('Notification' in window)) {
        throw new Error('Notifications API is not supported')
    }

    return Notification.requestPermission()
}

export async function createPushSubscription(
    registration: ServiceWorkerRegistration,
    publicKey: string,
): Promise<PushSubscription> {
    const existingSubscription = await registration.pushManager.getSubscription()
    if (existingSubscription) {
        return existingSubscription
    }

    return registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: base64UrlToUint8Array(publicKey),
    })
}

export function mapPushSubscriptionToRequest(
    subscription: PushSubscription,
): SavePushSubscriptionRequest {
    const json = subscription.toJSON()
    const endpoint = json.endpoint
    const p256dh = json.keys?.p256dh
    const auth = json.keys?.auth

    if (!endpoint || !p256dh || !auth) {
        throw new Error('Browser returned an incomplete push subscription')
    }

    return {
        endpoint,
        keys: {
            p256dh,
            auth,
        },
    }
}

export function mapPushSubscriptionToDeleteRequest(
    endpoint: string,
): DeletePushSubscriptionRequest {
    return {
        endpoint,
    }
}
