import { useEffect, useState } from 'react'

import { NumberInput, Textarea } from '@mantine/core'
import { useDebouncedValue } from '@mantine/hooks'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import {
    ResidentSearchCombobox,
    type ResidentSearchOption,
} from '../../../shared/ui/ResidentSearchCombobox'
import modalClasses from '../../../shared/ui/SettingsModal.module.css'
import {
    createPenalty,
    resolvePenalty,
    searchPenaltyResidents,
    updatePenaltyEntry,
} from '../api/penaltiesApi'
import type {
    CreatePenaltyRequest,
    PenaltyEntryType,
    PenaltyResidentOption,
    ResolvePenaltyRequest,
    UpdatePenaltyEntryRequest,
} from '../model/types'
import { countPenaltyReasonCharacters, formatPenaltyWeight } from '../model/utils'

type Props = {
    opened: boolean
    onClose: () => void
    onCreated: () => Promise<void> | void
    residentPreset?: {
        id: string
        name: string
    } | null
    lockResident?: boolean
    mode?: 'create' | 'edit'
    entryType?: PenaltyEntryType
    entryId?: string | null
    maxWeight?: number
    initialReason?: string
    initialWeight?: number | null
}

type FormValues = {
    residentId: string
    residentName: string
    reason: string
    weight: number | string
}

const emptyValues: FormValues = {
    residentId: '',
    residentName: '',
    reason: '',
    weight: '',
}

function toSearchOptions(options: PenaltyResidentOption[]): ResidentSearchOption[] {
    return options.map((item) => ({
        id: item.id,
        name: item.name,
    }))
}

function toIssueRequest(values: FormValues): CreatePenaltyRequest {
    return {
        user_id: values.residentId,
        reason: values.reason.trim(),
        weight: Number(values.weight),
    }
}

function toResolveRequest(values: FormValues): ResolvePenaltyRequest {
    return {
        user_id: values.residentId,
        reason: values.reason.trim(),
        weight: Number(values.weight),
    }
}

function toUpdateRequest(values: FormValues): UpdatePenaltyEntryRequest {
    return {
        reason: values.reason.trim(),
        weight: Number(values.weight),
    }
}

function resolveTitle(mode: 'create' | 'edit', entryType: PenaltyEntryType): string {
    if (mode === 'create') {
        return entryType === 'resolve' ? 'Погашение предупреждения' : 'Выдача предупреждения'
    }

    return entryType === 'resolve'
        ? 'Редактирование погашения'
        : 'Редактирование предупреждения'
}

