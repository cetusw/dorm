import { useState } from 'react'

import { Alert, Center, Loader } from '@mantine/core'

import { navigateTo } from '../../app/navigation'
import { useSelectedDormitoryId } from '../../features/dormitories/model/useDormitorySelection'
import { useGroups } from '../../features/groups/model/useGroups'
import type { GroupListItem } from '../../features/groups/model/types'
import { DeleteGroupModal } from '../../features/groups/ui/DeleteGroupModal'
import { GroupFormModal } from '../../features/groups/ui/GroupFormModal'
import { GroupsTable } from '../../features/groups/ui/GroupsTable'
import { EmptyState } from '../../shared/ui/EmptyState'
import { ManagementPageFrame } from '../../shared/ui/ManagementPageFrame'
import { PageActionButton } from '../../shared/ui/PageActionButton'

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
        <PageActionButton onClick={handleCreate} disabled={selectedDormitoryId === null}>
            Создать группу
        </PageActionButton>
    )

    const content = !selectedDormitoryId ? (
            <EmptyState
                title="Общежитие не выбрано"
                description="Выберите общежитие в верхней панели, чтобы посмотреть список групп."
            />
        ) : loading ? (
            <Center py="xl">
                <Loader />
            </Center>
        ) : error ? (
            <Alert color="red" title="Ошибка">
                {error}
            </Alert>
        ) : groups.length === 0 ? (
            <EmptyState
                title="Группы не найдены"
                description="В выбранном общежитии пока нет групп."
            />
        ) : (
            <GroupsTable
                groups={groups}
                onOpen={(groupId) => navigateTo(`/app/groups/${groupId}/teams`)}
                onEdit={handleEdit}
                onDelete={handleDelete}
            />
        )

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
