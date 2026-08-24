import { apiRequest } from '../../../shared/api/apiClient'
import type {
    CreateTeamRequest,
    TeamDetails,
    TeamListResponse,
    TeamMemberOptionsResponse,
    UpdateTeamRequest,
} from '../model/types'

function normalizeTeamListResponse(response: TeamListResponse): TeamListResponse {
    return {
        teams: Array.isArray(response.teams) ? response.teams : [],
    }
}

function normalizeTeamDetails(team: Record<string, unknown>): TeamDetails {
    return {
        id: String(team.id ?? ''),
        group_id: String(team.group_id ?? ''),
        group_name: String(team.group_name ?? ''),
        leader: team.leader && typeof team.leader === 'object'
            ? {
                id: String((team.leader as Record<string, unknown>).id ?? ''),
                name: String((team.leader as Record<string, unknown>).name ?? ''),
            }
            : { id: '', name: '' },
        member_ids: Array.isArray(team.member_ids)
            ? team.member_ids.map((memberId) => String(memberId))
            : [],
    }
}

function normalizeTeamMemberOptionsResponse(
    response: TeamMemberOptionsResponse,
): TeamMemberOptionsResponse {
    return {
        members: Array.isArray(response.members) ? response.members : [],
    }
}

export async function getTeams(groupId: string): Promise<TeamListResponse> {
    const response = await apiRequest(`/api/v1/teams?group_id=${encodeURIComponent(groupId)}`)
    return normalizeTeamListResponse(await response.json())
}

export async function getTeam(teamId: string): Promise<TeamDetails> {
    const response = await apiRequest(`/api/v1/teams/${teamId}`)
    return normalizeTeamDetails(await response.json())
}

export async function getTeamMemberOptions(
    groupId: string,
    teamId?: string | null,
): Promise<TeamMemberOptionsResponse> {
    const params = new URLSearchParams({ group_id: groupId })
    if (teamId) {
        params.set('team_id', teamId)
    }

    const response = await apiRequest(`/api/v1/teams/members?${params.toString()}`)
    return normalizeTeamMemberOptionsResponse(await response.json())
}

export async function createTeam(request: CreateTeamRequest): Promise<TeamDetails> {
    const response = await apiRequest('/api/v1/teams', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeTeamDetails(await response.json())
}

export async function updateTeam(teamId: string, request: UpdateTeamRequest): Promise<TeamDetails> {
    const response = await apiRequest(`/api/v1/teams/${teamId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeTeamDetails(await response.json())
}

export async function deleteTeam(teamId: string): Promise<void> {
    await apiRequest(`/api/v1/teams/${teamId}`, {
        method: 'DELETE',
    })
}
