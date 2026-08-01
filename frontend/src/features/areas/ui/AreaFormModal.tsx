import { useEffect, useMemo, useState } from 'react'

import { Select, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'

import { getGroups } from '../../groups/api/groupsApi'
import type { GroupListItem } from '../../groups/model/types'
import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import modalClasses from '../../../shared/ui/SettingsModal.module.css'
import { createArea, getArea, updateArea } from '../api/areasApi'
import type {
    AreaDetails,
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
    fixedGroupId?: string | null
    hideGroupField?: boolean
    requireFloor?: boolean
    variant?: 'default' | 'settings'
    loadGroups?: () => Promise<{ groups: GroupListItem[] }>
    loadArea?: (areaId: number) => Promise<AreaDetails>
    createAreaRequest?: (request: CreateAreaRequest) => Promise<AreaDetails>
    updateAreaRequest?: (areaId: number, request: UpdateAreaRequest) => Promise<AreaDetails>
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
    fixedGroupId = null,
    hideGroupField = false,
    requireFloor = false,
    variant = 'default',
    loadGroups,
    loadArea,
    createAreaRequest,
    updateAreaRequest,
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
            floor: (value) => {
                const baseError = areaFormValidation.floor(value)
                if (baseError) {
                    return baseError
                }

                if (requireFloor && value.trim().length === 0) {
                    return 'Введите этаж'
                }

                return null
            },
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
                    hideGroupField
                        ? Promise.resolve({ groups: [] as GroupListItem[] })
                        : (loadGroups ?? (() => getGroups(dormitoryId)))(),
                    mode === 'edit' && areaId !== null
                        ? (loadArea ?? getArea)(areaId)
                        : Promise.resolve(null),
                ])

                if (!active) {
                    return
                }

                setGroups(loadedGroups)

                const values: AreaFormValues = area
                    ? {
                        name: area.name,
                        groupId: fixedGroupId ?? area.group?.id ?? null,
                        floor: area.floor == null ? '' : String(area.floor),
                    }
                    : {
                        ...initialValues,
                        groupId: fixedGroupId,
                    }

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
    }, [areaId, dormitoryId, fixedGroupId, hideGroupField, loadArea, loadGroups, mode, opened])

    const groupOptions = useMemo(
        () =>
            groups.map((group) => ({
                value: group.id,
                label: group.name,
            })),
        [groups],
    )

    const title = mode === 'create' ? 'Создание территории' : 'Редактирование территории'
    const submitLabel = mode === 'create' ? 'Добавить' : 'Сохранить'
    const isSettingsVariant = variant === 'settings'

    return (
        <EntityFormModal
            opened={opened}
            onClose={onClose}
            title={isSettingsVariant ? <span className={modalClasses.title}>{title}</span> : title}
            loading={loading}
            saving={saving}
            error={submitError}
            size={isSettingsVariant ? 680 : 640}
            withCloseButton={!isSettingsVariant}
            modalClassNames={isSettingsVariant ? {
                header: modalClasses.header,
                body: modalClasses.body,
                content: modalClasses.content,
            } : undefined}
            contentGap={isSettingsVariant ? 15 : 'md'}
            actionsClassName={isSettingsVariant ? modalClasses.actions : undefined}
            cancelButtonClassName={isSettingsVariant ? modalClasses.cancelButton : undefined}
            submitButtonClassName={isSettingsVariant ? modalClasses.submitButton : undefined}
            submitLabel={submitLabel}
            onSubmit={form.onSubmit(async (values) => {
                setSubmitError(null)
                setSaving(true)

                try {
                    if (mode === 'create') {
                        await (createAreaRequest ?? ((request) => createArea(dormitoryId, request)))(toCreateRequest({
                            ...values,
                            groupId: fixedGroupId ?? values.groupId,
                        }))
                    } else if (areaId !== null) {
                        await (updateAreaRequest ?? ((targetAreaId, request) => updateArea(dormitoryId, targetAreaId, request)))(
                            areaId,
                            toUpdateRequest({
                                ...values,
                                groupId: fixedGroupId ?? values.groupId,
                            }),
                        )
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
                label={isSettingsVariant ? undefined : 'Название'}
                placeholder={isSettingsVariant ? 'Название территории*' : 'Название'}
                withAsterisk={!isSettingsVariant}
                maxLength={255}
                classNames={isSettingsVariant ? {
                    input: modalClasses.input,
                } : undefined}
                key={form.key('name')}
                {...form.getInputProps('name')}
            />

            {!hideGroupField && (
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
            )}

            <TextInput
                label={isSettingsVariant ? undefined : 'Этаж'}
                placeholder={isSettingsVariant ? (requireFloor ? 'Этаж*' : 'Этаж') : 'Этаж'}
                withAsterisk={requireFloor && !isSettingsVariant}
                inputMode="numeric"
                classNames={isSettingsVariant ? {
                    input: modalClasses.input,
                } : undefined}
                key={form.key('floor')}
                {...form.getInputProps('floor')}
            />
        </EntityFormModal>
    )
}
