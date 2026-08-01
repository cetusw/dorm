import { useEffect, useState } from 'react'

import { Alert, Button, Group, Modal, Stack, Text } from '@mantine/core'

import { ApiError } from '../api/ApiError'
import classes from './SettingsModal.module.css'

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
                {error && (
                    <Alert color="red">{error}</Alert>
                )}

                <Text className={classes.description}>
                    Вы уверены, что хотите удалить {entityLabel}
                    {entityName ? ` «${entityName}»` : ''}?
                </Text>

                <Group justify="flex-end" gap="15" className={classes.actions}>
                    <Button variant="default" onClick={onClose} className={classes.cancelButton}>
                        Отменить
                    </Button>
                    <Button
                        color="red"
                        loading={deleting}
                        className={classes.submitButton}
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
