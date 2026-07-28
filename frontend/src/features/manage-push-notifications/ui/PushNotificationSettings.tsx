import { BellIcon, BellSlashIcon, WarningCircleIcon } from '@phosphor-icons/react'
import { Alert, Button, Group, Loader, Stack, Text } from '@mantine/core'

import { SettingsBadge } from '../../../shared/ui/SettingsBadge'
import { usePushNotifications } from '../model/usePushNotifications'

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
                        <Group gap="sm" align="center">
                            <SettingsBadge color="success">Включены</SettingsBadge>
                            <Text c="dimmed">Это устройство будет получать напоминания о дежурстве и задачах.</Text>
                        </Group>
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
                    <Group gap="sm" wrap="nowrap">
                        <Loader size="sm" />
                        <Text c="dimmed">Подключаем уведомления…</Text>
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
                        <Group>
                            <Button variant="default" onClick={() => void refresh()}>
                                Повторить
                            </Button>
                        </Group>
                    </Stack>
                )
            case 'unsubscribed':
            default:
                return (
                    <Stack gap="sm">
                        <Text c="dimmed">
                            Получайте напоминания о дежурстве и задачах.
                        </Text>
                        <Group>
                            <Button
                                leftSection={<BellIcon size={18} />}
                                onClick={() => void enable()}
                            >
                                Включить уведомления
                            </Button>
                        </Group>
                    </Stack>
                )
        }
    }

    return (
        <Stack gap="md">
            <Text size="sm" c="dimmed">
                Браузерные push-уведомления для этого устройства.
            </Text>
            {renderBody()}
        </Stack>
    )
}
