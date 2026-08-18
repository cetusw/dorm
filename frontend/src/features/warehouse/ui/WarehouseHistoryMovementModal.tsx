import { useEffect, useState } from 'react'

import { NumberInput, Textarea } from '@mantine/core'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import modalClasses from '../../../shared/ui/SettingsModal.module.css'
import { updateWarehouseMovement } from '../api/warehouseApi'
import {
    initialWarehouseMovementFormValues,
    toWarehouseMovementUpdateRequest,
    warehouseMovementValidation,
    type WarehouseMovementFormValues,
} from '../model/movementForm'
import type { WarehouseItem, WarehouseMovement } from '../model/types'

type Props = {
    opened: boolean
    item: WarehouseItem | null
    movement: WarehouseMovement | null
    onClose: () => void
    onSaved: (item: WarehouseItem) => Promise<void> | void
}

export function WarehouseHistoryMovementModal({ opened, item, movement, onClose, onSaved }: Props) {
    const [saving, setSaving] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)

    const form = useForm<WarehouseMovementFormValues>({
        mode: 'controlled',
        initialValues: initialWarehouseMovementFormValues,
        validate: warehouseMovementValidation,
    })

    useEffect(() => {
        if (!opened || !movement) {
            form.setValues(initialWarehouseMovementFormValues)
            form.resetDirty(initialWarehouseMovementFormValues)
            form.clearErrors()
            setSaving(false)
            setSubmitError(null)
            return
        }

        const values = {
            quantity: String(movement.quantity),
            comment: movement.comment ?? '',
        }
        form.setValues(values)
        form.resetDirty(values)
        form.clearErrors()
        setSaving(false)
        setSubmitError(null)
    }, [movement, opened])

    const title = movement?.type === 'write-off'
        ? 'Редактирование списания'
        : 'Редактирование поступления'

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
            submitLabel="Редактировать"
            onSubmit={form.onSubmit(async (values) => {
                if (!item || !movement) {
                    return
                }

                setSubmitError(null)
                setSaving(true)

                try {
                    const updatedItem = await updateWarehouseMovement(item.id, movement.id, toWarehouseMovementUpdateRequest(values))
                    await onSaved(updatedItem)
                    onClose()
                } catch (error) {
                    if (error instanceof ApiError) {
                        setSubmitError(error.message)
                    } else {
                        setSubmitError('Не удалось отредактировать движение склада')
                    }
                } finally {
                    setSaving(false)
                }
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
