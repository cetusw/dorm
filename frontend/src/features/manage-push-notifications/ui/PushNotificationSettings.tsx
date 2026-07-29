import { BellIcon, BellSlashIcon, WarningCircleIcon } from '@phosphor-icons/react'
import { Alert, Button, Group, Loader, Stack, Text } from '@mantine/core'

import { usePushNotifications } from '../model/usePushNotifications'

const notificationDescription = 'Включите уведомления, чтобы не забывать о дежурстве и задачах'

export function PushNotificationSettings() {
    const {
        state,
        error,
        enable,
        disable,
        refresh,
    } = usePushNotifications()

    function renderBody() {
        switch (state) {
            case 'loading':
                return (
                    <Group gap="sm" wrap="nowrap">
                        <Loader size="sm" />
                        <Text c="dimmed">Проверяем состояние уведомлений…</Text>
                    </Group>
                )
            case 'unsupported':
                return (
                    <Alert color="yellow" icon={<WarningCircleIcon size={18} />}>
                        Ваш браузер не поддерживает push-уведомления.
                    </Alert>
                )
            case 'insecure-context':
                return (
                    <Alert color="yellow" icon={<WarningCircleIcon size={18} />}>
                        Для уведомлений требуется HTTPS. Откройте приложение по защищённому адресу.
                    </Alert>
                )
            case 'disabled-by-server':
                return (
                    <Alert color="gray">
                        Уведомления временно недоступны.
                    </Alert>
                )
            case 'installation-required':
                return (
                    <Alert color="blue">
                        Чтобы получать уведомления на iPhone или iPad, добавьте приложение на экран «Домой» и откройте его с главного экрана.
                    </Alert>
                )
            case 'permission-denied':
                return (
                    <Stack gap="sm">
                        <Alert color="red" icon={<WarningCircleIcon size={18} />}>
                            Разрешите уведомления в настройках браузера или системы, затем вернитесь в приложение.
                        </Alert>
                        <Group>
                            <Button variant="default" onClick={() => void refresh()}>
                                Проверить снова
                            </Button>
                        </Group>
                    </Stack>
                )
            case 'subscribed':
                return (
                    <Stack gap="sm">
                        <Group>
                            <Button
                                color="red"
                                variant="light"
                                leftSection={<BellSlashIcon size={18} />}
                                onClick={() => void disable()}
                            >
                                Отключить уведомления
                            </Button>
                        </Group>
                    </Stack>
                )
            case 'subscribing':
                return (
                    <Group>
                        <Button
                            leftSection={<BellIcon size={18} />}
                            rightSection={<Loader size="xs" color="white" />}
                            disabled
                        >
                            Подключаем уведомления...
                        </Button>
                    </Group>
                )
            case 'unsubscribing':
                return (
                    <Group gap="sm" wrap="nowrap">
                        <Loader size="sm" />
                        <Text c="dimmed">Отключаем уведомления…</Text>
                    </Group>
                )
            case 'error':
                return (
                    <Stack gap="sm">
                        <Alert color="red" icon={<WarningCircleIcon size={18} />}>
                            {error ?? 'Не удалось обновить настройки уведомлений.'}
                        </Alert>
                    </Stack>
                )
            case 'unsubscribed':
            default:
                return (
                    <Group>
                        <Button
                            leftSection={<BellIcon size={18} />}
                            onClick={() => void enable()}
                        >
                            Включить уведомления
                        </Button>
                    </Group>
                )
        }
    }

    return (
        <Stack gap="md">
            <Text c="dimmed">
                {notificationDescription}
            </Text>
            {renderBody()}
        </Stack>
    )
}
