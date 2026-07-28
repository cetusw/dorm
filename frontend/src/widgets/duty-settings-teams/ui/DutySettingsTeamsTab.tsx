import { useEffect, useMemo, useState } from 'react'

import { closestCenter, DndContext, KeyboardSensor, PointerSensor, useSensor, useSensors, type DragEndEvent } from '@dnd-kit/core'
import { arrayMove, SortableContext, sortableKeyboardCoordinates, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { RowsPlusBottomIcon } from '@phosphor-icons/react'
import { Alert, Stack } from '@mantine/core'

import {
    assignDutySettingsActiveTeam,
    createDutySettingsTeam,
    deleteDutySettingsTeam,
    getDutySettingsTeam,
    getDutySettingsTeamMemberOptions,
    reorderDutySettingsTeams,
    updateDutySettingsTeam,
} from '../../../features/duty-settings/api/dutySettingsApi'
import type { DutySettingsTeam } from '../../../features/duty-settings/model/types'
import type { TeamListItem } from '../../../features/teams/model/types'
import { ConfirmActionModal } from '../../../shared/ui/ConfirmActionModal'
import { DeleteTeamModal } from '../../../features/teams/ui/DeleteTeamModal'
import { TeamFormModal } from '../../../features/teams/ui/TeamFormModal'
import { SettingsAddAction } from '../../../shared/ui/SettingsAddAction'
import { DutySettingsTeamCard } from './DutySettingsTeamCard'
import { DutySettingsTeamMembersDrawer } from './DutySettingsTeamMembersDrawer'

type Props = {
    groupId: string
    teams: DutySettingsTeam[]
    activeDutyTeamId: string | null
    onReload: () => Promise<void>
}

function sortTeams(teams: DutySettingsTeam[]): DutySettingsTeam[] {
    return [...teams].sort((left, right) => left.rotation_position - right.rotation_position)
}

function renumberTeams(teams: DutySettingsTeam[]): DutySettingsTeam[] {
    return teams.map((team, index) => ({
        ...team,
        rotation_position: index + 1,
    }))
}

function toTeamListItem(team: DutySettingsTeam): TeamListItem {
    return {
        id: team.id,
        name: team.name,
        leader: team.leader,
        members_count: team.members_count,
        rotation_position: team.rotation_position,
	}
}

function getDutyTeamWarningName(team: DutySettingsTeam | null): string {
	if (!team) {
		return ""
	}

	return team.leader?.name ?? team.name
}

export function DutySettingsTeamsTab({ groupId, teams, activeDutyTeamId, onReload }: Props) {
    const [orderedTeams, setOrderedTeams] = useState<DutySettingsTeam[]>(() => sortTeams(teams))
    const [savingOrder, setSavingOrder] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [formOpened, setFormOpened] = useState(false)
    const [deletingTeam, setDeletingTeam] = useState<DutySettingsTeam | null>(null)
    const [assigningDutyTeam, setAssigningDutyTeam] = useState<DutySettingsTeam | null>(null)
    const [selectedTeam, setSelectedTeam] = useState<DutySettingsTeam | null>(null)

    useEffect(() => {
        setOrderedTeams(sortTeams(teams))
    }, [teams])

    useEffect(() => {
        if (!selectedTeam) {
            return
        }

        const nextSelectedTeam = teams.find((team) => team.id === selectedTeam.id) ?? null
        setSelectedTeam(nextSelectedTeam)
    }, [selectedTeam, teams])

    const sensors = useSensors(
        useSensor(PointerSensor, {
            activationConstraint: {
                distance: 6,
            },
        }),
        useSensor(KeyboardSensor, {
            coordinateGetter: sortableKeyboardCoordinates,
        }),
    )

    const orderedIds = useMemo(() => orderedTeams.map((team) => team.id), [orderedTeams])

    function handleCreate() {
        setFormOpened(true)
    }

    async function handleDragEnd(event: DragEndEvent) {
        const { active, over } = event
        if (!over || active.id === over.id || savingOrder) {
            return
        }

        const oldIndex = orderedTeams.findIndex((team) => team.id === active.id)
        const newIndex = orderedTeams.findIndex((team) => team.id === over.id)
        if (oldIndex < 0 || newIndex < 0 || oldIndex === newIndex) {
            return
        }

        const previousTeams = orderedTeams
        const nextTeams = renumberTeams(arrayMove(orderedTeams, oldIndex, newIndex))
        setOrderedTeams(nextTeams)
        setSavingOrder(true)
        setError(null)

        try {
            await reorderDutySettingsTeams(groupId, nextTeams.map((team) => team.id))
            await onReload()
        } catch (currentError) {
            setOrderedTeams(previousTeams)
            setError(currentError instanceof Error ? currentError.message : 'Не удалось сохранить порядок команд')
        } finally {
            setSavingOrder(false)
        }
    }

    return (
        <Stack gap="md">
            {error ? (
                <Alert color="red">{error}</Alert>
            ) : null}

            {orderedTeams.length > 0 ? (
                <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={(event) => void handleDragEnd(event)}>
                    <SortableContext items={orderedIds} strategy={verticalListSortingStrategy}>
                        <Stack gap={6}>
                            {orderedTeams.map((team) => (
                                <DutySettingsTeamCard
                                    key={team.id}
                                    team={team}
                                    isDutyTeam={activeDutyTeamId === team.id}
                                    disabled={savingOrder}
                                    onOpenMembers={setSelectedTeam}
                                    onAssignDuty={setAssigningDutyTeam}
                                    onDelete={setDeletingTeam}
                                />
                            ))}
                        </Stack>
                    </SortableContext>
                </DndContext>
            ) : null}

            <SettingsAddAction icon={<RowsPlusBottomIcon size={25} />} onClick={handleCreate}>
                Добавить команду
            </SettingsAddAction>

            <TeamFormModal
                opened={formOpened}
                mode="create"
                teamId={null}
                groupId={groupId}
                loadTeamRequest={(teamId) => getDutySettingsTeam(groupId, teamId)}
                loadTeamMemberOptionsRequest={(currentGroupId, teamId) => getDutySettingsTeamMemberOptions(currentGroupId, teamId)}
                createTeamRequest={(request) => createDutySettingsTeam(groupId, request)}
                updateTeamRequest={(teamId, request) => updateDutySettingsTeam(groupId, teamId, request)}
                onClose={() => {
                    setFormOpened(false)
                }}
                onSaved={onReload}
            />

            <DeleteTeamModal
                opened={deletingTeam !== null}
                team={deletingTeam ? toTeamListItem(deletingTeam) : null}
                deleteTeamRequest={(teamId) => deleteDutySettingsTeam(groupId, teamId)}
                onClose={() => setDeletingTeam(null)}
                onDeleted={async () => {
                    setDeletingTeam(null)
                    await onReload()
                }}
            />

            <ConfirmActionModal
                opened={assigningDutyTeam !== null}
                title="Назначить дежурной"
                description={`Вы уверены, что хотите назначить команду ${getDutyTeamWarningName(assigningDutyTeam)} дежурной? Прогресс по задачам текущей дежурной команды будет утерян.`}
                confirmLabel="Назначить"
                onClose={() => setAssigningDutyTeam(null)}
                onConfirm={async () => {
                    if (!assigningDutyTeam) {
                        return
                    }

                    await assignDutySettingsActiveTeam(groupId, assigningDutyTeam.id)
                    setAssigningDutyTeam(null)
                    await onReload()
                }}
                errorMessage="Не удалось назначить дежурную команду"
            />

            <DutySettingsTeamMembersDrawer
                opened={selectedTeam !== null}
                team={selectedTeam}
                groupId={groupId}
                onClose={() => setSelectedTeam(null)}
                onUpdated={onReload}
            />
        </Stack>
    )
}
