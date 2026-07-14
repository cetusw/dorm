import { useEffect, useState } from 'react'

import { Alert, Button, Group, Modal, Stack, Text } from '@mantine/core'

import { ApiError } from '../../../shared/api/ApiError'
import { deleteDormitory } from '../api/dormitoriesApi'
import type { DormitoryListItem } from '../model/types'

type Props = {
    opened: boolean
    dormitory: DormitoryListItem | null
    onClose: () => void
    onDeleted: () => Promise<void> | void
}

export function DeleteDormitoryModal({ opened, dormitory, onClose, onDeleted }: Props) {
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
            title="Удаление общежития"
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
                    Вы уверены, что хотите удалить общежитие
                    {dormitory ? ` «${dormitory.name}»` : ''}?
                </Text>

                <Group justify="flex-end">
                    <Button variant="default" onClick={onClose}>
                        Отменить
                    </Button>
                    <Button
                        color="red"
                        loading={deleting}
                        onClick={async () => {
                            if (!dormitory) {
                                return
                            }

                            setDeleting(true)
                            setError(null)

                            try {
                                await deleteDormitory(dormitory.id)
                                await onDeleted()
                                onClose()
                            } catch (currentError) {
                                if (currentError instanceof ApiError) {
                                    setError(currentError.message)
                                } else {
                                    setError('Не удалось удалить общежитие')
                                }
                            } finally {
                                setDeleting(false)
                            }
                        }}
                    >
                        Подтвердить
                    </Button>
                </Group>
            </Stack>
        </Modal>
    )
}
