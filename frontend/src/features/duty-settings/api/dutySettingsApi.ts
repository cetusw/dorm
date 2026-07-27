import { apiRequest } from '../../../shared/api/apiClient'
import type { AreaDetails, CreateAreaRequest, UpdateAreaRequest } from '../../areas/model/types'
import type { CreateTaskRequest, TaskDetails, UpdateTaskRequest } from '../../task-catalog/model/types'
import type {
    CreateTeamRequest,
    TeamDetails,
    TeamMemberOptionsResponse,
    UpdateTeamRequest,
} from '../../teams/model/types'
import type {
    DutySettingsResponse,
    DutySettingsTeamMembersResponse,
    DutySettingsTeamSearchResponse,
} from '../model/types'

function normalizeDutySettingsResponse(response: DutySettingsResponse): DutySettingsResponse {
    return {
        group: response.group,
        areas: Array.isArray(response.areas)
                    ? response.areas.map((area) => ({
                        ...area,
                        floor: area.floor == null ? null : Number(area.floor),
                        tasks: Array.isArray(area.tasks)
                            ? area.tasks.map((task) => ({
                                ...task,
                                last_completed_at: task.last_completed_at ?? null,
                                is_included: Boolean(task.is_included),
                                assignee_name: task.assignee_name ?? null,
                                status: typeof task.status === 'string' ? task.status : '',
                            }))
                            : [],
                    }))
            : [],
        teams: Array.isArray(response.teams)
            ? response.teams.map((team) => ({
                id: String(team.id ?? ''),
                name: String(team.name ?? ''),
                rotation_position: Number(team.rotation_position ?? 0),
                leader: team.leader && typeof team.leader === 'object'
                    ? {
                        id: String((team.leader as Record<string, unknown>).id ?? ''),
                        name: String((team.leader as Record<string, unknown>).name ?? ''),
                    }
                    : null,
                members_count: Number(team.members_count ?? 0),
            }))
            : [],
        active_duty_team_id: response.active_duty_team_id ?? null,
        task_editor_state: response.task_editor_state === 'active'
            || response.task_editor_state === 'no_duties'
            || response.task_editor_state === 'no_active_duty'
            ? response.task_editor_state
            : 'no_active_duty',
        task_editor_alert: String(response.task_editor_alert ?? ''),
        active_duty: response.active_duty && typeof response.active_duty === 'object'
            ? {
                id: String((response.active_duty as Record<string, unknown>).id ?? ''),
                team_id: String((response.active_duty as Record<string, unknown>).team_id ?? ''),
                team_name: String((response.active_duty as Record<string, unknown>).team_name ?? ''),
                start_date: String((response.active_duty as Record<string, unknown>).start_date ?? ''),
                end_date: String((response.active_duty as Record<string, unknown>).end_date ?? ''),
                summary: {
                    task_count: Number((((response.active_duty as Record<string, unknown>).summary as Record<string, unknown>)?.task_count) ?? 0),
                    total_cost: Number((((response.active_duty as Record<string, unknown>).summary as Record<string, unknown>)?.total_cost) ?? 0),
                    cost_per_member: Number((((response.active_duty as Record<string, unknown>).summary as Record<string, unknown>)?.cost_per_member) ?? 0),
                    team_member_count: Number((((response.active_duty as Record<string, unknown>).summary as Record<string, unknown>)?.team_member_count) ?? 0),
                },
            }
            : null,
    }
}

function normalizeAreaDetails(area: Record<string, unknown>): AreaDetails {
    return {
        id: Number(area.id),
        name: String(area.name ?? ''),
        floor: area.floor == null ? null : Number(area.floor),
        group: area.group && typeof area.group === 'object'
            ? {
                id: String((area.group as Record<string, unknown>).id ?? ''),
                name: String((area.group as Record<string, unknown>).name ?? ''),
            }
            : null,
    }
}

function normalizeTaskDetails(task: Record<string, unknown>): TaskDetails {
    const area = task.area && typeof task.area === 'object'
        ? (task.area as Record<string, unknown>)
        : {}

    return {
        id: String(task.id ?? ''),
        title: String(task.title ?? ''),
        cost: Number(task.cost ?? 0),
        frequency: Number(task.frequency ?? 0),
        area: {
            id: Number(area.id ?? 0),
            name: String(area.name ?? ''),
        },
    }
}

function normalizeTeamDetails(team: Record<string, unknown>): TeamDetails {
    return {
        id: String(team.id ?? ''),
        name: String(team.name ?? ''),
        group_id: String(team.group_id ?? ''),
        group_name: String(team.group_name ?? ''),
        leader: team.leader && typeof team.leader === 'object'
            ? {
                id: String((team.leader as Record<string, unknown>).id ?? ''),
                name: String((team.leader as Record<string, unknown>).name ?? ''),
            }
            : null,
        member_ids: Array.isArray(team.member_ids)
            ? team.member_ids.map((memberId) => String(memberId))
            : [],
    }
}

function normalizeTeamMemberOptionsResponse(response: TeamMemberOptionsResponse): TeamMemberOptionsResponse {
    return {
        members: Array.isArray(response.members) ? response.members : [],
    }
}

function normalizeDutySettingsTeamMembersResponse(response: Record<string, unknown>): DutySettingsTeamMembersResponse {
    return {
        team_name: String(response.team_name ?? ''),
        leader: response.leader && typeof response.leader === 'object'
            ? {
                id: String((response.leader as Record<string, unknown>).id ?? ''),
                name: String((response.leader as Record<string, unknown>).name ?? ''),
            }
            : null,
        members: Array.isArray(response.members)
            ? response.members.map((member) => {
                const normalized = member as Record<string, unknown>
                return {
                    id: String(normalized.id ?? ''),
                    name: String(normalized.name ?? ''),
                    is_leader: Boolean(normalized.is_leader),
                }
            })
            : [],
    }
}

