import { useEffect, useState } from 'react'

import { TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import modalClasses from '../../../shared/ui/SettingsModal.module.css'
import { updateWarehouseItem } from '../api/warehouseApi'
import type { UpdateWarehouseItemRequest, WarehouseItem } from '../model/types'

type Props = {
    opened: boolean
    item: WarehouseItem | null
    onClose: () => void
    onSaved: (item: WarehouseItem) => Promise<void> | void
}

type WarehouseItemEditFormValues = {
    name: string
}

const initialValues: WarehouseItemEditFormValues = {
    name: '',
}

function toRequest(values: WarehouseItemEditFormValues): UpdateWarehouseItemRequest {
    return {
        name: values.name.trim(),
    }
}

const validation = {
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
}

export function WarehouseItemEditModal({ opened, item, onClose, onSaved }: Props) {
    const [saving, setSaving] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)

    const form = useForm<WarehouseItemEditFormValues>({
        mode: 'controlled',
        initialValues,
        validate: validation,
    })

    useEffect(() => {
        if (!opened) {
            setSaving(false)
            setSubmitError(null)
            return
        }

        if (!item) {
            return
        }

        const values = { name: item.name }
        form.setValues(values)
        form.resetDirty(values)
        form.clearErrors()
        setSaving(false)
        setSubmitError(null)
    }, [item, opened])

    return (
        <EntityFormModal
            opened={opened}
            onClose={onClose}
            title={<span className={modalClasses.title}>Редактирование элемента учёта</span>}
            saving={saving}
            error={submitError}
            size={680}
            withCloseButton={false}
            withInitialFocus={false}
            modalClassNames={{
                header: modalClasses.header,
                body: modalClasses.body,
                content: modalClasses.content,
            }}
            contentGap={15}
            actionsClassName={modalClasses.actions}
            cancelButtonClassName={modalClasses.cancelButton}
            submitButtonClassName={[modalClasses.submitButton, modalClasses.accentButton].join(' ')}
            submitLabel="Редактировать"
            onSubmit={form.onSubmit(async (values) => {
                if (!item) {
                    return
                }

                const nextName = values.name.trim()
                if (nextName === item.name.trim()) {
                    onClose()
                    return
                }

                setSubmitError(null)
                setSaving(true)

                try {
                    const updatedItem = await updateWarehouseItem(item.id, toRequest(values))
                    await onSaved(updatedItem)
                    onClose()
                } catch (error) {
                    if (error instanceof ApiError) {
                        setSubmitError(error.message)
                    } else {
                        setSubmitError('Не удалось отредактировать элемент учёта')
                    }
                } finally {
                    setSaving(false)
                }
            })}
        >
            <TextInput
                data-autofocus
                autoFocus
                placeholder="Название*"
                maxLength={255}
                classNames={{
                    input: modalClasses.input,
                }}
                key={form.key('name')}
                {...form.getInputProps('name')}
            />
        </EntityFormModal>
    )
}
