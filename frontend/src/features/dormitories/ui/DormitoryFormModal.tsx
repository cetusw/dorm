import { useEffect, useMemo, useState } from 'react'

import {
    Select,
    SimpleGrid,
    TextInput,
} from '@mantine/core'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import {
    createDormitory,
    getDormitory,
    getUserOptions,
    updateDormitory,
} from '../api/dormitoriesApi'
import { dormitoryFormValidation } from '../model/validation'
import type { DormitoryFormValues, UserOption } from '../model/types'

type Props = {
    opened: boolean
    mode: 'create' | 'edit'
    dormitoryId: number | null
    onClose: () => void
    onSaved: () => Promise<void> | void
}

const initialValues: DormitoryFormValues = {
    name: '',
    city: '',
    streetType: '',
    streetName: '',
    houseNumber: '',
    leaderId: null,
}

function toRequest(values: DormitoryFormValues) {
    return {
        name: values.name.trim(),
        city: values.city.trim(),
        street_type: values.streetType.trim(),
        street_name: values.streetName.trim(),
        house_number: values.houseNumber.trim(),
        leader_id: values.leaderId,
    }
}

export function DormitoryFormModal({
    opened,
    mode,
    dormitoryId,
    onClose,
    onSaved,
}: Props) {
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)
    const [users, setUsers] = useState<UserOption[]>([])

    const form = useForm<DormitoryFormValues>({
        mode: 'controlled',
        initialValues,
        validate: dormitoryFormValidation,
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
                const [{ users: loadedUsers }, dormitory] = await Promise.all([
                    getUserOptions(),
                    mode === 'edit' && dormitoryId !== null
                        ? getDormitory(dormitoryId)
                        : Promise.resolve(null),
                ])

                if (!active) {
                    return
                }

                setUsers(loadedUsers)

                const values: DormitoryFormValues = dormitory
                    ? {
                        name: dormitory.name,
                        city: dormitory.city,
                        streetType: dormitory.streetType,
                        streetName: dormitory.streetName,
                        houseNumber: dormitory.houseNumber,
                        leaderId: dormitory.leader?.id ?? null,
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
                    setSubmitError('Не удалось загрузить данные формы')
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
    }, [dormitoryId, mode, opened])

    const title = mode === 'create' ? 'Создание общежития' : 'Редактирование общежития'

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
                        await createDormitory(toRequest(values))
                    } else if (dormitoryId !== null) {
                        await updateDormitory(dormitoryId, toRequest(values))
                    }

                    await onSaved()
                    onClose()
                } catch (error) {
                    if (error instanceof ApiError) {
                        setSubmitError(error.message)
                    } else {
                        setSubmitError('Не удалось сохранить общежитие')
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

            <TextInput
                label="Город"
                placeholder="Город"
                withAsterisk
                maxLength={255}
                key={form.key('city')}
                {...form.getInputProps('city')}
            />

            <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="md" verticalSpacing="md">
                <TextInput
                    label="Тип улицы"
                    placeholder="Тип"
                    withAsterisk
                    maxLength={100}
                    key={form.key('streetType')}
                    {...form.getInputProps('streetType')}
                />

                <TextInput
                    label="Название улицы"
                    placeholder="Название ул."
                    withAsterisk
                    maxLength={100}
                    key={form.key('streetName')}
                    {...form.getInputProps('streetName')}
                />

                <TextInput
                    label="Номер дома"
                    placeholder="Номер"
                    withAsterisk
                    maxLength={50}
                    key={form.key('houseNumber')}
                    {...form.getInputProps('houseNumber')}
                />
            </SimpleGrid>

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
