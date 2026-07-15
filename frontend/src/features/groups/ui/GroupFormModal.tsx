import { useEffect, useMemo, useState } from 'react'

import { Select, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import {
    createGroup,
    getGroup,
    getGroupUserOptions,
    updateGroup,
} from '../api/groupsApi'
import type {
    CreateGroupRequest,
    GroupFormValues,
    GroupUserOption,
    UpdateGroupRequest,
} from '../model/types'
import { groupFormValidation } from '../model/validation'

type Props = {
    opened: boolean
    mode: 'create' | 'edit'
    groupId: string | null
    dormitoryId: string
    onClose: () => void
    onSaved: () => Promise<void> | void
}

const initialValues: GroupFormValues = {
    name: '',
    leaderId: null,
}

function toCreateRequest(values: GroupFormValues, dormitoryId: string): CreateGroupRequest {
    return {
        name: values.name.trim(),
        leader_id: values.leaderId,
        dormitory_id: Number(dormitoryId),
    }
}

function toUpdateRequest(values: GroupFormValues, dormitoryId: string): UpdateGroupRequest {
    return toCreateRequest(values, dormitoryId)
}

export function GroupFormModal({
    opened,
    mode,
    groupId,
    dormitoryId,
    onClose,
    onSaved,
}: Props) {
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)
    const [users, setUsers] = useState<GroupUserOption[]>([])

    const form = useForm<GroupFormValues>({
        mode: 'controlled',
        initialValues,
        validate: groupFormValidation,
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
                const [{ users: loadedUsers }, group] = await Promise.all([
                    getGroupUserOptions(dormitoryId),
                    mode === 'edit' && groupId !== null
                        ? getGroup(groupId)
                        : Promise.resolve(null),
                ])

                if (!active) {
                    return
                }

                setUsers(loadedUsers)

                const values: GroupFormValues = group
                    ? {
                        name: group.name,
                        leaderId: group.leader?.id ?? null,
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
                    setSubmitError('Не удалось загрузить данные группы')
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
    }, [dormitoryId, groupId, mode, opened])

    const title = mode === 'create' ? 'Создание группы' : 'Редактирование группы'

    const userOptions = useMemo(
        () =>
            users.map((user) => ({
                value: user.id,
                label: user.name,
            })),
        [users],
    )

    return (
        <EntityFormModal
            opened={opened}
            onClose={onClose}
            title={title}
            loading={loading}
            saving={saving}
            error={submitError}
            onSubmit={form.onSubmit(async (values) => {
                setSubmitError(null)
                setSaving(true)

                try {
                    if (mode === 'create') {
                        await createGroup(toCreateRequest(values, dormitoryId))
                    } else if (groupId !== null) {
                        await updateGroup(groupId, toUpdateRequest(values, dormitoryId))
                    }

                    await onSaved()
                    onClose()
                } catch (error) {
                    if (error instanceof ApiError) {
                        setSubmitError(error.message)
                    } else {
                        setSubmitError('Не удалось сохранить группу')
                    }
                } finally {
                    setSaving(false)
                }
            })}
            size={640}
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
                label="Глава"
                placeholder="Выберите главу"
                searchable
                clearable
                data={userOptions}
                nothingFoundMessage="Житель не найден"
                value={form.values.leaderId}
                onChange={(value) => form.setFieldValue('leaderId', value)}
            />
        </EntityFormModal>
    )
}
