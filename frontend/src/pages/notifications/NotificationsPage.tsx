import { PushNotificationSettings } from '../../features/manage-push-notifications'
import { PageFrame } from '../../shared/ui/PageFrame'

export function NotificationsPage() {
    return (
        <PageFrame title="Уведомления">
            <PushNotificationSettings />
        </PageFrame>
    )
}
