import { useEffect, useState } from 'react'

import { Alert, Button, Checkbox, Group, Modal, NumberInput, SimpleGrid, Stack, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import type { CreateTaskRequest } from '../../task-catalog/model/types'
import classes from './DutySettingsCreateTaskModal.module.css'

type FormValues = {
    title: string
    cost: string
    frequency: string
    oneTime: boolean
    includeInCurrentDuty: boolean
}

type Props = {
    opened: boolean
    onClose: () => void
    onCreate: (request: CreateTaskRequest) => Promise<void>
}

const initialValues: FormValues = {
    title: '',
    cost: '',
    frequency: '',
    oneTime: false,
    includeInCurrentDuty: false,
}

function checkboxStyles(checked: boolean) {
    return {
        body: {
            alignItems: 'center',
        },
        label: {
            paddingInlineStart: '15px',
            minHeight: '24px',
            display: 'flex',
            alignItems: 'center',
            lineHeight: '24px',
        },
        icon: {
            width: '14px',
            height: '14px',
        },
        input: checked
            ? {
                backgroundColor: '#8C8C8C',
                borderColor: '#8C8C8C',
            }
            : undefined,
    }
}

export function DutySettingsCreateTaskModal({ opened, onClose, onCreate }: Props) {
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const form = useForm<FormValues>({
        mode: 'controlled',
        initialValues,
        validate: {
            title: (value) => {
                const trimmed = value.trim()
                if (trimmed.length === 0) {
                    return 'Введите название'
                }
                if ([...trimmed].length > 255) {
                    return 'Название не должно превышать 255 символов'
                }
                return null
            },
            cost: (value) => {
                const trimmed = value.trim()
                if (trimmed.length === 0) {
                    return 'Введите стоимость'
                }
                const parsed = Number(trimmed)
                if (!Number.isInteger(parsed) || parsed <= 0) {
                    return 'Стоимость должна быть больше нуля'
                }
                return null
            },
            frequency: (value, values) => {
                if (values.oneTime) {
                    return null
                }
                const trimmed = value.trim()
                if (trimmed.length === 0) {
                    return 'Введите частоту'
                }
                const parsed = Number(trimmed)
                if (!Number.isInteger(parsed) || parsed <= 0) {
                    return 'Частота должна быть больше нуля'
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
            setSaving(false)
            setError(null)
        }
    }, [opened])

    return (
        <Modal
            opened={opened}
            onClose={onClose}
            title={<span className={classes.title}>Добавление задачи</span>}
            withCloseButton={false}
            centered
            radius={32}
            size={680}
            classNames={{
                header: classes.header,
                body: classes.body,
                content: classes.content,
            }}
        >
            <form
                onSubmit={form.onSubmit(async (values) => {
                    setSaving(true)
                    setError(null)

                    try {
                        await onCreate({
                            title: values.title.trim(),
                            cost: Number(values.cost.trim()),
                            frequency: values.oneTime ? 0 : Number(values.frequency.trim()),
                            area_id: 0,
                            one_time: values.oneTime,
                            include_in_current_duty: values.oneTime ? true : values.includeInCurrentDuty,
                        })
                        onClose()
                    } catch (currentError) {
                        if (currentError instanceof ApiError) {
                            setError(currentError.message)
                        } else if (currentError instanceof Error) {
                            setError(currentError.message)
                        } else {
                            setError('Не удалось сохранить задачу')
                        }
                    } finally {
                        setSaving(false)
                    }
                })}
            >
                <Stack gap="15">
                    {error ? <Alert color="red">{error}</Alert> : null}

                    <TextInput
                        placeholder="Название задачи*"
                        maxLength={255}
                        classNames={{
                            input: classes.input,
                        }}
                        key={form.key('title')}
                        {...form.getInputProps('title')}
                    />

                    <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="15">
                        <NumberInput
                            placeholder="Стоимость*"
                            allowDecimal={false}
                            allowNegative={false}
                            hideControls
                            clampBehavior="strict"
                            classNames={{
                                input: classes.input,
                            }}
                            value={form.values.cost}
                            error={form.errors.cost}
                            onChange={(value) => form.setFieldValue('cost', value === '' ? '' : String(value))}
                        />

                        {!form.values.oneTime ? (
                            <NumberInput
                                placeholder="Частота*"
                                allowDecimal={false}
                                allowNegative={false}
                                hideControls
                                clampBehavior="strict"
                                classNames={{
                                    input: classes.input,
                                }}
                                value={form.values.frequency}
                                error={form.errors.frequency}
                                onChange={(value) => form.setFieldValue('frequency', value === '' ? '' : String(value))}
                            />
                        ) : null}
                    </SimpleGrid>

                    <Checkbox
                        label="Одноразовая задача"
                        checked={form.values.oneTime}
                        size="24px"
                        radius={6}
                        iconColor="#FFFFFF"
                        styles={checkboxStyles(form.values.oneTime)}
                        onChange={(event) => {
                            const checked = event.currentTarget.checked
                            form.setFieldValue('oneTime', checked)
                            if (checked) {
                                form.setFieldValue('includeInCurrentDuty', false)
                                form.setFieldValue('frequency', '')
                            }
                        }}
                        classNames={{
                            root: classes.checkboxRoot,
                            body: classes.checkboxBody,
                            input: classes.checkboxIcon,
                            label: classes.checkboxLabel,
                        }}
                    />

                    {!form.values.oneTime ? (
                        <Checkbox
                            label="Включить в текущее дежурство"
                            checked={form.values.includeInCurrentDuty}
                            size="24px"
                            radius={6}
                            iconColor="#FFFFFF"
                            styles={checkboxStyles(form.values.includeInCurrentDuty)}
                            onChange={(event) => form.setFieldValue('includeInCurrentDuty', event.currentTarget.checked)}
                            classNames={{
                                root: classes.checkboxRoot,
                                body: classes.checkboxBody,
                                input: classes.checkboxIcon,
                                label: classes.checkboxLabel,
                            }}
                        />
                    ) : null}

                    <Group justify="flex-end" gap="15" className={classes.actions}>
                        <Button
                            type="button"
                            variant="default"
                            onClick={onClose}
                            className={classes.cancelButton}
                        >
                            Отменить
                        </Button>
                        <Button
                            type="submit"
                            loading={saving}
                            className={classes.submitButton}
                        >
                            Добавить
                        </Button>
                    </Group>
                </Stack>
            </form>
        </Modal>
    )
}
