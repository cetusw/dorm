import { useEffect, useState } from 'react'

import { Button, Group, PasswordInput, SimpleGrid, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import {
    createResident,
    getResident,
    updateResident,
} from '../api/residentsApi'
import { generateResidentLogin, generateResidentPassword } from '../model/transliteration'
import type {
    CreateResidentRequest,
    ResidentFormValues,
    UpdateResidentRequest,
} from '../model/types'
import { residentFormValidation } from '../model/validation'

type Props = {
    opened: boolean
    mode: 'create' | 'edit'
    residentId: string | null
    dormitoryId: string
    onClose: () => void
    onSaved: () => Promise<void> | void
}

const initialValues: ResidentFormValues = {
    lastName: '',
    firstName: '',
    middleName: '',
    login: '',
    password: '',
    floor: '',
    roomNumber: '',
}

function toNullableString(value: string): string | null {
    const trimmed = value.trim()
    return trimmed.length === 0 ? null : trimmed
}

function toNullableNumber(value: string): number | null {
    const trimmed = value.trim()
    return trimmed.length === 0 ? null : Number(trimmed)
}

function toCreateRequest(values: ResidentFormValues, dormitoryId: string): CreateResidentRequest {
    return {
        first_name: values.firstName.trim(),
        last_name: values.lastName.trim(),
        middle_name: toNullableString(values.middleName),
        login: values.login.trim(),
        password: values.password.trim(),
        dormitory_id: Number(dormitoryId),
        floor: toNullableNumber(values.floor),
        room_number: toNullableString(values.roomNumber),
    }
}

function toUpdateRequest(values: ResidentFormValues, dormitoryId: string): UpdateResidentRequest {
    return toCreateRequest(values, dormitoryId)
}

export function ResidentFormModal({
    opened,
    mode,
    residentId,
    dormitoryId,
    onClose,
    onSaved,
}: Props) {
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)
    const [loginEditedManually, setLoginEditedManually] = useState(false)

    const form = useForm<ResidentFormValues>({
        mode: 'controlled',
        initialValues,
        validate: residentFormValidation,
    })

    useEffect(() => {
        if (loginEditedManually) {
            return
        }

        const generatedLogin = generateResidentLogin(form.values.firstName, form.values.lastName)
        if (generatedLogin !== form.values.login) {
            form.setFieldValue('login', generatedLogin)
        }
    }, [form.values.firstName, form.values.lastName, form.values.login, loginEditedManually])

    useEffect(() => {
        if (!opened) {
            form.setValues(initialValues)
            form.resetDirty(initialValues)
            form.clearErrors()
            setSubmitError(null)
            setLoading(false)
            setSaving(false)
            setLoginEditedManually(false)
            return
        }

        let active = true

        async function loadResident() {
            if (mode !== 'edit' || residentId === null) {
                const createValues = {
                    ...initialValues,
                    password: generateResidentPassword(),
                }
                setLoading(false)
                form.setValues(createValues)
                form.resetDirty(createValues)
                form.clearErrors()
                setLoginEditedManually(false)
                return
            }

            setLoading(true)
            setSubmitError(null)

            try {
                const resident = await getResident(residentId)

                if (!active) {
                    return
                }

                const values: ResidentFormValues = {
                    lastName: resident.last_name,
                    firstName: resident.first_name,
                    middleName: resident.middle_name ?? '',
                    login: resident.login,
                    password: '',
                    floor: resident.floor == null ? '' : String(resident.floor),
                    roomNumber: resident.room_number ?? '',
                }

                form.setValues(values)
                form.resetDirty(values)
                form.clearErrors()
                setLoginEditedManually(true)
            } catch (error) {
                if (!active) {
                    return
                }

                if (error instanceof ApiError) {
                    setSubmitError(error.message)
                } else {
                    setSubmitError('Не удалось загрузить данные жителя')
                }
            } finally {
                if (active) {
                    setLoading(false)
                }
            }
        }

        void loadResident()

        return () => {
            active = false
        }
    }, [mode, opened, residentId])

    const title = mode === 'create' ? 'Создание жителя' : 'Редактирование жителя'

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
                        await createResident(toCreateRequest(values, dormitoryId))
                    } else if (residentId !== null) {
                        await updateResident(residentId, toUpdateRequest(values, dormitoryId))
                    }

                    await onSaved()
                    onClose()
                } catch (error) {
                    if (error instanceof ApiError) {
                        setSubmitError(error.message)
                    } else {
                        setSubmitError('Не удалось сохранить жителя')
                    }
                } finally {
                    setSaving(false)
                }
            })}
        >
            <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="md" verticalSpacing="md">
                <TextInput
                    label="Фамилия"
                    placeholder="Фамилия"
                    withAsterisk
                    maxLength={255}
                    key={form.key('lastName')}
                    {...form.getInputProps('lastName')}
                />

                <TextInput
                    label="Имя"
                    placeholder="Имя"
                    withAsterisk
                    maxLength={255}
                    key={form.key('firstName')}
                    {...form.getInputProps('firstName')}
                />
            </SimpleGrid>

            <TextInput
                label="Отчество"
                placeholder="Отчество"
                maxLength={255}
                key={form.key('middleName')}
                {...form.getInputProps('middleName')}
            />

            <TextInput
                label="Логин"
                placeholder="Логин"
                withAsterisk
                maxLength={255}
                key={form.key('login')}
                {...form.getInputProps('login')}
                onChange={(event) => {
                    setLoginEditedManually(true)
                    form.setFieldValue('login', event.currentTarget.value)
                }}
            />

            <Group align="flex-end" wrap="nowrap">
                <PasswordInput
                    label="Пароль"
                    placeholder="Пароль"
                    withAsterisk
                    style={{ flex: 1 }}
                    key={form.key('password')}
                    {...form.getInputProps('password')}
                />
                <Button
                    type="button"
                    variant="default"
                    onClick={() => form.setFieldValue('password', generateResidentPassword())}
                >
                    Сгенерировать
                </Button>
            </Group>

            <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="md" verticalSpacing="md">
                <TextInput
                    label="Этаж"
                    placeholder="Этаж"
                    inputMode="numeric"
                    key={form.key('floor')}
                    {...form.getInputProps('floor')}
                />

                <TextInput
                    label="Комната"
                    placeholder="Комната"
                    maxLength={255}
                    key={form.key('roomNumber')}
                    {...form.getInputProps('roomNumber')}
                />
            </SimpleGrid>
        </EntityFormModal>
    )
}
