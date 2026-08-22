import { useEffect, useMemo, useState } from 'react'

import { NumberInput, Select, TextInput } from '@mantine/core'
import { DatePickerInput } from '@mantine/dates'
import { useForm } from '@mantine/form'
import { useDebouncedValue } from '@mantine/hooks'
import dayjs from 'dayjs'

import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import { ResidentSearchCombobox, type ResidentSearchOption } from '../../../shared/ui/ResidentSearchCombobox'
import modalClasses from '../../../shared/ui/SettingsModal.module.css'
import {
    createIndividualTask,
    getIndividualTaskAreas,
    searchIndividualTaskResidents,
    updateIndividualTask,
} from '../api/individualTasksApi'
import type { IndividualTask, IndividualTaskResidentOption } from '../model/types'

type Props = {
    opened: boolean
    task?: IndividualTask | null
    residentPreset?: { id: string; name: string } | null
    onClose: () => void
    onSaved: () => Promise<void> | void
}

type Values = {
    residentId: string
    residentName: string
    title: string
    areaId: string | null
    weight: number | string
    deadline: string | null
}

const emptyValues: Values = {
    residentId: '',
    residentName: '',
    title: '',
    areaId: null,
    weight: '',
    deadline: null,
}

function errorMessage(error: unknown): string {
    return error instanceof ApiError ? error.message : 'Не удалось сохранить индивидуальную задачу'
}

function initialValues(task: IndividualTask | null | undefined, residentPreset: Props['residentPreset']): Values {
    if (task) {
        return {
            residentId: task.resident.id,
            residentName: task.resident.name,
            title: task.title,
            areaId: task.area ? String(task.area.id) : null,
            weight: task.redemption_weight,
            deadline: task.deadline,
        }
    }

    return {
        ...emptyValues,
        residentId: residentPreset?.id ?? '',
        residentName: residentPreset?.name ?? '',
    }
}

