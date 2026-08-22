import { useCallback, useEffect, useState } from 'react'

import { RowsPlusBottomIcon } from '@phosphor-icons/react'
import { Alert, Box, Center, Drawer, FocusTrap, Loader, Stack } from '@mantine/core'
import { useMediaQuery } from '@mantine/hooks'

import { ApiError } from '../../../shared/api/ApiError'
import { EmptyState } from '../../../shared/ui/EmptyState'
import { EntityDeleteModal } from '../../../shared/ui/EntityDeleteModal'
import { SettingsAddAction } from '../../../shared/ui/SettingsAddAction'
import { deleteIndividualTask, getResidentIndividualTasks, rejectIndividualTask, verifyIndividualTask } from '../api/individualTasksApi'
import type { IndividualTask } from '../model/types'
import { IndividualTaskCard } from './IndividualTaskCard'
import { IndividualTaskFormModal } from './IndividualTaskFormModal'

type Props = { resident: { id: string; name: string } | null; opened: boolean; refreshToken?: number; onClose: () => void; onChanged: () => Promise<void> | void }
function errorMessage(error: unknown): string {
    return error instanceof ApiError ? error.message : 'Не удалось загрузить индивидуальные задачи'
}

export function ResidentIndividualTasksDrawer({ resident, opened, refreshToken = 0, onClose, onChanged }: Props) {
    const mobile = useMediaQuery('(max-width: 48em)')
    const [tasks, setTasks] = useState<IndividualTask[] | null>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [editing, setEditing] = useState<IndividualTask | null>(null)
    const [deleting, setDeleting] = useState<IndividualTask | null>(null)
    const [pending, setPending] = useState<string | null>(null)
    const load = useCallback(async (clearError = true) => {
        if (!resident) return

        setLoading(true)
        if (clearError) setError(null)

        try {
            const response = await getResidentIndividualTasks(resident.id)
            setTasks(response.filter((task) => task.status !== 'verified'))
        } catch (currentError) {
            setError(errorMessage(currentError))
            setTasks(null)
        } finally {
            setLoading(false)
        }
    }, [resident])
    useEffect(() => { if (opened) void load() }, [opened, load, refreshToken])
    async function action(id: string, request: (taskID: string) => Promise<IndividualTask>) {
        setPending(id)
        try {
            await request(id)
            await load()
            await onChanged()
        } catch (currentError) {
            setError(errorMessage(currentError))
            void load(false)
        } finally {
            setPending(null)
        }
    }
    return <Drawer opened={opened} onClose={onClose} position={mobile ? 'bottom' : 'right'} size={mobile ? '90%' : 760} title={resident?.name ?? 'Индивидуальные задачи'} closeOnEscape={editing === null && deleting === null}><FocusTrap.InitialFocus /><Stack gap={0}>{error ? <Alert color="red" mb="md">{error}</Alert> : null}{loading ? <Center py="xl"><Loader /></Center> : tasks?.length === 0 ? <EmptyState title="Нет индивидуальных задач" description="У жителя нет активных индивидуальных задач" /> : tasks?.map((task) => <IndividualTaskCard key={task.id} task={task} pending={pending === task.id} onEdit={() => setEditing(task)} onDelete={() => setDeleting(task)} onVerify={() => void action(task.id, verifyIndividualTask)} onReject={() => void action(task.id, rejectIndividualTask)} />)}{!loading && resident ? <Box mt={20}><SettingsAddAction icon={<RowsPlusBottomIcon size={24} />} onClick={() => setEditing({} as IndividualTask)}>Выдать индивидуальную задачу</SettingsAddAction></Box> : null}</Stack><IndividualTaskFormModal opened={editing !== null} task={editing?.id ? editing : null} residentPreset={resident} onClose={() => setEditing(null)} onSaved={async () => { await load(); await onChanged() }} /><EntityDeleteModal opened={deleting !== null} title="Удаление задачи" entityLabel="задачу" entityName={deleting?.title ?? null} submitLabel="Удалить" onClose={() => setDeleting(null)} onConfirm={async () => { if (deleting) { await deleteIndividualTask(deleting.id); await load(); await onChanged() } }} errorMessage="Не удалось удалить задачу" /></Drawer>
}
