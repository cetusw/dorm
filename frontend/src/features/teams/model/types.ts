export type TeamLeader = {
    id: string
    name: string
}

export type TeamListItem = {
    id: string
    leader: TeamLeader
    members_count: number
    rotation_position?: number
}

export type TeamListResponse = {
    teams: TeamListItem[]
}

export type TeamDetails = {
    id: string
    group_id: string
    group_name: string
    leader: TeamLeader
    member_ids: string[]
}

export type TeamMemberOption = {
    id: string
    name: string
    current_team_id: string | null
    current_team_name: string
	is_in_current_team: boolean
	is_team_leader: boolean
}

export type TeamMemberOptionsResponse = {
    members: TeamMemberOption[]
}

export type TeamFormValues = {
	leaderId: string | null
    memberIds: string[]
}

export type CreateTeamRequest = {
	group_id: string
	leader_id: string
    member_ids: string[]
}

export type UpdateTeamRequest = CreateTeamRequest
