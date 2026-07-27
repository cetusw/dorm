import { useEffect, useMemo, useState } from 'react'

import { Select, Stack, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'

import { ApiError } from '../../../shared/api/ApiError'
import { EntityFormModal } from '../../../shared/ui/EntityFormModal'
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
import { teamFormValidation } from '../model/validation'
import { TeamMembersSelection } from './TeamMembersSelection'

type Props = {
    opened: boolean
    mode: 'create' | 'edit'
    teamId: string | null
    groupId: string
    onClose: () => void
    onSaved: () => Promise<void> | void
    loadTeamRequest?: (teamId: string) => Promise<TeamDetails>
    loadTeamMemberOptionsRequest?: (groupId: string, teamId?: string | null) => Promise<TeamMemberOptionsResponse>
    createTeamRequest?: (request: CreateTeamRequest) => Promise<TeamDetails>
    updateTeamRequest?: (teamId: string, request: UpdateTeamRequest) => Promise<TeamDetails>
}

const initialValues: TeamFormValues = {
    name: '',
    leaderId: null,
    memberIds: [],
}

function toCreateRequest(values: TeamFormValues, groupId: string): CreateTeamRequest {
    return {
        name: values.name.trim(),
        group_id: groupId,
        leader_id: values.leaderId,
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
    loadTeamRequest = getTeam,
    loadTeamMemberOptionsRequest = getTeamMemberOptions,
    createTeamRequest = createTeam,
    updateTeamRequest = updateTeam,
}: Props) {
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)
    const [members, setMembers] = useState<TeamMemberOption[]>([])

    const form = useForm<TeamFormValues>({
        mode: 'controlled',
        initialValues,
        validate: {
            name: teamFormValidation.name,
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
                        name: team.name,
                        leaderId: team.leader?.id ?? null,
                        memberIds: ensureLeaderIncluded(team.member_ids, team.leader?.id ?? null),
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
            })),
        [members],
    )

    const title = mode === 'create' ? 'Создание команды' : 'Редактирование команды'

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
            onSubmit={form.onSubmit(async (values) => {
                setSubmitError(null)
                setSaving(true)

                const normalizedValues = {
                    ...values,
                    memberIds: ensureLeaderIncluded(values.memberIds, values.leaderId),
                }

                try {
                    if (mode === 'create') {
                        await createTeamRequest(toCreateRequest(normalizedValues, groupId))
                    } else if (teamId !== null) {
                        await updateTeamRequest(teamId, toUpdateRequest(normalizedValues, groupId))
                    }

                    await onSaved()
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
            <TextInput
                label="Название"
                placeholder="Название"
                withAsterisk
                maxLength={255}
                key={form.key('name')}
                {...form.getInputProps('name')}
            />

            <Select
                label="Глава"
                placeholder="Выберите главу"
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

            <Stack gap="xs">
                <TeamMembersSelection
                    members={members}
                    selectedMemberIds={form.values.memberIds}
                    onToggleMember={toggleMember}
                />
            </Stack>
        </EntityFormModal>
    )
}
