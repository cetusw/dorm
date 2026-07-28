export type PushNotificationState =
    | 'loading'
    | 'unsupported'
    | 'insecure-context'
    | 'disabled-by-server'
    | 'installation-required'
    | 'permission-denied'
    | 'unsubscribed'
    | 'subscribed'
    | 'subscribing'
    | 'unsubscribing'
    | 'error'

export type UsePushNotificationsResult = {
    state: PushNotificationState
    error: string | null
    enable: () => Promise<void>
    disable: () => Promise<void>
    refresh: () => Promise<void>
}
