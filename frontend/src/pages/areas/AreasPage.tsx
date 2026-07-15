import { useState } from 'react'

import { Alert, Button, Center, Loader } from '@mantine/core'

import { useSelectedDormitoryId } from '../../features/dormitories/model/useDormitorySelection'
import { useAreas } from '../../features/areas/model/useAreas'
import type { AreaListItem } from '../../features/areas/model/types'
import { AreasTable } from '../../features/areas/ui/AreasTable'
import { AreaFormModal } from '../../features/areas/ui/AreaFormModal'
import { DeleteAreaModal } from '../../features/areas/ui/DeleteAreaModal'
import { ManagementPageFrame } from '../../shared/ui/ManagementPageFrame'

export function AreasPage() {
    const selectedDormitoryId = useSelectedDormitoryId()
    const { areas, loading, error, reload } = useAreas(selectedDormitoryId)
    const [formOpened, setFormOpened] = useState(false)
    const [editingAreaId, setEditingAreaId] = useState<number | null>(null)
    const [deletingArea, setDeletingArea] = useState<AreaListItem | null>(null)

    function handleCreate() {
        setEditingAreaId(null)
        setFormOpened(true)
    }

    function handleEdit(areaId: number) {
        setEditingAreaId(areaId)
        setFormOpened(true)
    }

    function handleDelete(area: AreaListItem) {
        setDeletingArea(area)
    }

    const titleActions = (
        <Button radius="md" onClick={handleCreate} disabled={!selectedDormitoryId}>
            + Создать территорию
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
    } else if (areas.length === 0) {
        content = (
            <Alert color="gray">
                В выбранном общежитии пока нет территорий.
            </Alert>
        )
    } else {
        content = (
            <AreasTable
                areas={areas}
                onEdit={handleEdit}
                onDelete={handleDelete}
            />
        )
    }

    return (
        <ManagementPageFrame title="Территории" titleActions={titleActions}>
            {content}

            {selectedDormitoryId && (
                <>
                    <AreaFormModal
                        opened={formOpened}
                        mode={editingAreaId === null ? 'create' : 'edit'}
                        areaId={editingAreaId}
                        dormitoryId={selectedDormitoryId}
                        onClose={() => {
                            setFormOpened(false)
                            setEditingAreaId(null)
                        }}
                        onSaved={reload}
                    />

                    <DeleteAreaModal
                        opened={deletingArea !== null}
                        area={deletingArea}
                        dormitoryId={selectedDormitoryId}
                        onClose={() => setDeletingArea(null)}
                        onDeleted={reload}
                    />
                </>
            )}
        </ManagementPageFrame>
    )
}
