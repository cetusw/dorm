import { useState } from 'react'

import { Alert, Center, Loader } from '@mantine/core'

import { usePenalties } from '../../features/penalties/model/usePenalties'
import { CreatePenaltyDrawer } from '../../features/penalties/ui/CreatePenaltyDrawer'
import { PenaltiesTable } from '../../features/penalties/ui/PenaltiesTable'
import { ResidentPenaltiesDrawer } from '../../features/penalties/ui/ResidentPenaltiesDrawer'
import { EmptyState } from '../../shared/ui/EmptyState'
import { PageActionButton } from '../../shared/ui/PageActionButton'
import { PageFrame } from '../../shared/ui/PageFrame'

export function PenaltiesPage() {
    const { residents, loading, error, reload } = usePenalties()
    const [createDrawerOpened, setCreateDrawerOpened] = useState(false)
    const [selectedResidentId, setSelectedResidentId] = useState<string | null>(null)

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
                />
            )}

            <CreatePenaltyDrawer
                opened={createDrawerOpened}
                onClose={() => setCreateDrawerOpened(false)}
                onCreated={reload}
            />

            <ResidentPenaltiesDrawer
                residentId={selectedResidentId}
                opened={selectedResidentId !== null}
                onClose={() => setSelectedResidentId(null)}
                onChanged={reload}
            />
        </PageFrame>
    )
}
