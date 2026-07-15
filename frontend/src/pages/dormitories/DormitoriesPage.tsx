import type { ReactNode } from 'react'
import { useState } from 'react'

import { Alert, Box, Button, Center, Group, Loader, Stack, Title } from '@mantine/core'

import { useDormitories } from '../../features/dormitories/model/useDormitories'
import type { DormitoryListItem } from '../../features/dormitories/model/types'
import { DeleteDormitoryModal } from '../../features/dormitories/ui/DeleteDormitoryModal'
import { DormitoryFormModal } from '../../features/dormitories/ui/DormitoryFormModal'
import { DormitoriesTable } from '../../features/dormitories/ui/DormitoriesTable'

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
            <DormitoriesPageLayout titleActions={titleActions}>
                <Center py="xl">
                    <Loader />
                </Center>
            </DormitoriesPageLayout>
        )
    }

    if (error) {
        return (
            <DormitoriesPageLayout titleActions={titleActions}>
                <Alert color="red" title="Ошибка">
                    {error}
                </Alert>
            </DormitoriesPageLayout>
        )
    }

    if (dormitories.length === 0) {
        return (
            <DormitoriesPageLayout titleActions={titleActions}>
                <Alert color="gray">
                    Общежития пока не добавлены.
                </Alert>
            </DormitoriesPageLayout>
        )
    }

    return (
        <DormitoriesPageLayout titleActions={titleActions}>
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
        </DormitoriesPageLayout>
    )
}

function DormitoriesPageLayout({
    children,
    titleActions,
}: {
    children: ReactNode
    titleActions: ReactNode
}) {
    return (
        <Box px={{ base: 'md', md: 'xl' }} py="xl">
            <Stack gap="lg" maw={1240} mx="auto">
                <Group justify="space-between" align="center" wrap="wrap" gap="md">
                    <Title order={1}>Общежития</Title>
                    {titleActions}
                </Group>

                {children}
            </Stack>
        </Box>
    )
}
