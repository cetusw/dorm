import { useEffect, useMemo, useState } from 'react'

import { Select, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'

import { getAreas } from '../../areas/api/areasApi'
import type { AreaListItem } from '../../areas/model/types'
import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import {
    createTaskDefinition,
    getTaskDefinition,
    updateTaskDefinition,
} from '../api/taskCatalogApi'
import type {
    CreateTaskRequest,
    TaskFormValues,
    UpdateTaskRequest,
} from '../model/types'
import { taskFormValidation } from '../model/validation'

type Props = {
    opened: boolean
    mode: 'create' | 'edit'
    taskId: string | null
    dormitoryId: string
    onClose: () => void
    onSaved: () => Promise<void> | void
}

const initialValues: TaskFormValues = {
    title: '',
    cost: '',
    frequency: '',
    areaId: null,
}

function toRequest(values: TaskFormValues): CreateTaskRequest | UpdateTaskRequest {
    return {
        title: values.title.trim(),
        cost: Number(values.cost.trim()),
        frequency: Number(values.frequency.trim()),
        area_id: Number(values.areaId),
    }
}

export function TaskFormModal({
    opened,
    mode,
    taskId,
    dormitoryId,
    onClose,
    onSaved,
}: Props) {
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)
    const [areas, setAreas] = useState<AreaListItem[]>([])

    const form = useForm<TaskFormValues>({
        mode: 'controlled',
        initialValues,
        validate: {
            title: taskFormValidation.title,
            cost: taskFormValidation.cost,
            frequency: taskFormValidation.frequency,
            areaId: (value) => (value ? null : 'Выберите территорию'),
        },
    })

    useEffect(() => {
        if (!opened) {
            form.setValues(initialValues)
            form.resetDirty(initialValues)
            form.clearErrors()
            setSubmitError(null)
            setLoading(false)
            setSaving(false)
            return
        }

        let active = true

        async function loadModalData() {
            setLoading(true)
            setSubmitError(null)

            try {
                const [{ areas: loadedAreas }, task] = await Promise.all([
                    getAreas(dormitoryId),
                    mode === 'edit' && taskId !== null
                        ? getTaskDefinition(dormitoryId, taskId)
                        : Promise.resolve(null),
                ])

                if (!active) {
                    return
                }

                setAreas(loadedAreas)

                const values: TaskFormValues = task
                    ? {
                        title: task.title,
                        cost: String(task.cost),
                        frequency: String(task.frequency),
                        areaId: String(task.area.id),
                    }
                    : initialValues

                form.setValues(values)
                form.resetDirty(values)
                form.clearErrors()
            } catch (error) {
                if (!active) {
                    return
                }

                if (error instanceof ApiError) {
                    setSubmitError(error.message)
                } else {
                    setSubmitError('Не удалось загрузить данные задачи')
                }
            } finally {
                if (active) {
                    setLoading(false)
                }
            }
        }

        void loadModalData()

        return () => {
            active = false
        }
    }, [dormitoryId, mode, opened, taskId])

    const areaOptions = useMemo(
        () =>
            areas.map((area) => ({
                value: String(area.id),
                label: area.group ? `${area.name} (${area.group.name})` : area.name,
            })),
        [areas],
    )

    const title = mode === 'create' ? 'Создание задачи' : 'Редактирование задачи'

    return (
        <EntityFormModal
            opened={opened}
            onClose={onClose}
            title={title}
            loading={loading}
            saving={saving}
            error={submitError}
            size={640}
            onSubmit={form.onSubmit(async (values) => {
                setSubmitError(null)
                setSaving(true)

                try {
                    if (mode === 'create') {
                        await createTaskDefinition(dormitoryId, toRequest(values))
                    } else if (taskId !== null) {
                        await updateTaskDefinition(dormitoryId, taskId, toRequest(values))
                    }

                    await onSaved()
                    onClose()
                } catch (error) {
                    if (error instanceof ApiError) {
                        setSubmitError(error.message)
                    } else {
                        setSubmitError('Не удалось сохранить задачу')
                    }
                } finally {
                    setSaving(false)
                }
            })}
        >
            <TextInput
                label="Название"
                placeholder="Название"
                withAsterisk
                maxLength={255}
                key={form.key('title')}
                {...form.getInputProps('title')}
            />

            <TextInput
                label="Стоимость"
                placeholder="Стоимость"
                withAsterisk
                inputMode="numeric"
                key={form.key('cost')}
                {...form.getInputProps('cost')}
            />

            <TextInput
                label="Частота"
                placeholder="Частота"
                withAsterisk
                inputMode="numeric"
                key={form.key('frequency')}
                {...form.getInputProps('frequency')}
            />

            <Select
                label="Территория"
                placeholder="Выберите территорию"
                searchable
                withAsterisk
                data={areaOptions}
                nothingFoundMessage="Территория не найдена"
                value={form.values.areaId}
                onChange={(value) => form.setFieldValue('areaId', value)}
                error={form.errors.areaId}
            />
        </EntityFormModal>
    )
}
