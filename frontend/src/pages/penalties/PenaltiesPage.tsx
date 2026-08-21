import { useState } from 'react'

import { Alert, Center, Loader, Menu } from '@mantine/core'
import { CheckCircleIcon, WarningIcon } from '@phosphor-icons/react'

import { usePenalties } from '../../features/penalties/model/usePenalties'
import type { PenaltyEntryType, PenaltyResidentSummary } from '../../features/penalties/model/types'
import { CreatePenaltyDrawer } from '../../features/penalties/ui/CreatePenaltyDrawer'
import { PenaltiesTable } from '../../features/penalties/ui/PenaltiesTable'
import { ResidentPenaltiesDrawer } from '../../features/penalties/ui/ResidentPenaltiesDrawer'
import { IndividualTaskFormModal } from '../../features/individual-tasks/ui/IndividualTaskFormModal'
import { ResidentIndividualTasksDrawer } from '../../features/individual-tasks/ui/ResidentIndividualTasksDrawer'
import { EmptyState } from '../../shared/ui/EmptyState'
import { PageActionButton } from '../../shared/ui/PageActionButton'
import { PageFrame } from '../../shared/ui/PageFrame'

export function PenaltiesPage() {
    const { residents, loading, error, synchronizeResidents } = usePenalties()
    const [createDrawerOpened, setCreateDrawerOpened] = useState(false)
    const [selectedResidentId, setSelectedResidentId] = useState<string | null>(null)
    const [individualTaskResident, setIndividualTaskResident] = useState<PenaltyResidentSummary | null>(null)
    const [individualTaskFormResident, setIndividualTaskFormResident] = useState<{ id: string; name: string } | null | undefined>(undefined)
    const [individualTasksRefreshToken, setIndividualTasksRefreshToken] = useState(0)
    const [residentAction, setResidentAction] = useState<{
        resident: PenaltyResidentSummary
        entryType: PenaltyEntryType
    } | null>(null)

    return (
        <PageFrame
            title="Предупреждения"
            titleActions={(
                <Menu position="bottom-end" withinPortal>
                    <Menu.Target><PageActionButton>Выдать</PageActionButton></Menu.Target>
                    <Menu.Dropdown>
                        <Menu.Item leftSection={<WarningIcon size={24} />} onClick={() => setCreateDrawerOpened(true)}>Выдать предупреждение</Menu.Item>
                        <Menu.Item leftSection={<CheckCircleIcon size={24} />} onClick={() => setIndividualTaskFormResident(null)}>Выдать индивидуальную задачу</Menu.Item>
                    </Menu.Dropdown>
                </Menu>
            )}
        >
            {loading ? (
                <Center py="xl">
                    <Loader />
                </Center>
            ) : error ? (
                <Alert color="red" title="Ошибка">
                    {error}
                </Alert>
            ) : residents.length === 0 ? (
                <EmptyState
                    title="Предупреждения не найдены"
                    description="Сейчас нет жителей с предупреждениями."
                />
            ) : (
                <PenaltiesTable
                    residents={residents}
                    onOpenResident={setSelectedResidentId}
                    onIssuePenalty={(resident) => setResidentAction({ resident, entryType: 'issue' })}
                    onResolvePenalty={(resident) => setResidentAction({ resident, entryType: 'resolve' })}
                    onIssueIndividualTask={(resident) => setIndividualTaskFormResident({ id: resident.user_id, name: resident.full_name })}
                    onOpenIndividualTasks={setIndividualTaskResident}
                />
            )}

            <CreatePenaltyDrawer
                opened={createDrawerOpened || residentAction !== null}
                onClose={() => {
                    setCreateDrawerOpened(false)
                    setResidentAction(null)
                }}
                onCreated={synchronizeResidents}
                residentPreset={residentAction ? {
                    id: residentAction.resident.user_id,
                    name: residentAction.resident.full_name,
                } : null}
                mode="create"
                entryType={residentAction?.entryType ?? 'issue'}
                maxWeight={residentAction?.entryType === 'resolve'
                    ? residentAction.resident.total_weight
                    : undefined}
            />

            <ResidentPenaltiesDrawer
                residentId={selectedResidentId}
                opened={selectedResidentId !== null}
                onClose={() => setSelectedResidentId(null)}
                onChanged={synchronizeResidents}
            />
            <IndividualTaskFormModal
                opened={individualTaskFormResident !== undefined}
                residentPreset={individualTaskFormResident ?? null}
                onClose={() => setIndividualTaskFormResident(undefined)}
                onSaved={async () => { await synchronizeResidents(); setIndividualTasksRefreshToken((value) => value + 1) }}
            />
            <ResidentIndividualTasksDrawer
                resident={individualTaskResident ? { id: individualTaskResident.user_id, name: individualTaskResident.full_name } : null}
                opened={individualTaskResident !== null}
                refreshToken={individualTasksRefreshToken}
                onClose={() => setIndividualTaskResident(null)}
                onChanged={synchronizeResidents}
            />
        </PageFrame>
    )
}
