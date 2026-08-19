import { useState } from 'react'

import { Alert, Center, Loader } from '@mantine/core'

import { useSelectedDormitoryId } from '../../features/dormitories/model/useDormitorySelection'
import { useAreas } from '../../features/areas/model/useAreas'
import type { AreaListItem } from '../../features/areas/model/types'
import { AreasTable } from '../../features/areas/ui/AreasTable'
import { AreaFormModal } from '../../features/areas/ui/AreaFormModal'
import { DeleteAreaModal } from '../../features/areas/ui/DeleteAreaModal'
import { EmptyState } from '../../shared/ui/EmptyState'
import { ManagementPageFrame } from '../../shared/ui/ManagementPageFrame'
import { PageActionButton } from '../../shared/ui/PageActionButton'

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
        <PageActionButton onClick={handleCreate} disabled={!selectedDormitoryId}>
            Создать территорию
        </PageActionButton>
    )

    const content = !selectedDormitoryId ? (
            <EmptyState
                title="Общежитие не выбрано"
                description="Выберите общежитие в верхней панели, чтобы посмотреть список территорий."
            />
        ) : loading ? (
            <Center py="xl">
                <Loader />
            </Center>
        ) : error ? (
            <Alert color="red" title="Ошибка">
                {error}
            </Alert>
        ) : areas.length === 0 ? (
            <EmptyState
                title="Территории не найдены"
                description="В выбранном общежитии пока нет территорий."
            />
        ) : (
            <AreasTable
                areas={areas}
                onEdit={handleEdit}
                onDelete={handleDelete}
            />
        )

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
