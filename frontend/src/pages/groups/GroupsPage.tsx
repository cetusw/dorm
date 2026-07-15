import { useState } from 'react'

import { Alert, Button, Center, Loader } from '@mantine/core'

import { useSelectedDormitoryId } from '../../features/dormitories/model/useDormitorySelection'
import { useGroups } from '../../features/groups/model/useGroups'
import type { GroupListItem } from '../../features/groups/model/types'
import { DeleteGroupModal } from '../../features/groups/ui/DeleteGroupModal'
import { GroupFormModal } from '../../features/groups/ui/GroupFormModal'
import { GroupsTable } from '../../features/groups/ui/GroupsTable'
import { ManagementPageFrame } from '../../shared/ui/ManagementPageFrame'

export function GroupsPage() {
    const selectedDormitoryId = useSelectedDormitoryId()
    const { groups, loading, error, reload } = useGroups(selectedDormitoryId)
    const [formOpened, setFormOpened] = useState(false)
    const [editingGroupId, setEditingGroupId] = useState<string | null>(null)
    const [deletingGroup, setDeletingGroup] = useState<GroupListItem | null>(null)

    function handleCreate() {
        setEditingGroupId(null)
        setFormOpened(true)
    }

    function handleEdit(groupId: string) {
        setEditingGroupId(groupId)
        setFormOpened(true)
    }

    function handleDelete(group: GroupListItem) {
        setDeletingGroup(group)
    }

    const titleActions = (
        <Button radius="md" onClick={handleCreate} disabled={selectedDormitoryId === null}>
            + Создать группу
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
    } else if (groups.length === 0) {
        content = (
            <Alert color="gray">
                В выбранном общежитии пока нет групп.
            </Alert>
        )
    } else {
        content = (
            <GroupsTable
                groups={groups}
                onEdit={handleEdit}
                onDelete={handleDelete}
            />
        )
    }

    return (
        <ManagementPageFrame title="Группы" titleActions={titleActions}>
            {content}

            {selectedDormitoryId !== null && (
                <>
                    <GroupFormModal
                        opened={formOpened}
                        mode={editingGroupId === null ? 'create' : 'edit'}
                        groupId={editingGroupId}
                        dormitoryId={selectedDormitoryId}
                        onClose={() => {
                            setFormOpened(false)
                            setEditingGroupId(null)
                        }}
                        onSaved={reload}
                    />

                    <DeleteGroupModal
                        opened={deletingGroup !== null}
                        group={deletingGroup}
                        onClose={() => setDeletingGroup(null)}
                        onDeleted={reload}
                    />
                </>
            )}
        </ManagementPageFrame>
    )
}
