import { useEffect, useMemo, useState } from 'react'

import dayjs from 'dayjs'
import 'dayjs/locale/ru'
import { DatePickerInput } from '@mantine/dates'
import { TextInput } from '@mantine/core'

import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import { createDormitoryDutyWeek } from '../api/currentDutyApi'

type Props = {
    opened: boolean
    dormitoryId: string
    onClose: () => void
    onCreated: () => Promise<void> | void
}

export function CreateDutyWeekModal({
    opened,
    dormitoryId,
    onClose,
    onCreated,
}: Props) {
    const [startDate, setStartDate] = useState<string | null>(dayjs().format('YYYY-MM-DD'))
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        if (!opened) {
            setStartDate(dayjs().format('YYYY-MM-DD'))
            setSaving(false)
            setError(null)
        }
    }, [opened])

    const endDate = useMemo(() => {
        if (!startDate) {
            return null
        }

        return dayjs(startDate).add(7, 'day').format('YYYY-MM-DD')
    }, [startDate])

    return (
        <EntityFormModal
            opened={opened}
            onClose={onClose}
            title="Новая дежурная неделя"
            saving={saving}
            error={error}
            size={560}
            onSubmit={async (event) => {
                event.preventDefault()

                if (!startDate || !endDate) {
                    setError('Выберите дату начала')
                    return
                }

                setSaving(true)
                setError(null)

                try {
                    await createDormitoryDutyWeek(
                        dormitoryId,
                        startDate,
                        endDate,
                    )
                    await onCreated()
                    onClose()
                } catch (currentError) {
                    if (currentError instanceof ApiError) {
                        setError(currentError.message)
                    } else {
                        setError('Не удалось создать дежурную неделю')
                    }
                } finally {
                    setSaving(false)
                }
            }}
        >
            <DatePickerInput
                label="Начало дежурства"
                placeholder="Выберите дату"
                locale="ru"
                valueFormat="DD.MM.YYYY"
                withAsterisk
                value={startDate}
                onChange={setStartDate}
            />

            <TextInput
                label="Конец дежурства"
                readOnly
                value={endDate ? dayjs(endDate).format('DD.MM.YYYY') : ''}
            />
        </EntityFormModal>
    )
}
