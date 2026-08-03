import { useState } from 'react'

import { PlusIcon } from '@phosphor-icons/react'
import { Alert, Button, Center, Loader } from '@mantine/core'

import { usePenalties } from '../../features/penalties/model/usePenalties'
import { CreatePenaltyDrawer } from '../../features/penalties/ui/CreatePenaltyDrawer'
import { PenaltiesTable } from '../../features/penalties/ui/PenaltiesTable'
import { ResidentPenaltiesDrawer } from '../../features/penalties/ui/ResidentPenaltiesDrawer'
import { PageFrame } from '../../shared/ui/PageFrame'

export function PenaltiesPage() {
    const { residents, loading, error, reload } = usePenalties()
    const [createDrawerOpened, setCreateDrawerOpened] = useState(false)
    const [selectedResidentId, setSelectedResidentId] = useState<string | null>(null)

    return (
        <PageFrame
            title="Предупреждения"
            titleActions={(
                <Button
                    radius="md"
                    h={42}
                    leftSection={<PlusIcon size={18} />}
                    onClick={() => setCreateDrawerOpened(true)}
                >
                    Выдать предупреждение
                </Button>
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
                <Alert color="gray">
                    Нет жителей с предупреждениями
                </Alert>
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
