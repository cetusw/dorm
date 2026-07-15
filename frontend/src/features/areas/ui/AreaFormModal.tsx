import { useEffect, useMemo, useState } from 'react'

import { Select, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'

import { getGroups } from '../../groups/api/groupsApi'
import type { GroupListItem } from '../../groups/model/types'
import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import { createArea, getArea, updateArea } from '../api/areasApi'
import type {
    AreaFormValues,
    CreateAreaRequest,
    UpdateAreaRequest,
} from '../model/types'
import { areaFormValidation } from '../model/validation'

type Props = {
    opened: boolean
    mode: 'create' | 'edit'
    areaId: number | null
    dormitoryId: string
    onClose: () => void
    onSaved: () => Promise<void> | void
}

const initialValues: AreaFormValues = {
    name: '',
    groupId: null,
    floor: '',
}

function toNullableNumber(value: string): number | null {
    const trimmed = value.trim()
    return trimmed.length === 0 ? null : Number(trimmed)
}

function toCreateRequest(values: AreaFormValues): CreateAreaRequest {
    return {
        name: values.name.trim(),
        group_id: values.groupId,
        floor: toNullableNumber(values.floor),
    }
}

function toUpdateRequest(values: AreaFormValues): UpdateAreaRequest {
    return toCreateRequest(values)
}

export function AreaFormModal({
    opened,
    mode,
    areaId,
    dormitoryId,
    onClose,
    onSaved,
}: Props) {
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)
    const [groups, setGroups] = useState<GroupListItem[]>([])

    const form = useForm<AreaFormValues>({
        mode: 'controlled',
        initialValues,
        validate: {
            name: areaFormValidation.name,
            floor: areaFormValidation.floor,
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
                const [{ groups: loadedGroups }, area] = await Promise.all([
                    getGroups(dormitoryId),
                    mode === 'edit' && areaId !== null
                        ? getArea(areaId)
                        : Promise.resolve(null),
                ])

                if (!active) {
                    return
                }

                setGroups(loadedGroups)

                const values: AreaFormValues = area
                    ? {
                        name: area.name,
                        groupId: area.group?.id ?? null,
                        floor: area.floor == null ? '' : String(area.floor),
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
                    setSubmitError('Не удалось загрузить данные территории')
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
    }, [areaId, dormitoryId, mode, opened])

    const groupOptions = useMemo(
        () =>
            groups.map((group) => ({
                value: group.id,
                label: group.name,
            })),
        [groups],
    )

    const title = mode === 'create' ? 'Создание территории' : 'Редактирование территории'

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
                        await createArea(dormitoryId, toCreateRequest(values))
                    } else if (areaId !== null) {
                        await updateArea(dormitoryId, areaId, toUpdateRequest(values))
                    }

                    await onSaved()
                    onClose()
                } catch (error) {
                    if (error instanceof ApiError) {
                        setSubmitError(error.message)
                    } else {
                        setSubmitError('Не удалось сохранить территорию')
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
                key={form.key('name')}
                {...form.getInputProps('name')}
            />

            <Select
                label="Группа"
                placeholder="Выберите группу"
                searchable
                clearable
                data={groupOptions}
                nothingFoundMessage="Группа не найдена"
                value={form.values.groupId}
                onChange={(value) => form.setFieldValue('groupId', value)}
            />

            <TextInput
                label="Этаж"
                placeholder="Этаж"
                inputMode="numeric"
                key={form.key('floor')}
                {...form.getInputProps('floor')}
            />
        </EntityFormModal>
    )
}
