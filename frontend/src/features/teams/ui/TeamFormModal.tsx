import { useEffect, useMemo, useState } from 'react'

import { Select, Stack } from '@mantine/core'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
import { ResidentSearchCombobox, type ResidentSearchOption } from '../../../shared/ui/ResidentSearchCombobox'
import {
    createTeam,
    getTeam,
    getTeamMemberOptions,
    updateTeam,
} from '../api/teamsApi'
import type {
    CreateTeamRequest,
    TeamDetails,
    TeamFormValues,
    TeamMemberOption,
    TeamMemberOptionsResponse,
    UpdateTeamRequest,
} from '../model/types'
import { TeamMembersSelection } from './TeamMembersSelection'

type Props = {
    opened: boolean
    mode: 'create' | 'edit'
    teamId: string | null
    groupId: string
    onClose: () => void
    onSaved: (team: TeamDetails) => Promise<void> | void
    createTitle?: string
    simpleCreate?: boolean
    leaderLabel?: string
    loadTeamRequest?: (teamId: string) => Promise<TeamDetails>
    loadTeamMemberOptionsRequest?: (groupId: string, teamId?: string | null) => Promise<TeamMemberOptionsResponse>
    createTeamRequest?: (request: CreateTeamRequest) => Promise<TeamDetails>
    updateTeamRequest?: (teamId: string, request: UpdateTeamRequest) => Promise<TeamDetails>
}

const initialValues: TeamFormValues = {
    leaderId: null,
    memberIds: [],
}

function toCreateRequest(values: TeamFormValues, groupId: string): CreateTeamRequest {
    return {
        group_id: groupId,
        leader_id: values.leaderId ?? '',
        member_ids: values.memberIds,
    }
}

function toUpdateRequest(values: TeamFormValues, groupId: string): UpdateTeamRequest {
    return toCreateRequest(values, groupId)
}

function ensureLeaderIncluded(memberIds: string[], leaderId: string | null): string[] {
    if (!leaderId) {
        return memberIds
    }

    return memberIds.includes(leaderId) ? memberIds : [...memberIds, leaderId]
}

