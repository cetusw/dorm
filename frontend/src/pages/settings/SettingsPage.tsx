import { Button, Stack, Title } from '@mantine/core'

import { logoutResident } from '../../features/auth/api/authApi'
import { PushNotificationsSettingsSection } from '../../features/manage-push-notifications'
import { PageFrame } from '../../shared/ui/PageFrame'

async function handleResidentLogout() {
    try {
        await logoutResident()
    } finally {
        window.location.assign('/app/login')
    }
}

export function SettingsPage() {
    return (
        <PageFrame title="Настройки">
            <Stack gap="xl">
                <Stack gap="md">
                    <Title order={3}>Уведомления</Title>

                    <PushNotificationsSettingsSection />
                </Stack>

                <Stack gap="sm" align="flex-start">
                    <Button color="brand" onClick={() => void handleResidentLogout()}>
                        Выйти
                    </Button>
                </Stack>
            </Stack>
        </PageFrame>
    )
}
