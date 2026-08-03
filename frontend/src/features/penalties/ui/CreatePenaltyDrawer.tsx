import { useEffect, useRef, useState } from 'react'

import dayjs from 'dayjs'
import 'dayjs/locale/ru'
import { DatePickerInput } from '@mantine/dates'
import { Button, Drawer, FocusTrap, Group, NumberInput, Stack, Textarea } from '@mantine/core'
import { useDebouncedValue } from '@mantine/hooks'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import {
    ResidentSearchCombobox,
    type ResidentSearchOption,
} from '../../../shared/ui/ResidentSearchCombobox'
import { createPenalty, searchPenaltyResidents } from '../api/penaltiesApi'
import type { CreatePenaltyRequest, PenaltyResidentOption } from '../model/types'
import { countPenaltyReasonCharacters, getTodayPenaltyDate } from '../model/utils'
import classes from './CreatePenaltyDrawer.module.css'

type Props = {
    opened: boolean
    onClose: () => void
    onCreated: () => Promise<void> | void
    residentPreset?: {
        id: string
        name: string
    } | null
    lockResident?: boolean
}

type FormValues = {
    residentId: string
    residentName: string
    reason: string
    weight: number | string
    issuedOn: string | null
}

function toSearchOptions(options: PenaltyResidentOption[]): ResidentSearchOption[] {
    return options.map((item) => ({
        id: item.id,
        name: item.name,
    }))
}

function toRequest(values: FormValues): CreatePenaltyRequest {
    return {
        user_id: values.residentId,
        reason: values.reason.trim(),
        weight: Number(values.weight),
        issued_on: values.issuedOn ?? '',
    }
}