export function TeamFormModal({
    opened,
    mode,
    teamId,
    groupId,
    onClose,
    onSaved,
    createTitle = 'Создание команды',
    simpleCreate = false,
    leaderLabel = 'Глава',
    loadTeamRequest = getTeam,
    loadTeamMemberOptionsRequest = getTeamMemberOptions,
    createTeamRequest = createTeam,
    updateTeamRequest = updateTeam,
}: Props) {
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)
    const [members, setMembers] = useState<TeamMemberOption[]>([])
    const [leaderSearch, setLeaderSearch] = useState('')

    const form = useForm<TeamFormValues>({
        mode: 'controlled',
        initialValues,
        validate: {
            leaderId: (value) => value ? null : 'Выберите главу команды',
        },
    })

    useEffect(() => {
        if (!opened) {
            form.setValues(initialValues)
            form.resetDirty(initialValues)
            form.clearErrors()
            setSubmitError(null)
            setLoading(false)
            setSaving(false)
            setMembers([])
            setLeaderSearch('')
            return
        }

        let active = true

        async function loadModalData() {
            setLoading(true)
            setSubmitError(null)

            try {
                const [{ members: loadedMembers }, team] = await Promise.all([
                    loadTeamMemberOptionsRequest(groupId, teamId),
                    mode === 'edit' && teamId !== null
                        ? loadTeamRequest(teamId)
                        : Promise.resolve(null),
                ])

                if (!active) {
                    return
                }

                setMembers(loadedMembers)

                const values: TeamFormValues = team
                    ? {
                        leaderId: team.leader.id,
                        memberIds: ensureLeaderIncluded(team.member_ids, team.leader.id),
                    }
                    : initialValues

                form.setValues(values)
                form.resetDirty(values)
                form.clearErrors()
            } catch (error) {
                if (!active) {
                    return
                }

                if (error instanceof ApiError) {
                    setSubmitError(error.message)
                } else {
                    setSubmitError('Не удалось загрузить данные команды')
                }
            } finally {
                if (active) {
                    setLoading(false)
                }
            }
        }

        void loadModalData()

        return () => {
            active = false
        }
    }, [groupId, mode, opened, teamId])

    const leaderOptions = useMemo(
        () =>
            members.map((member) => ({
                value: member.id,
                label: member.name,
                disabled: mode === 'create' && member.is_team_leader,
            })),
        [members],
    )
    const leaderSearchOptions = useMemo<ResidentSearchOption[]>(() => {
        const query = leaderSearch.trim().toLowerCase()
        return members
            .filter((member) => !member.is_team_leader)
            .filter((member) => query === '' || member.name.toLowerCase().includes(query))
            .map((member) => ({
                id: member.id,
                name: member.name,
                description: member.current_team_name ? `В команде ${member.current_team_name}` : null,
            }))
    }, [leaderSearch, members])

    const title = mode === 'create' ? createTitle : 'Редактирование команды'

    function toggleMember(memberId: string) {
        if (form.values.memberIds.includes(memberId)) {
            if (form.values.leaderId === memberId) {
                return
            }

            form.setFieldValue(
                'memberIds',
                form.values.memberIds.filter((id) => id !== memberId),
            )
            return
        }

        form.setFieldValue('memberIds', [...form.values.memberIds, memberId])
    }

    return (
        <EntityFormModal
            opened={opened}
            onClose={onClose}
            title={title}
            loading={loading}
            saving={saving}
            error={submitError}
            size={760}
            submitLabel={simpleCreate && mode === 'create' ? 'Добавить' : undefined}
            onSubmit={form.onSubmit(async (values) => {
                setSubmitError(null)
                setSaving(true)

                const normalizedValues = {
                    ...values,
                    memberIds: ensureLeaderIncluded(values.memberIds, values.leaderId),
                }

                try {
                    if (mode === 'create') {
                        const created = await createTeamRequest(toCreateRequest(normalizedValues, groupId))
                        await onSaved(created)
                    } else if (teamId !== null) {
                        const updated = await updateTeamRequest(teamId, toUpdateRequest(normalizedValues, groupId))
                        await onSaved(updated)
                    }
                    onClose()
                } catch (error) {
                    if (error instanceof ApiError) {
                        setSubmitError(error.message)
                    } else {
                        setSubmitError('Не удалось сохранить команду')
                    }
                } finally {
                    setSaving(false)
                }
            })}
        >
            {simpleCreate ? (
                <ResidentSearchCombobox
                    overlayOpened={opened}
                    searchValue={leaderSearch}
                    options={leaderSearchOptions}
                    loading={loading}
                    placeholder={leaderLabel}
                    selectedId={form.values.leaderId}
                    selectedLabel={leaderSearch}
                    hideDropdownWhenSelected
                    onSearchChange={(value) => {
                        setLeaderSearch(value)
                        form.setFieldValue('leaderId', null)
                    }}
                    onOptionSelect={(option) => {
                        setLeaderSearch(option.name)
                        form.setFieldValue('leaderId', option.id)
                        form.setFieldValue('memberIds', ensureLeaderIncluded(form.values.memberIds, option.id))
                    }}
                />
            ) : (
                <Select
                    label={leaderLabel}
                    placeholder={leaderLabel}
                    searchable
                    clearable
                    data={leaderOptions}
                    nothingFoundMessage="Житель не найден"
                    value={form.values.leaderId}
                    onChange={(value) => {
                        form.setFieldValue('leaderId', value)
                        form.setFieldValue('memberIds', ensureLeaderIncluded(form.values.memberIds, value))
                    }}
                />
            )}

            {!simpleCreate ? (
                <Stack gap="xs">
                    <TeamMembersSelection
                        members={members}
                        selectedMemberIds={form.values.memberIds}
                        onToggleMember={toggleMember}
                    />
                </Stack>
            ) : null}
        </EntityFormModal>
    )
}