export function IndividualTaskFormModal({
    opened,
    task,
    residentPreset = null,
    onClose,
    onSaved,
}: Props) {
    const editing = Boolean(task)
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [options, setOptions] = useState<IndividualTaskResidentOption[]>([])
    const [selectedResident, setSelectedResident] = useState<IndividualTaskResidentOption | null>(null)
    const [areas, setAreas] = useState<{ value: string; label: string }[]>([])
    const [search, setSearch] = useState('')
    const [debouncedSearch] = useDebouncedValue(search, 300)
    const [searchLoading, setSearchLoading] = useState(false)

    const form = useForm<Values>({
        mode: 'controlled',
        initialValues: emptyValues,
        validate: {
            residentId: (value) => value ? null : 'Выберите жителя',
            title: (value) => {
                const title = value.trim()
                if (!title) return 'Укажите название'
                return Array.from(title).length > 255 ? 'Название не должно превышать 255 символов' : null
            },
            weight: (value, values) => {
                const maxWeight = maximumWeight(selectedResident, editing, task, values.residentId)
                if (maxWeight === 0) return null
                if (value === '' || !Number.isFinite(Number(value)) || Number(value) < 0 || Number(value) > maxWeight) {
                    return `Укажите вес от 0 до ${maxWeight}`
                }
                return null
            },
        },
    })

    const maxWeight = maximumWeight(selectedResident, editing, task, form.values.residentId)
    const searchOptions = useMemo<ResidentSearchOption[]>(
        () => options.map((option) => ({ id: option.id, name: option.name })),
        [options],
    )

    useEffect(() => {
        if (!opened) return

        const values = initialValues(task, residentPreset)
        form.setValues(values)
        form.resetDirty(values)
        setSearch(values.residentName)
        setSelectedResident(null)
        setAreas([])
        setError(null)
    }, [opened, task, residentPreset?.id, residentPreset?.name])

    useEffect(() => {
        if (!opened) return

        let cancelled = false
        setSearchLoading(true)
        void searchIndividualTaskResidents(debouncedSearch)
            .then((nextOptions) => {
                if (cancelled) return
                setOptions(nextOptions)
                setSelectedResident((current) => nextOptions.find((option) => option.id === current?.id)
                    ?? nextOptions.find((option) => option.id === form.values.residentId)
                    ?? current)
            })
            .catch(() => {
                if (!cancelled) setOptions([])
            })
            .finally(() => {
                if (!cancelled) setSearchLoading(false)
            })

        return () => { cancelled = true }
    }, [debouncedSearch, form.values.residentId, opened])

    useEffect(() => {
        if (!selectedResident) {
            setAreas([])
            return
        }

        let cancelled = false
        void getIndividualTaskAreas(selectedResident.dormitory_id)
            .then((items) => {
                if (cancelled) return
                const nextAreas = items.map((area) => ({
                    value: String(area.id),
                    label: area.floor === null ? area.name : `${area.floor} этаж. ${area.name}`,
                }))
                setAreas(nextAreas)
                if (form.values.areaId && !nextAreas.some((area) => area.value === form.values.areaId)) {
                    form.setFieldValue('areaId', null)
                }
            })
            .catch(() => {
                if (!cancelled) setAreas([])
            })

        return () => { cancelled = true }
    }, [form.values.areaId, selectedResident?.dormitory_id])

    async function submit(values: Values) {
        setSaving(true)
        setError(null)

        try {
            const request = {
                resident_id: values.residentId,
                title: values.title.trim(),
                area_id: values.areaId ? Number(values.areaId) : null,
                redemption_weight: maxWeight === 0 ? 0 : Number(values.weight),
                deadline: values.deadline,
                version: task?.version ?? 0,
            }
            if (task) await updateIndividualTask(task.id, request)
            else await createIndividualTask(request)

            await onSaved()
            onClose()
        } catch (currentError) {
            setError(errorMessage(currentError))
            if (currentError instanceof ApiError && currentError.status === 409) {
                void onSaved()
            }
        } finally {
            setSaving(false)
        }
    }

    function selectResident(option: ResidentSearchOption) {
        const changed = option.id !== form.values.residentId
        const resident = options.find((item) => item.id === option.id) ?? null

        setSelectedResident(resident)
        form.setValues({
            ...form.values,
            residentId: option.id,
            residentName: option.name,
            areaId: changed ? null : form.values.areaId,
            weight: changed ? '' : form.values.weight,
        })
        setSearch(option.name)
    }

    return (
        <EntityFormModal
            opened={opened}
            onClose={onClose}
            title={<span className={modalClasses.title}>{editing ? 'Редактирование индивидуальной задачи' : 'Выдача индивидуальной задачи'}</span>}
            saving={saving}
            error={error}
            size={680}
            withCloseButton={false}
            modalClassNames={{ header: modalClasses.header, body: modalClasses.body, content: modalClasses.content }}
            contentGap={15}
            actionsClassName={modalClasses.actions}
            cancelButtonClassName={modalClasses.cancelButton}
            submitButtonClassName={[modalClasses.submitButton, modalClasses.accentButton].join(' ')}
            submitLabel={editing ? 'Редактировать' : 'Добавить'}
            onSubmit={form.onSubmit(submit)}
        >
            <ResidentSearchCombobox
                overlayOpened={opened}
                placeholder="Житель*"
                searchValue={form.values.residentName}
                selectedId={form.values.residentId || null}
                selectedLabel={form.values.residentName || null}
                hideDropdownWhenSelected
                options={searchOptions}
                loading={searchLoading}
                error={typeof form.errors.residentId === 'string' ? form.errors.residentId : null}
                inputClassName={modalClasses.input}
                onSearchChange={(value) => {
                    setSearch(value)
                    if (value !== form.values.residentName) {
                        setSelectedResident(null)
                        form.setValues({
                            ...form.values,
                            residentId: '',
                            residentName: value,
                            areaId: null,
                            weight: '',
                        })
                    }
                }}
                onOptionSelect={selectResident}
            />
            <TextInput placeholder="Название*" maxLength={255} classNames={{ input: modalClasses.input }} {...form.getInputProps('title')} />
            <Select
                placeholder="Территория"
                data={areas}
                value={form.values.areaId}
                onChange={(value) => form.setFieldValue('areaId', value)}
                classNames={{ input: modalClasses.input }}
                clearable
            />
            {maxWeight > 0 ? (
                <NumberInput
                    placeholder="Вес погашения*"
                    hideControls
                    decimalScale={1}
                    step={0.1}
                    allowNegative={false}
                    min={0}
                    max={maxWeight}
                    clampBehavior="strict"
                    value={form.values.weight}
                    error={form.errors.weight}
                    onChange={(value) => form.setFieldValue('weight', value === '' ? '' : value)}
                    classNames={{ input: modalClasses.input }}
                />
            ) : null}
            <DatePickerInput
                placeholder="Дедлайн"
                locale="ru"
                valueFormat="DD.MM.YYYY"
                minDate={dayjs().startOf('day').toDate()}
                value={form.values.deadline ? dayjs(form.values.deadline).toDate() : null}
                onChange={(value) => form.setFieldValue('deadline', value ? dayjs(value).format('YYYY-MM-DD') : null)}
                classNames={{ input: modalClasses.input }}
                clearable
            />
        </EntityFormModal>
    )
}

function maximumWeight(
    resident: IndividualTaskResidentOption | null,
    editing: boolean,
    task: IndividualTask | null | undefined,
    residentID: string,
): number {
    if (!resident) return 0

    return resident.available_redemption_weight
        + (editing && task?.resident.id === residentID ? task.redemption_weight : 0)
}
