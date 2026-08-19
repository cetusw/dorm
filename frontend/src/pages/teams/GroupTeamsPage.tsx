import { useEffect, useState } from 'react'

import { Alert, Center, Loader } from '@mantine/core'

import { getGroup } from '../../features/groups/api/groupsApi'
import { useSelectedDormitoryId } from '../../features/dormitories/model/useDormitorySelection'
import { useTeams } from '../../features/teams/model/useTeams'
import type { TeamListItem } from '../../features/teams/model/types'
import { DeleteTeamModal } from '../../features/teams/ui/DeleteTeamModal'
import { TeamFormModal } from '../../features/teams/ui/TeamFormModal'
import { TeamsTable } from '../../features/teams/ui/TeamsTable'
import { EmptyState } from '../../shared/ui/EmptyState'
import { ManagementPageFrame } from '../../shared/ui/ManagementPageFrame'
import { PageActionButton } from '../../shared/ui/PageActionButton'

type Props = {
    groupId: string
}

export function GroupTeamsPage({ groupId }: Props) {
    const selectedDormitoryId = useSelectedDormitoryId()
    const { teams, loading, error, reload } = useTeams(groupId)
    const [groupName, setGroupName] = useState<string>('')
    const [groupLoading, setGroupLoading] = useState(true)
    const [groupError, setGroupError] = useState<string | null>(null)
    const [formOpened, setFormOpened] = useState(false)
    const [editingTeamId, setEditingTeamId] = useState<string | null>(null)
    const [deletingTeam, setDeletingTeam] = useState<TeamListItem | null>(null)

    useEffect(() => {
        let active = true

        async function loadGroup() {
            setGroupLoading(true)
            setGroupError(null)

            try {
                const group = await getGroup(groupId)
                if (!active) {
                    return
                }

                setGroupName(group.name)
            } catch (error) {
                if (!active) {
                    return
                }

                setGroupError(error instanceof Error ? error.message : 'Не удалось загрузить группу')
            } finally {
                if (active) {
                    setGroupLoading(false)
                }
            }
        }

        void loadGroup()

        return () => {
            active = false
        }
    }, [groupId])

    function handleCreate() {
        setEditingTeamId(null)
        setFormOpened(true)
    }

    function handleEdit(teamId: string) {
        setEditingTeamId(teamId)
        setFormOpened(true)
    }

    function handleDelete(team: TeamListItem) {
        setDeletingTeam(team)
    }

    const titleActions = (
        <PageActionButton onClick={handleCreate} disabled={!selectedDormitoryId}>
            Создать команду
        </PageActionButton>
    )

    const content = !selectedDormitoryId ? (
            <EmptyState
                title="Общежитие не выбрано"
                description="Выберите общежитие в верхней панели, чтобы посмотреть список команд."
            />
        ) : groupLoading ? (
            <Center py="xl">
                <Loader />
            </Center>
        ) : groupError ? (
            <Alert color="red" title="Ошибка">
                {groupError}
            </Alert>
        ) : loading ? (
            <Center py="xl">
                <Loader />
            </Center>
        ) : error ? (
            <Alert color="red" title="Ошибка">
                {error}
            </Alert>
        ) : teams.length === 0 ? (
            <EmptyState
                title="Команды не найдены"
                description="В выбранной группе пока нет команд."
            />
        ) : (
            <>
                <TeamsTable
                    teams={teams}
                    onEdit={handleEdit}
                    onDelete={handleDelete}
                />
            </>
        )

    return (
        <ManagementPageFrame title={`Команды группы ${groupName}`} titleActions={titleActions}>

            {content}

            {selectedDormitoryId && (
                <>
                    <TeamFormModal
                        opened={formOpened}
                        mode={editingTeamId === null ? 'create' : 'edit'}
                        teamId={editingTeamId}
                        groupId={groupId}
                        onClose={() => {
                            setFormOpened(false)
                            setEditingTeamId(null)
                        }}
                        onSaved={reload}
                    />

                    <DeleteTeamModal
                        opened={deletingTeam !== null}
                        team={deletingTeam}
                        onClose={() => setDeletingTeam(null)}
                        onDeleted={reload}
                    />
                </>
            )}
        </ManagementPageFrame>
    )
}
