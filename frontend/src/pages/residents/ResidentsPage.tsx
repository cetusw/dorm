import { useState } from 'react'

import { Alert, Button, Center, Loader } from '@mantine/core'

import { useSelectedDormitoryId } from '../../features/dormitories/model/useDormitorySelection'
import { useResidents } from '../../features/residents/model/useResidents'
import type { ResidentListItem } from '../../features/residents/model/types'
import { DeleteResidentModal } from '../../features/residents/ui/DeleteResidentModal'
import { ResidentFormModal } from '../../features/residents/ui/ResidentFormModal'
import { ResidentsTable } from '../../features/residents/ui/ResidentsTable'
import { ManagementPageFrame } from '../../shared/ui/ManagementPageFrame'

export function ResidentsPage() {
    const selectedDormitoryId = useSelectedDormitoryId()
    const { residents, loading, error, reload } = useResidents(selectedDormitoryId)
    const [formOpened, setFormOpened] = useState(false)
    const [editingResidentId, setEditingResidentId] = useState<string | null>(null)
    const [deletingResident, setDeletingResident] = useState<ResidentListItem | null>(null)

    function handleCreate() {
        setEditingResidentId(null)
        setFormOpened(true)
    }

    function handleEdit(residentId: string) {
        setEditingResidentId(residentId)
        setFormOpened(true)
    }

    function handleDelete(resident: ResidentListItem) {
        setDeletingResident(resident)
    }

    const titleActions = (
        <Button radius="md" onClick={handleCreate} disabled={selectedDormitoryId === null}>
            + Создать жителя
        </Button>
    )

    let content = null

    if (!selectedDormitoryId) {
        content = (
            <Alert color="gray">
                Выберите общежитие в верхней панели.
            </Alert>
        )
    } else if (loading) {
        content = (
            <Center py="xl">
                <Loader />
            </Center>
        )
    } else if (error) {
        content = (
            <Alert color="red" title="Ошибка">
                {error}
            </Alert>
        )
    } else if (residents.length === 0) {
        content = (
            <Alert color="gray">
                В выбранном общежитии пока нет жителей.
            </Alert>
        )
    } else {
        content = (
            <ResidentsTable
                residents={residents}
                onEdit={handleEdit}
                onDelete={handleDelete}
            />
        )
    }

    return (
        <ManagementPageFrame title="Жители" titleActions={titleActions}>
            {content}

            {selectedDormitoryId !== null && (
                <>
                    <ResidentFormModal
                        opened={formOpened}
                        mode={editingResidentId === null ? 'create' : 'edit'}
                        residentId={editingResidentId}
                        dormitoryId={selectedDormitoryId}
                        onClose={() => {
                            setFormOpened(false)
                            setEditingResidentId(null)
                        }}
                        onSaved={reload}
                    />

                    <DeleteResidentModal
                        opened={deletingResident !== null}
                        resident={deletingResident}
                        onClose={() => setDeletingResident(null)}
                        onDeleted={reload}
                    />
                </>
            )}
        </ManagementPageFrame>
    )
}
