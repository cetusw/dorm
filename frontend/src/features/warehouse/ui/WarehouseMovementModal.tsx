import { useEffect, useMemo, useState } from 'react'

import { NumberInput, Textarea } from '@mantine/core'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import modalClasses from '../../../shared/ui/SettingsModal.module.css'
import { addWarehouseItemQuantity, writeOffWarehouseItemQuantity } from '../api/warehouseApi'
import {
    initialWarehouseMovementFormValues,
    toWarehouseMovementRequest,
    warehouseMovementValidation,
    type WarehouseMovementFormValues,
} from '../model/movementForm'
import type { WarehouseItem } from '../model/types'

type Props = {
    mode: 'add' | 'write-off'
    opened: boolean
    item: WarehouseItem | null
    onClose: () => void
    onSaved: (item: WarehouseItem) => Promise<void> | void
}

export function WarehouseMovementModal({ mode, opened, item, onClose, onSaved }: Props) {
    const [saving, setSaving] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)

    const form = useForm<WarehouseMovementFormValues>({
        mode: 'controlled',
        initialValues: initialWarehouseMovementFormValues,
        validate: warehouseMovementValidation,
    })

    useEffect(() => {
        if (!opened) {
            form.setValues(initialWarehouseMovementFormValues)
            form.resetDirty(initialWarehouseMovementFormValues)
            form.clearErrors()
            setSaving(false)
            setSubmitError(null)
        }
    }, [opened])

    const title = useMemo(() => {
        const itemName = item?.name ?? ''

        return mode === 'add'
            ? `Добавление ${itemName}`
            : `Списание ${itemName}`
    }, [item, mode])

    const submitLabel = mode === 'add' ? 'Добавить' : 'Списать'

    return (
        <EntityFormModal
            opened={opened}
            onClose={onClose}
            title={<span className={modalClasses.title}>{title}</span>}
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
            submitLabel={submitLabel}
            onSubmit={form.onSubmit(async (values) => {
                if (!item) {
                    return
                }

                setSubmitError(null)
                setSaving(true)

                try {
                    let updatedItem: WarehouseItem

                    if (mode === 'add') {
                        updatedItem = await addWarehouseItemQuantity(item.id, toWarehouseMovementRequest(values))
                    } else {
                        updatedItem = await writeOffWarehouseItemQuantity(item.id, toWarehouseMovementRequest(values))
                    }

                    await onSaved(updatedItem)
                } catch (error) {
                    if (error instanceof ApiError) {
                        setSubmitError(error.message)
                    } else if (mode === 'add') {
                        setSubmitError('Не удалось добавить единицы на склад')
                    } else {
                        setSubmitError('Не удалось списать единицы со склада')
                    }

                    return
                } finally {
                    setSaving(false)
                }

                onClose()
            })}
        >
            <NumberInput
                placeholder="Количество*"
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
