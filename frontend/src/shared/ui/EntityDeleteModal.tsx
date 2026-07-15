import { useEffect, useState } from 'react'

import { Alert, Button, Group, Modal, Stack, Text } from '@mantine/core'

import { ApiError } from '../api/ApiError'

type Props = {
    opened: boolean
    title: string
    entityLabel: string
    entityName: string | null
    submitLabel?: string
    onClose: () => void
    onConfirm: () => Promise<void>
    errorMessage: string
}

export function EntityDeleteModal({
    opened,
    title,
    entityLabel,
    entityName,
    submitLabel = 'Подтвердить',
    onClose,
    onConfirm,
    errorMessage,
}: Props) {
    const [deleting, setDeleting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        if (!opened) {
            setDeleting(false)
            setError(null)
        }
    }, [opened])

    return (
        <Modal
            opened={opened}
            onClose={onClose}
            title={title}
            centered
            radius="xl"
            size={520}
            styles={{
                title: {
                    fontSize: '1.4rem',
                    fontWeight: 700,
                },
            }}
        >
            <Stack gap="lg">
                {error && (
                    <Alert color="red">{error}</Alert>
                )}

                <Text size="md">
                    Вы уверены, что хотите удалить {entityLabel}
                    {entityName ? ` «${entityName}»` : ''}?
                </Text>

                <Group justify="flex-end">
                    <Button variant="default" onClick={onClose}>
                        Отменить
                    </Button>
                    <Button
                        color="red"
                        loading={deleting}
                        onClick={async () => {
                            setDeleting(true)
                            setError(null)

                            try {
                                await onConfirm()
                                onClose()
                            } catch (currentError) {
                                if (currentError instanceof ApiError) {
                                    setError(currentError.message)
                                } else {
                                    setError(errorMessage)
                                }
                            } finally {
                                setDeleting(false)
                            }
                        }}
                    >
                        {submitLabel}
                    </Button>
                </Group>
            </Stack>
        </Modal>
    )
}
