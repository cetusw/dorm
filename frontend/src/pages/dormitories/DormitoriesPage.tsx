import { useState } from 'react'

import { Alert, Button, Center, Loader } from '@mantine/core'

import { useDormitories } from '../../features/dormitories/model/useDormitories'
import type { DormitoryListItem } from '../../features/dormitories/model/types'
import { DeleteDormitoryModal } from '../../features/dormitories/ui/DeleteDormitoryModal'
import { DormitoryFormModal } from '../../features/dormitories/ui/DormitoryFormModal'
import { DormitoriesTable } from '../../features/dormitories/ui/DormitoriesTable'
import { PageFrame } from '../../shared/ui/PageFrame'

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

    const titleActions = (
        <Button radius="md" onClick={handleCreate}>
            + Создать общежитие
        </Button>
    )

    if (loading) {
        return (
            <PageFrame title="Общежития" titleActions={titleActions}>
                <Center py="xl">
                    <Loader />
                </Center>
            </PageFrame>
        )
    }

    if (error) {
        return (
            <PageFrame title="Общежития" titleActions={titleActions}>
                <Alert color="red" title="Ошибка">
                    {error}
                </Alert>
            </PageFrame>
        )
    }

    if (dormitories.length === 0) {
        return (
            <PageFrame title="Общежития" titleActions={titleActions}>
                <Alert color="gray">
                    Общежития пока не добавлены.
                </Alert>
            </PageFrame>
        )
    }

    return (
        <PageFrame title="Общежития" titleActions={titleActions}>
            <DormitoriesTable
                dormitories={dormitories}
                onEdit={handleEdit}
                onDelete={handleDelete}
            />

            <DormitoryFormModal
                opened={formOpened}
                mode={editingDormitoryId === null ? 'create' : 'edit'}
                dormitoryId={editingDormitoryId}
                onClose={() => {
                    setFormOpened(false)
                    setEditingDormitoryId(null)
                }}
                onSaved={reload}
            />

            <DeleteDormitoryModal
                opened={deletingDormitory !== null}
                dormitory={deletingDormitory}
                onClose={() => setDeletingDormitory(null)}
                onDeleted={reload}
            />
        </PageFrame>
    )
}
