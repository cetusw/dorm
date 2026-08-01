import { useEffect, useState } from 'react'

import { Alert, Button, Center, Loader } from '@mantine/core'

import { getGroup } from '../../features/groups/api/groupsApi'
import { useSelectedDormitoryId } from '../../features/dormitories/model/useDormitorySelection'
import { useTeams } from '../../features/teams/model/useTeams'
import type { TeamListItem } from '../../features/teams/model/types'
import { DeleteTeamModal } from '../../features/teams/ui/DeleteTeamModal'
import { TeamFormModal } from '../../features/teams/ui/TeamFormModal'
import { TeamsTable } from '../../features/teams/ui/TeamsTable'
import { ManagementPageFrame } from '../../shared/ui/ManagementPageFrame'

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
        <Button radius="md" onClick={handleCreate} disabled={!selectedDormitoryId}>
            + Создать команду
        </Button>
    )

    let content = null

    if (!selectedDormitoryId) {
        content = (
            <Alert color="gray">
                Выберите общежитие в верхней панели.
            </Alert>
        )
    } else if (groupLoading) {
        content = (
            <Center py="xl">
                <Loader />
            </Center>
        )
    } else if (groupError) {
        content = (
            <Alert color="red" title="Ошибка">
                {groupError}
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
    } else if (teams.length === 0) {
        content = (
            <>
                <Alert color="gray">
                    В выбранной группе пока нет команд.
                </Alert>
            </>
        )
    } else {
        content = (
            <>
                <TeamsTable
                    teams={teams}
                    onEdit={handleEdit}
                    onDelete={handleDelete}
                />
            </>
        )
    }

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
