import { useEffect, useMemo, useState } from 'react'
import { DatePickerInput } from '@mantine/dates'
import { NumberInput, Select, TextInput } from '@mantine/core'
import { useDebouncedValue } from '@mantine/hooks'
import { useForm } from '@mantine/form'
import dayjs from 'dayjs'
import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import { ResidentSearchCombobox, type ResidentSearchOption } from '../../../shared/ui/ResidentSearchCombobox'
import modalClasses from '../../../shared/ui/SettingsModal.module.css'
import { createIndividualTask, getIndividualTaskAreas, searchIndividualTaskResidents, updateIndividualTask } from '../api/individualTasksApi'
import type { IndividualTask, IndividualTaskResidentOption } from '../model/types'

type Props = { opened: boolean; task?: IndividualTask | null; residentPreset?: { id: string; name: string } | null; onClose: () => void; onSaved: () => Promise<void> | void }
type Values = { residentId: string; residentName: string; title: string; areaId: string | null; weight: number | string; deadline: string | null }
const empty: Values = { residentId: '', residentName: '', title: '', areaId: null, weight: '', deadline: null }
function errorText(error: unknown) { return error instanceof ApiError ? error.message : 'Не удалось сохранить индивидуальную задачу' }

export function IndividualTaskFormModal({ opened, task, residentPreset = null, onClose, onSaved }: Props) {
    const editing = Boolean(task)
    const [saving, setSaving] = useState(false); const [error, setError] = useState<string | null>(null)
    const [options, setOptions] = useState<IndividualTaskResidentOption[]>([]); const [areas, setAreas] = useState<{ value: string; label: string }[]>([])
    const [search, setSearch] = useState(''); const [debounced] = useDebouncedValue(search, 300); const [searchLoading, setSearchLoading] = useState(false)
    const form = useForm<Values>({ mode: 'controlled', initialValues: empty, validate: { residentId: (v) => v ? null : 'Выберите жителя', title: (v) => { const t = v.trim(); return !t ? 'Укажите название' : Array.from(t).length > 255 ? 'Название не должно превышать 255 символов' : null }, weight: (v, values) => { const selected = options.find((x) => x.id === values.residentId); const max = selected ? selected.available_redemption_weight + (editing && task?.resident.id === selected.id ? task.redemption_weight : 0) : 0; if (max === 0) return null; if (v === '' || !Number.isFinite(Number(v)) || Number(v) < 0 || Number(v) > max) return `Укажите вес от 0 до ${max}`; return null } } })
    const selected = options.find((x) => x.id === form.values.residentId)
    const maxWeight = selected ? selected.available_redemption_weight + (editing && task?.resident.id === selected.id ? task.redemption_weight : 0) : 0
    const searchOptions: ResidentSearchOption[] = useMemo(() => options.map((x) => ({ id: x.id, name: x.name })), [options])
    useEffect(() => { if (!opened) return; const initial: Values = task ? { residentId: task.resident.id, residentName: task.resident.name, title: task.title, areaId: task.area ? String(task.area.id) : null, weight: task.redemption_weight, deadline: task.deadline } : { ...empty, residentId: residentPreset?.id ?? '', residentName: residentPreset?.name ?? '' }; form.setValues(initial); form.resetDirty(initial); setSearch(initial.residentName); setError(null); void searchIndividualTaskResidents(initial.residentName).then(setOptions).catch(() => setOptions([])) }, [opened, task, residentPreset?.id, residentPreset?.name])
    useEffect(() => { if (!opened) return; let cancelled = false; setSearchLoading(true); void searchIndividualTaskResidents(debounced).then((x) => { if (!cancelled) setOptions(x) }).catch(() => { if (!cancelled) setOptions([]) }).finally(() => { if (!cancelled) setSearchLoading(false) }); return () => { cancelled = true } }, [debounced, opened])
    useEffect(() => { if (!selected) { setAreas([]); return }; void getIndividualTaskAreas(selected.dormitory_id).then((x) => { const next = x.map((a) => ({ value: String(a.id), label: a.floor === null ? a.name : `${a.floor} этаж. ${a.name}` })); setAreas(next); if (form.values.areaId && !next.some((a) => a.value === form.values.areaId)) form.setFieldValue('areaId', null) }).catch(() => setAreas([])) }, [selected?.dormitory_id])
    return <EntityFormModal opened={opened} onClose={onClose} title={<span className={modalClasses.title}>{editing ? 'Редактирование индивидуальной задачи' : 'Выдача индивидуальной задачи'}</span>} saving={saving} error={error} size={680} withCloseButton={false} modalClassNames={{ header: modalClasses.header, body: modalClasses.body, content: modalClasses.content }} contentGap={15} actionsClassName={modalClasses.actions} cancelButtonClassName={modalClasses.cancelButton} submitButtonClassName={[modalClasses.submitButton, modalClasses.accentButton].join(' ')} submitLabel={editing ? 'Редактировать' : 'Добавить'} onSubmit={form.onSubmit(async (v) => { setSaving(true); setError(null); try { const weight = maxWeight === 0 ? 0 : Number(v.weight); const request = { resident_id: v.residentId, title: v.title.trim(), area_id: v.areaId ? Number(v.areaId) : null, redemption_weight: weight, deadline: v.deadline, version: task?.version ?? 0 }; if (task) await updateIndividualTask(task.id, request); else await createIndividualTask(request); await onSaved(); onClose() } catch (e) { setError(errorText(e)); if (e instanceof ApiError && e.status === 409) void onSaved() } finally { setSaving(false) } })}>
        <ResidentSearchCombobox overlayOpened={opened} placeholder="Житель*" searchValue={form.values.residentName} selectedId={form.values.residentId || null} selectedLabel={form.values.residentName || null} hideDropdownWhenSelected options={searchOptions} loading={searchLoading} error={typeof form.errors.residentId === 'string' ? form.errors.residentId : null} inputClassName={modalClasses.input} onSearchChange={(v) => { setSearch(v); form.setFieldValue('residentName', v); if (!v.trim()) form.setFieldValue('residentId', '') }} onOptionSelect={(option) => { const changed = option.id !== form.values.residentId; form.setValues({ ...form.values, residentId: option.id, residentName: option.name, areaId: changed ? null : form.values.areaId, weight: changed ? '' : form.values.weight }); setSearch(option.name) }} />
        <TextInput placeholder="Название*" maxLength={255} classNames={{ input: modalClasses.input }} {...form.getInputProps('title')} />
        <Select placeholder="Территория" data={areas} value={form.values.areaId} onChange={(v) => form.setFieldValue('areaId', v)} classNames={{ input: modalClasses.input }} clearable />
        {maxWeight > 0 ? <NumberInput placeholder="Вес погашения*" hideControls decimalScale={1} step={0.1} allowNegative={false} min={0} max={maxWeight} clampBehavior="strict" value={form.values.weight} error={form.errors.weight} onChange={(v) => form.setFieldValue('weight', v === '' ? '' : v)} classNames={{ input: modalClasses.input }} /> : null}
        <DatePickerInput placeholder="Дедлайн" locale="ru" valueFormat="DD.MM.YYYY" minDate={dayjs().startOf('day').toDate()} value={form.values.deadline ? dayjs(form.values.deadline).toDate() : null} onChange={(v) => form.setFieldValue('deadline', v ? dayjs(v).format('YYYY-MM-DD') : null)} classNames={{ input: modalClasses.input }} clearable />
    </EntityFormModal>
}
