export type TeamLeader = {
    id: string
    name: string
}

export type TeamListItem = {
    id: string
    name: string
    leader: TeamLeader | null
    members_count: number
    rotation_position?: number
}

export type TeamListResponse = {
    teams: TeamListItem[]
}

export type TeamDetails = {
    id: string
    name: string
    group_id: string
    group_name: string
    leader: TeamLeader | null
    member_ids: string[]
}

export type TeamMemberOption = {
    id: string
    name: string
    current_team_id: string | null
    current_team_name: string
    is_in_current_team: boolean
}

export type TeamMemberOptionsResponse = {
    members: TeamMemberOption[]
}

export type TeamFormValues = {
    name: string
    leaderId: string | null
    memberIds: string[]
}

export type CreateTeamRequest = {
    name: string
    group_id: string
    leader_id: string | null
    member_ids: string[]
}

export type UpdateTeamRequest = CreateTeamRequest
