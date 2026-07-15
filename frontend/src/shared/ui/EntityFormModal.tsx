import type { FormEventHandler, ReactNode } from 'react'

import { Alert, Button, Center, Group, Loader, Modal, Stack } from '@mantine/core'

type Props = {
    opened: boolean
    onClose: () => void
    title: string
    loading?: boolean
    saving?: boolean
    error?: string | null
    size?: number | string
    onSubmit: FormEventHandler<HTMLFormElement>
    children: ReactNode
}

export function EntityFormModal({
    opened,
    onClose,
    title,
    loading = false,
    saving = false,
    error = null,
    size = 760,
    onSubmit,
    children,
}: Props) {
    return (
        <Modal
            opened={opened}
            onClose={onClose}
            title={title}
            centered
            radius="xl"
            size={size}
            styles={{
                title: {
                    fontSize: '1.5rem',
                    fontWeight: 700,
                },
                content: {
                    padding: 8,
                },
                body: {
                    paddingTop: 12,
                },
            }}
        >
            {loading ? (
                <Center py="xl">
                    <Loader />
                </Center>
            ) : (
                <form onSubmit={onSubmit}>
                    <Stack gap="md">
                        {error && (
                            <Alert color="red">{error}</Alert>
                        )}

                        {children}

                        <Group justify="flex-end" mt="sm">
                            <Button type="button" variant="default" onClick={onClose}>
                                Отменить
                            </Button>
                            <Button type="submit" loading={saving}>
                                Сохранить
                            </Button>
                        </Group>
                    </Stack>
                </form>
            )}
        </Modal>
    )
}
