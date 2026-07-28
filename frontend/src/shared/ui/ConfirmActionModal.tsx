import { useEffect, useState } from 'react'

import { Alert, Button, Group, Modal, Stack, Text } from '@mantine/core'

import { ApiError } from '../api/ApiError'
import classes from './SettingsModal.module.css'

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
            title={<span className={classes.title}>{title}</span>}
            withCloseButton={false}
            centered
            radius="xl"
            size={520}
            classNames={{
                header: classes.header,
                body: classes.body,
                content: classes.content,
            }}
        >
            <Stack gap="lg">
                {error ? <Alert color="red">{error}</Alert> : null}

                <Text className={classes.description}>{description}</Text>

                <Group justify="flex-end" gap="15" className={classes.actions}>
                    <Button variant="default" onClick={onClose} className={classes.cancelButton}>
                        Отменить
                    </Button>
                    <Button
                        color={confirmColor}
                        loading={submitting}
                        className={classes.submitButton}
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
