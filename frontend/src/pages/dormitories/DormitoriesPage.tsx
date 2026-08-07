import { useState } from 'react'

import { Alert, Center, Loader } from '@mantine/core'

import { useDormitories } from '../../features/dormitories/model/useDormitories'
import type { DormitoryListItem } from '../../features/dormitories/model/types'
import { notifyDormitoriesChanged } from '../../features/dormitories/model/useDormitorySelection'
import { DeleteDormitoryModal } from '../../features/dormitories/ui/DeleteDormitoryModal'
import { DormitoryFormModal } from '../../features/dormitories/ui/DormitoryFormModal'
import { DormitoriesTable } from '../../features/dormitories/ui/DormitoriesTable'
import { EmptyState } from '../../shared/ui/EmptyState'
import { ManagementPageFrame } from '../../shared/ui/ManagementPageFrame'
import { PageActionButton } from '../../shared/ui/PageActionButton'

export function DormitoriesPage() {
    const { dormitories, loading, error, reload } = useDormitories()
    const [formOpened, setFormOpened] = useState(false)
    const [editingDormitoryId, setEditingDormitoryId] = useState<number | null>(null)
    const [deletingDormitory, setDeletingDormitory] = useState<DormitoryListItem | null>(null)

    function handleCreate() {
        setEditingDormitoryId(null)
        setFormOpened(true)
    }

    function handleEdit(dormitoryId: number) {
        setEditingDormitoryId(dormitoryId)
        setFormOpened(true)
    }

    function handleDelete(dormitory: DormitoryListItem) {
        setDeletingDormitory(dormitory)
    }

    async function handleDormitoriesChanged() {
        await reload()
        notifyDormitoriesChanged()
    }

    const titleActions = (
        <PageActionButton onClick={handleCreate}>Создать общежитие</PageActionButton>
    )

    let content = null

    if (loading) {
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
    } else if (dormitories.length === 0) {
        content = (
            <EmptyState
                title="Общежития не найдены"
                description="Общежития пока не добавлены."
            />
        )
    } else {
        content = (
            <DormitoriesTable
                dormitories={dormitories}
                onEdit={handleEdit}
                onDelete={handleDelete}
            />
        )
    }

    return (
        <ManagementPageFrame title="Общежития" titleActions={titleActions}>
            {content}

            <DormitoryFormModal
                opened={formOpened}
                mode={editingDormitoryId === null ? 'create' : 'edit'}
                dormitoryId={editingDormitoryId}
                onClose={() => {
                    setFormOpened(false)
                    setEditingDormitoryId(null)
                }}
                onSaved={handleDormitoriesChanged}
            />

            <DeleteDormitoryModal
                opened={deletingDormitory !== null}
                dormitory={deletingDormitory}
                onClose={() => setDeletingDormitory(null)}
                onDeleted={handleDormitoriesChanged}
            />
        </ManagementPageFrame>
    )
}