export function CreatePenaltyDrawer({
    opened,
    onClose,
    onCreated,
    residentPreset = null,
    lockResident = false,
}: Props) {
    const [submitting, setSubmitting] = useState(false)
    const [searchValue, setSearchValue] = useState('')
    const [residentOptions, setResidentOptions] = useState<PenaltyResidentOption[]>([])
    const [searchLoading, setSearchLoading] = useState(false)
    const [searchError, setSearchError] = useState<string | null>(null)
    const [debouncedSearch] = useDebouncedValue(searchValue, 300)
    const [submitError, setSubmitError] = useState<string | null>(null)
    const wasOpenedRef = useRef(false)
    const appliedPresetKeyRef = useRef<string | null>(null)

    const form = useForm<FormValues>({
        initialValues: {
            residentId: '',
            residentName: '',
            reason: '',
            weight: '',
            issuedOn: getTodayPenaltyDate(),
        },
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
                    return 'Укажите вес предупреждения'
                }
                const normalized = Number(value)
                if (!Number.isFinite(normalized)) {
                    return 'Укажите вес предупреждения'
                }
                if (normalized < 0.1) {
                    return 'Минимальный вес — 0,1'
                }
                if (normalized > 999999999.9) {
                    return 'Укажите вес предупреждения'
                }
                return null
            },
            issuedOn: (value) => {
                if (!value) {
                    return 'Укажите дату получения'
                }
                if (dayjs(value).isAfter(dayjs(getTodayPenaltyDate()), 'day')) {
                    return 'Дата получения не может быть в будущем'
                }
                return null
            },
        },
    })
    const residentPresetId = residentPreset?.id ?? null
    const residentPresetName = residentPreset?.name ?? null
    const {
        clearFieldError,
        errors,
        getInputProps,
        onSubmit,
        reset,
        setFieldValue,
        values,
    } = form

    const residentError = typeof errors.residentId === 'string'
        ? errors.residentId
        : searchError

    function resetState() {
        reset()
        setFieldValue('issuedOn', getTodayPenaltyDate())
        setSubmitting(false)
        setResidentOptions([])
        setSearchLoading(false)
        setSearchError(null)
        setSubmitError(null)

        if (residentPresetId == null || residentPresetName == null) {
            setFieldValue('residentId', '')
            setFieldValue('residentName', '')
            setSearchValue('')
            return
        }

        setFieldValue('residentId', residentPresetId)
        setFieldValue('residentName', residentPresetName)
        setSearchValue(residentPresetName)
    }

    function handleClose() {
        resetState()
        onClose()
    }

    useEffect(() => {
        if (!opened) {
            wasOpenedRef.current = false
            appliedPresetKeyRef.current = null
            return
        }
    }, [opened])

    useEffect(() => {
        if (!opened) {
            return
        }

        const presetKey = residentPresetId == null || residentPresetName == null
            ? null
            : `${residentPresetId}:${residentPresetName}`
        const shouldApplyPreset = !wasOpenedRef.current || appliedPresetKeyRef.current !== presetKey

        wasOpenedRef.current = true

        if (!shouldApplyPreset) {
            return
        }

        appliedPresetKeyRef.current = presetKey

        const timeoutId = window.setTimeout(() => {
            if (residentPresetId == null || residentPresetName == null) {
                setFieldValue('residentId', '')
                setFieldValue('residentName', '')
                setSearchValue('')
                return
            }

            setFieldValue('residentId', residentPresetId)
            setFieldValue('residentName', residentPresetName)
            setSearchValue(residentPresetName)
        }, 0)

        return () => {
            window.clearTimeout(timeoutId)
        }
    }, [opened, residentPresetId, residentPresetName, setFieldValue])

    useEffect(() => {
        if (!opened) {
            return
        }

        if (lockResident && residentPreset != null) {
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
    }, [debouncedSearch, lockResident, opened, residentPreset])

    return (
        <Drawer
            opened={opened}
            onClose={handleClose}
            position="right"
            size={680}
            title="Выдача предупреждения"
        >
            <FocusTrap.InitialFocus />

            <form
                className={classes.drawerBody}
                onSubmit={onSubmit(async (values) => {
                    setSubmitting(true)
                    setSubmitError(null)

                    try {
                        await createPenalty(toRequest(values))
                        handleClose()
                        await onCreated()
                    } catch (currentError) {
                        if (currentError instanceof ApiError) {
                            setSubmitError(currentError.message)
                        } else {
                            setSubmitError('Не удалось выполнить запрос')
                        }
                    } finally {
                        setSubmitting(false)
                    }
                })}
            >
                <Stack gap="md">
                    <ResidentSearchCombobox
                        searchValue={searchValue}
                        options={toSearchOptions(residentOptions)}
                        loading={searchLoading}
                        error={residentError}
                        placeholder="Житель"
                        selectedId={values.residentId || null}
                        selectedLabel={values.residentName || null}
                        hideDropdownWhenSelected
                        disabled={lockResident}
                        onSearchChange={(value) => {
                            setSearchValue(value)
                            if (value !== values.residentName) {
                                setFieldValue('residentId', '')
                            }
                            setFieldValue('residentName', value)
                            clearFieldError('residentId')
                        }}
                        onOptionSelect={(option) => {
                            setFieldValue('residentId', option.id)
                            setFieldValue('residentName', option.name)
                            setSearchValue(option.name)
                            clearFieldError('residentId')
                        }}
                    />

                    <Textarea
                        minRows={3}
                        maxLength={256}
                        placeholder="Причина выдачи предупреждения"
                        className={classes.field}
                        {...getInputProps('reason')}
                    />

                    <div className={classes.rowFields}>
                        <NumberInput
                            min={0.1}
                            max={999999999.9}
                            step={0.1}
                            decimalScale={1}
                            allowNegative={false}
                            hideControls
                            placeholder="Вес предупреждения"
                            className={classes.field}
                            value={values.weight}
                            onChange={(value) => setFieldValue('weight', value)}
                            error={errors.weight}
                        />

                        <DatePickerInput
                            locale="ru"
                            valueFormat="DD.MM.YYYY"
                            placeholder="Дата получения"
                            className={classes.field}
                            value={values.issuedOn}
                            onChange={(value) => setFieldValue('issuedOn', value)}
                            maxDate={getTodayPenaltyDate()}
                            error={errors.issuedOn}
                        />
                    </div>

                    {submitError ? (
                        <div className={classes.submitError}>{submitError}</div>
                    ) : null}

                    <Group justify="flex-end" mt="sm">
                        <Button type="button" variant="default" onClick={handleClose}>
                            Отменить
                        </Button>
                        <Button type="submit" loading={submitting}>
                            Выдать
                        </Button>
                    </Group>
                </Stack>
            </form>
        </Drawer>
    )
}
