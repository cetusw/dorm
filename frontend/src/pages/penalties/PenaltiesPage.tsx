import { useState } from 'react'

import { Alert, Center, Loader } from '@mantine/core'

import { usePenalties } from '../../features/penalties/model/usePenalties'
import type { PenaltyEntryType, PenaltyResidentSummary } from '../../features/penalties/model/types'
import { CreatePenaltyDrawer } from '../../features/penalties/ui/CreatePenaltyDrawer'
import { PenaltiesTable } from '../../features/penalties/ui/PenaltiesTable'
import { ResidentPenaltiesDrawer } from '../../features/penalties/ui/ResidentPenaltiesDrawer'
import { EmptyState } from '../../shared/ui/EmptyState'
import { PageActionButton } from '../../shared/ui/PageActionButton'
import { PageFrame } from '../../shared/ui/PageFrame'

export function PenaltiesPage() {
    const { residents, loading, error, synchronizeResidents } = usePenalties()
    const [createDrawerOpened, setCreateDrawerOpened] = useState(false)
    const [selectedResidentId, setSelectedResidentId] = useState<string | null>(null)
    const [residentAction, setResidentAction] = useState<{
        resident: PenaltyResidentSummary
        entryType: PenaltyEntryType
    } | null>(null)

    return (
        <PageFrame
            title="Предупреждения"
            titleActions={(
                <PageActionButton onClick={() => setCreateDrawerOpened(true)}>
                    Выдать предупреждение
                </PageActionButton>
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
        </PageFrame>
    )
}
