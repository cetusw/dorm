import { useEffect, useState } from 'react'

import { Alert, Button, Checkbox, Group, Loader, Modal, NumberInput, Select, SimpleGrid, Stack, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import modalClasses from '../../../shared/ui/SettingsModal.module.css'
import type { CreateTaskRequest, TaskDetails, UpdateTaskRequest } from '../../task-catalog/model/types'
import { RECURRENCE_OPTIONS } from '../../task-catalog/model/recurrence'
import classes from './DutySettingsCreateTaskModal.module.css'

type FormValues = {
    title: string
    cost: string
    recurrenceInterval: string | null
    includeInCurrentDuty: boolean
}

type Props = {
    opened: boolean
    mode: 'create' | 'edit'
    taskId?: string | null
    task?: TaskDetails | null
    onClose: () => void
    onCreate?: (request: CreateTaskRequest) => Promise<void>
    onUpdate?: (taskId: string, request: UpdateTaskRequest) => Promise<void>
}

const initialValues: FormValues = {
    title: '',
    cost: '',
    recurrenceInterval: '1',
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

export function DutySettingsCreateTaskModal({
    opened,
    mode,
    taskId = null,
    task = null,
    onClose,
    onCreate,
    onUpdate,
}: Props) {
    const [saving, setSaving] = useState(false)
    const [loading, setLoading] = useState(false)
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
            recurrenceInterval: (value) => {
                if (!value) {
                    return 'Выберите частоту'
                }
                return ['0', '1', '2', '4', '12'].includes(value) ? null : 'Выберите корректную частоту'
            },
        },
    })

    useEffect(() => {
        if (!opened) {
            form.setValues(initialValues)
            form.resetDirty(initialValues)
            form.clearErrors()
            setSaving(false)
            setLoading(false)
            setError(null)
        }
    }, [opened])

    useEffect(() => {
        if (!opened || mode !== 'edit') {
            return
        }

        if (!taskId || !task) {
            setLoading(true)
            setError(null)
            return
        }

        const values: FormValues = {
            title: task.title,
            cost: String(task.cost),
            recurrenceInterval: String(task.recurrenceInterval),
            includeInCurrentDuty: false,
        }

        setLoading(false)
        setError(null)
        form.setValues(values)
        form.resetDirty(values)
        form.clearErrors()
    }, [mode, opened, task?.id, task?.title, task?.cost, task?.recurrenceInterval, taskId])

    const title = mode === 'create' ? 'Добавление задачи' : 'Редактирование задачи'
    const submitLabel = mode === 'create' ? 'Добавить' : 'Сохранить'

    return (
        <Modal
            opened={opened}
            onClose={onClose}
            title={<span className={modalClasses.title}>{title}</span>}
            withCloseButton={false}
            centered
            radius={32}
            size={680}
            classNames={{
                header: modalClasses.header,
                body: modalClasses.body,
                content: modalClasses.content,
            }}
        >
            <form
                onSubmit={form.onSubmit(async (values) => {
                    setSaving(true)
                    setError(null)

                    try {
                        const request: CreateTaskRequest = {
                            title: values.title.trim(),
                            cost: Number(values.cost.trim()),
                            recurrenceInterval: Number(values.recurrenceInterval),
                            area_id: 0,
                        }

                        if (mode === 'create') {
                            await onCreate?.({
                                ...request,
                                include_in_current_duty: values.includeInCurrentDuty,
                            })
                        } else if (taskId && task) {
                            await onUpdate?.(taskId, {
                                ...request,
                                area_id: task.area.id,
                            })
                        }

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
                    {loading ? <Loader size="sm" mx="auto" /> : (
                        <>
                            <TextInput
                                placeholder="Название задачи*"
                                maxLength={255}
                                classNames={{
                                    input: modalClasses.input,
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
                                        input: modalClasses.input,
                                    }}
                                    value={form.values.cost}
                                    error={form.errors.cost}
                                    onChange={(value) => form.setFieldValue('cost', value === '' ? '' : String(value))}
                                />

                                <Select
                                    placeholder="Частота*"
                                    data={RECURRENCE_OPTIONS}
                                    allowDeselect={false}
                                    classNames={{
                                        input: modalClasses.input,
                                    }}
                                    value={form.values.recurrenceInterval}
                                    error={form.errors.recurrenceInterval}
                                    onChange={(value) => form.setFieldValue('recurrenceInterval', value)}
                                />
                            </SimpleGrid>

                            {mode === 'create' ? (
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
                        </>
                    )}

                    <Group justify="flex-end" gap="15" className={modalClasses.actions}>
                        <Button
                            type="button"
                            variant="default"
                            onClick={onClose}
                            className={modalClasses.cancelButton}
                        >
                            Отменить
                        </Button>
                        <Button
                            type="submit"
                            loading={saving || loading}
                            className={[modalClasses.submitButton, modalClasses.accentButton].join(' ')}
                        >
                            {submitLabel}
                        </Button>
                    </Group>
                </Stack>
            </form>
        </Modal>
    )
}
