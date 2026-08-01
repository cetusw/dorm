export { getWebPushConfig, savePushSubscription, deletePushSubscription } from './api/pushSubscriptionApi'
export {
    detectPushSupport,
    detectPushEnvironment,
    isIOSLikeDevice,
    base64UrlToUint8Array,
    getCurrentPushSubscription,
    requestNotificationPermission,
    createPushSubscription,
    mapPushSubscriptionToRequest,
    mapPushSubscriptionToDeleteRequest,
} from './model/pushSubscription.lib'
export type {
    WebPushConfig,
    SavePushSubscriptionRequest,
    DeletePushSubscriptionRequest,
    PushSupport,
    PushEnvironment,
} from './model/pushSubscription.types'