export function CreatePenaltyDrawer({
    opened,
    onClose,
    onCreated,
    residentPreset = null,
    lockResident = false,
    mode = 'create',
    entryType = 'issue',
    entryId = null,
    maxWeight,
    initialReason = '',
    initialWeight = null,
}: Props) {
    const isResolveMode = entryType === 'resolve'
    const isEditMode = mode === 'edit'
    const residentPresetId = residentPreset?.id ?? ''
    const residentPresetName = residentPreset?.name ?? ''
    const [saving, setSaving] = useState(false)
    const [searchValue, setSearchValue] = useState('')
    const [residentOptions, setResidentOptions] = useState<PenaltyResidentOption[]>([])
    const [searchLoading, setSearchLoading] = useState(false)
    const [searchError, setSearchError] = useState<string | null>(null)
    const [debouncedSearch] = useDebouncedValue(searchValue, 300)
    const [submitError, setSubmitError] = useState<string | null>(null)

    const form = useForm<FormValues>({
        mode: 'controlled',
        initialValues: emptyValues,
        validate: {
            residentId: (value) => value.trim() === '' ? 'Выберите жителя' : null,
            reason: (value) => {
                const trimmed = value.trim()
                if (trimmed === '') {
                    return 'Укажите причину'
                }
                if (countPenaltyReasonCharacters(trimmed) > 256) {
                    return 'Причина не должна превышать 256 символов'
                }
                return null
            },
            weight: (value) => {
                if (value === '' || value == null) {
                    return 'Укажите вес'
                }
                const normalized = Number(value)
                if (!Number.isFinite(normalized)) {
                    return 'Укажите вес'
                }
                if (normalized < 0.1) {
                    return 'Минимальный вес — 0,1'
                }
                if (normalized > 999999999.9) {
                    return 'Укажите вес'
                }
                if (!isEditMode && isResolveMode && typeof maxWeight === 'number' && normalized > maxWeight) {
                    return `Нельзя снять больше ${formatPenaltyWeight(maxWeight)}`
                }
                return null
            },
        },
    })

    useEffect(() => {
        if (!opened) {
            form.setValues(emptyValues)
            form.resetDirty(emptyValues)
            form.clearErrors()
            setSaving(false)
            setSubmitError(null)
            setSearchLoading(false)
            setSearchError(null)
            setResidentOptions([])
            setSearchValue('')
            return
        }

        const nextValues: FormValues = {
            residentId: residentPresetId,
            residentName: residentPresetName,
            reason: initialReason,
            weight: initialWeight ?? '',
        }
        form.setValues(nextValues)
        form.resetDirty(nextValues)
        form.clearErrors()
        setSaving(false)
        setSubmitError(null)
        setSearchError(null)
        setSearchValue(nextValues.residentName)
    }, [initialReason, initialWeight, opened, residentPresetId, residentPresetName])

    useEffect(() => {
        if (!opened) {
            return
        }

        if (lockResident && residentPresetId !== '') {
            return
        }

        let cancelled = false
        const timeoutId = window.setTimeout(() => {
            setSearchLoading(true)
            setSearchError(null)

            void searchPenaltyResidents(debouncedSearch)
                .then((response) => {
                    if (!cancelled) {
                        setResidentOptions(response.residents)
                    }
                })
                .catch((currentError) => {
                    if (cancelled) {
                        return
                    }

                    setResidentOptions([])
                    setSearchError(
                        currentError instanceof ApiError
                            ? currentError.message
                            : 'Не удалось выполнить запрос',
                    )
                })
                .finally(() => {
                    if (!cancelled) {
                        setSearchLoading(false)
                    }
                })
        }, 0)

        return () => {
            cancelled = true
            window.clearTimeout(timeoutId)
        }
    }, [debouncedSearch, lockResident, opened, residentPresetId])

    const residentError = typeof form.errors.residentId === 'string'
        ? form.errors.residentId
        : searchError

    return (
        <EntityFormModal
            opened={opened}
            onClose={onClose}
            title={<span className={modalClasses.title}>{resolveTitle(mode, entryType)}</span>}
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
            submitLabel={isEditMode ? 'Сохранить' : isResolveMode ? 'Погасить' : 'Выдать'}
            onSubmit={form.onSubmit(async (values) => {
                setSaving(true)
                setSubmitError(null)

                try {
                    if (isEditMode) {
                        if (!entryId) {
                            return
                        }
                        await updatePenaltyEntry(entryId, toUpdateRequest(values))
                    } else if (isResolveMode) {
                        await resolvePenalty(toResolveRequest(values))
                    } else {
                        await createPenalty(toIssueRequest(values))
                    }

                    await onCreated()
                    onClose()
                } catch (currentError) {
                    if (currentError instanceof ApiError) {
                        setSubmitError(currentError.message)
                    } else {
                        setSubmitError('Не удалось выполнить запрос')
                    }
                } finally {
                    setSaving(false)
                }
            })}
        >
            <ResidentSearchCombobox
                overlayOpened={opened}
                placeholder="Житель*"
                searchValue={form.values.residentName}
                selectedId={form.values.residentId === '' ? null : form.values.residentId}
                selectedLabel={form.values.residentName === '' ? null : form.values.residentName}
                hideDropdownWhenSelected
                options={toSearchOptions(residentOptions)}
                loading={searchLoading}
                error={residentError}
                disabled={lockResident || isEditMode}
                inputClassName={modalClasses.input}
                onSearchChange={(value: string) => {
                    setSearchValue(value)
                    form.setFieldValue('residentName', value)
                    if (value.trim() === '') {
                        form.setFieldValue('residentId', '')
                    }
                    form.clearFieldError('residentId')
                }}
                onOptionSelect={(option: ResidentSearchOption) => {
                    form.setFieldValue('residentId', option.id)
                    form.setFieldValue('residentName', option.name)
                    setSearchValue(option.name)
                    form.clearFieldError('residentId')
                }}
            />

            <NumberInput
                placeholder="Вес*"
                hideControls
                decimalScale={1}
                fixedDecimalScale={false}
                allowNegative={false}
                min={0.1}
                max={!isEditMode && isResolveMode && typeof maxWeight === 'number' ? maxWeight : 999999999.9}
                clampBehavior="strict"
                classNames={{
                    input: modalClasses.input,
                }}
                value={form.values.weight}
                error={form.errors.weight}
                onChange={(value) => form.setFieldValue('weight', value === '' ? '' : value)}
            />

            <Textarea
                placeholder="Причина*"
                minRows={4}
                autosize
                maxRows={8}
                classNames={{
                    input: modalClasses.input,
                }}
                key={form.key('reason')}
                {...form.getInputProps('reason')}
            />
        </EntityFormModal>
    )
}