function normalizeDutySettingsTeamSearchResponse(response: Record<string, unknown>): DutySettingsTeamSearchResponse {
    return {
        users: Array.isArray(response.users)
            ? response.users.map((user) => {
                const normalized = user as Record<string, unknown>
                const currentTeamLeader = normalized.current_team_leader && typeof normalized.current_team_leader === 'object'
                    ? {
                        id: String((normalized.current_team_leader as Record<string, unknown>).id ?? ''),
                        name: String((normalized.current_team_leader as Record<string, unknown>).name ?? ''),
                    }
                    : null

                return {
                    id: String(normalized.id ?? ''),
                    name: String(normalized.name ?? ''),
                    current_team_id: normalized.current_team_id == null ? null : String(normalized.current_team_id),
                    current_team_leader: currentTeamLeader,
                }
            })
            : [],
    }
}

export async function getDutySettings(groupId: string): Promise<DutySettingsResponse> {
    const response = await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings`)
    return normalizeDutySettingsResponse(await response.json())
}

export async function createDutySettingsArea(groupId: string, request: CreateAreaRequest): Promise<AreaDetails> {
    const response = await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/areas`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeAreaDetails(await response.json())
}

export async function updateDutySettingsArea(
    groupId: string,
    areaId: number,
    request: UpdateAreaRequest,
): Promise<AreaDetails> {
    const response = await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/areas/${areaId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeAreaDetails(await response.json())
}

export async function deleteDutySettingsArea(groupId: string, areaId: number): Promise<void> {
    await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/areas/${areaId}`, {
        method: 'DELETE',
    })
}

export async function createDutySettingsTask(
    groupId: string,
    areaId: number,
    request: CreateTaskRequest,
): Promise<TaskDetails> {
    const response = await apiRequest(
        `/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/areas/${areaId}/tasks`,
        {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(request),
        },
    )

    return normalizeTaskDetails(await response.json())
}

export async function updateDutySettingsTask(
    groupId: string,
    taskId: string,
    request: UpdateTaskRequest,
): Promise<TaskDetails> {
    const response = await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/tasks/${taskId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeTaskDetails(await response.json())
}

export async function deleteDutySettingsTask(groupId: string, taskId: string): Promise<void> {
    await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/tasks/${taskId}`, {
        method: 'DELETE',
    })
}

export async function includeDutySettingsTask(groupId: string, taskId: string): Promise<void> {
    await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/tasks/${taskId}/include`, {
        method: 'POST',
    })
}

export async function excludeDutySettingsTask(groupId: string, taskId: string): Promise<void> {
    await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/tasks/${taskId}/include`, {
        method: 'DELETE',
    })
}

export async function getDutySettingsTeam(groupId: string, teamId: string): Promise<TeamDetails> {
    const response = await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/teams/${encodeURIComponent(teamId)}`)
    return normalizeTeamDetails(await response.json())
}

export async function getDutySettingsTeamMembers(groupId: string, teamId: string): Promise<DutySettingsTeamMembersResponse> {
    const response = await apiRequest(
        `/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/teams/${encodeURIComponent(teamId)}/members`,
    )
    return normalizeDutySettingsTeamMembersResponse(await response.json())
}

export async function searchDutySettingsTeamMembers(
    groupId: string,
    teamId: string,
    query: string,
): Promise<DutySettingsTeamSearchResponse> {
    const params = new URLSearchParams({ q: query })
    const response = await apiRequest(
        `/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/teams/${encodeURIComponent(teamId)}/member-search?${params.toString()}`,
    )
    return normalizeDutySettingsTeamSearchResponse(await response.json())
}

export async function getDutySettingsTeamMemberOptions(
    groupId: string,
    teamId?: string | null,
): Promise<TeamMemberOptionsResponse> {
    const params = new URLSearchParams()
    if (teamId) {
        params.set('team_id', teamId)
    }

    const suffix = params.toString().length > 0 ? `?${params.toString()}` : ''
    const response = await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/teams/members${suffix}`)
    return normalizeTeamMemberOptionsResponse(await response.json())
}

export async function createDutySettingsTeam(groupId: string, request: CreateTeamRequest): Promise<TeamDetails> {
    const response = await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/teams`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeTeamDetails(await response.json())
}

export async function addDutySettingsTeamMember(groupId: string, teamId: string, userId: string): Promise<void> {
    await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/teams/${encodeURIComponent(teamId)}/members`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ user_id: userId }),
    })
}

export async function assignDutySettingsTeamLeader(groupId: string, teamId: string, userId: string): Promise<void> {
    await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/teams/${encodeURIComponent(teamId)}/leader`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ user_id: userId }),
    })
}

export async function removeDutySettingsTeamMember(groupId: string, teamId: string, userId: string): Promise<void> {
    await apiRequest(
        `/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/teams/${encodeURIComponent(teamId)}/members/${encodeURIComponent(userId)}`,
        {
            method: 'DELETE',
        },
    )
}

export async function updateDutySettingsTeam(groupId: string, teamId: string, request: UpdateTeamRequest): Promise<TeamDetails> {
    const response = await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/teams/${encodeURIComponent(teamId)}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeTeamDetails(await response.json())
}

export async function deleteDutySettingsTeam(groupId: string, teamId: string): Promise<void> {
    await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/teams/${encodeURIComponent(teamId)}`, {
        method: 'DELETE',
    })
}

export async function reorderDutySettingsTeams(groupId: string, teamIds: string[]): Promise<void> {
    await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/teams/reorder`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ team_ids: teamIds }),
    })
}
