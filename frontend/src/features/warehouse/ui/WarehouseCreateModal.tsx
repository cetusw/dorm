import { useEffect, useState } from 'react'

import { NumberInput, TextInput, Textarea } from '@mantine/core'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import modalClasses from '../../../shared/ui/SettingsModal.module.css'
import { createWarehouseItem } from '../api/warehouseApi'
import type { CreateWarehouseItemRequest, WarehouseItem } from '../model/types'

type Props = {
    opened: boolean
    onClose: () => void
    onSaved: (item: WarehouseItem) => Promise<void> | void
}

type WarehouseCreateFormValues = {
    name: string
    quantity: string
    comment: string
}

const initialValues: WarehouseCreateFormValues = {
    name: '',
    quantity: '',
    comment: '',
}

function toNullableString(value: string): string | null {
    const trimmed = value.trim()
    return trimmed.length === 0 ? null : trimmed
}

function toRequest(values: WarehouseCreateFormValues): CreateWarehouseItemRequest {
    const quantity = values.quantity.trim()

    return {
        name: values.name.trim(),
        quantity: quantity === '' ? undefined : Number(quantity),
        comment: toNullableString(values.comment),
    }
}

const warehouseCreateValidation = {
    name: (value: string) => {
        const trimmed = value.trim()

        if (trimmed.length === 0) {
            return 'Введите название'
        }

        if ([...trimmed].length > 255) {
            return 'Название не должно превышать 255 символов'
        }

        return null
    },
    quantity: (value: string) => {
        const trimmed = value.trim()

        if (trimmed.length === 0) {
            return null
        }

        if (!/^\d+$/.test(trimmed)) {
            return 'Количество должно быть целым числом'
        }

        if (Number(trimmed) <= 0) {
            return 'Количество должно быть больше нуля'
        }

        return null
    },
    comment: (value: string) => {
        if ([...value.trim()].length > 256) {
            return 'Комментарий не должен превышать 256 символов'
        }

        return null
    },
}

export function WarehouseCreateModal({ opened, onClose, onSaved }: Props) {
    const [saving, setSaving] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)

    const form = useForm<WarehouseCreateFormValues>({
        mode: 'controlled',
        initialValues,
        validate: warehouseCreateValidation,
    })

    useEffect(() => {
        if (!opened) {
            form.setValues(initialValues)
            form.resetDirty(initialValues)
            form.clearErrors()
            setSaving(false)
            setSubmitError(null)
        }
    }, [opened])

    return (
        <EntityFormModal
            opened={opened}
            onClose={onClose}
            title={<span className={modalClasses.title}>Добавление на склад</span>}
            saving={saving}
            error={submitError}
            size={680}
            withCloseButton={false}
            modalClassNames={{
                header: modalClasses.header,
                body: modalClasses.body,
                content: modalClasses.content,
            }}
            contentGap={15}
            actionsClassName={modalClasses.actions}
            cancelButtonClassName={modalClasses.cancelButton}
            submitButtonClassName={[modalClasses.submitButton, modalClasses.accentButton].join(' ')}
            submitLabel="Добавить"
            onSubmit={form.onSubmit(async (values) => {
                setSubmitError(null)
                setSaving(true)

                try {
                    const item = await createWarehouseItem(toRequest(values))
                    await onSaved(item)
                    onClose()
                } catch (error) {
                    if (error instanceof ApiError) {
                        setSubmitError(error.message)
                    } else {
                        setSubmitError('Не удалось добавить предмет на склад')
                    }
                } finally {
                    setSaving(false)
                }
            })}
        >
            <TextInput
                data-autofocus
                placeholder="Название*"
                maxLength={255}
                classNames={{
                    input: modalClasses.input,
                }}
                key={form.key('name')}
                {...form.getInputProps('name')}
            />

            <NumberInput
                placeholder="Количество"
                allowDecimal={false}
                allowNegative={false}
                hideControls
                clampBehavior="strict"
                classNames={{
                    input: modalClasses.input,
                }}
                value={form.values.quantity}
                error={form.errors.quantity}
                onChange={(value) => form.setFieldValue('quantity', value === '' ? '' : String(value))}
            />

            <Textarea
                placeholder="Комментарий"
                maxLength={256}
                minRows={3}
                autosize
                classNames={{
                    input: modalClasses.input,
                }}
                key={form.key('comment')}
                {...form.getInputProps('comment')}
            />
        </EntityFormModal>
    )
}
