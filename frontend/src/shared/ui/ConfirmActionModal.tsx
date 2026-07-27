import { useEffect, useState } from 'react'

import { Alert, Button, Group, Modal, Stack, Text } from '@mantine/core'

import { ApiError } from '../api/ApiError'

type Props = {
    opened: boolean
    title: string
    description: string
    confirmLabel: string
    confirmColor?: string
    onClose: () => void
    onConfirm: () => Promise<void>
    errorMessage: string
}

export function ConfirmActionModal({
    opened,
    title,
    description,
    confirmLabel,
    confirmColor = 'dark',
    onClose,
    onConfirm,
    errorMessage,
}: Props) {
    const [submitting, setSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        if (!opened) {
            setSubmitting(false)
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
                {error ? <Alert color="red">{error}</Alert> : null}

                <Text size="md">{description}</Text>

                <Group justify="flex-end">
                    <Button variant="default" onClick={onClose}>
                        Отменить
                    </Button>
                    <Button
                        color={confirmColor}
                        loading={submitting}
                        onClick={async () => {
                            setSubmitting(true)
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
                                setSubmitting(false)
                            }
                        }}
                    >
                        {confirmLabel}
                    </Button>
                </Group>
            </Stack>
        </Modal>
    )
}
