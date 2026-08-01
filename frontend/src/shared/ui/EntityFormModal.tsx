import type { FormEventHandler, ReactNode } from 'react'

import { Alert, Button, Center, Group, Loader, Modal, Stack } from '@mantine/core'
import type { ModalProps } from '@mantine/core'

type Props = {
    opened: boolean
    onClose: () => void
    title: ReactNode
    loading?: boolean
    saving?: boolean
    error?: string | null
    size?: number | string
    withCloseButton?: boolean
    modalClassNames?: ModalProps['classNames']
    contentGap?: string | number
    actionsClassName?: string
    cancelButtonClassName?: string
    submitButtonClassName?: string
    cancelLabel?: string
    submitLabel?: string
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
    withCloseButton = true,
    modalClassNames,
    contentGap = 'md',
    actionsClassName,
    cancelButtonClassName,
    submitButtonClassName,
    cancelLabel = 'Отменить',
    submitLabel = 'Сохранить',
    onSubmit,
    children,
}: Props) {
    return (
        <Modal
            opened={opened}
            onClose={onClose}
            title={title}
            withCloseButton={withCloseButton}
            centered
            radius="xl"
            size={size}
            classNames={modalClassNames}
            styles={modalClassNames ? undefined : {
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
                    <Stack gap={contentGap}>
                        {error && (
                            <Alert color="red">{error}</Alert>
                        )}

                        {children}

                        <Group justify="flex-end" gap="15" mt="sm" className={actionsClassName}>
                            <Button type="button" variant="default" onClick={onClose} className={cancelButtonClassName}>
                                {cancelLabel}
                            </Button>
                            <Button type="submit" loading={saving} className={submitButtonClassName}>
                                {submitLabel}
                            </Button>
                        </Group>
                    </Stack>
                </form>
            )}
        </Modal>
    )
}
