export type WebPushConfig = {
    enabled: boolean
    publicKey: string | null
}

export type SavePushSubscriptionRequest = {
    endpoint: string
    keys: {
        p256dh: string
        auth: string
    }
}

export type DeletePushSubscriptionRequest = {
    endpoint: string
}

export type PushSupport = {
    serviceWorkerSupported: boolean
    notificationsSupported: boolean
    pushManagerSupported: boolean
    secureContext: boolean
    standalone: boolean
    supported: boolean
}

export type PushEnvironment = PushSupport & {
    installRequired: boolean
}
