export type DutySettingsTask = {
    id: string
    title: string
    cost: number
    recurrenceInterval: number
    startSequence: number
    last_completed_at: string | null
    is_included: boolean
    assignee_name: string | null
    status: 'free' | 'assigned' | 'completed' | 'verified' | ''
}

export type DutySettingsTaskSummary = {
    task_count: number
    total_cost: number
    cost_per_participant: number
    participant_count: number
}

export type DutySettingsActiveDuty = {
    id: string
    team_id: string
    team_name: string
    start_date: string
    end_date: string
    summary: DutySettingsTaskSummary
}

export type DutySettingsTeamLeader = {
    id: string
    name: string
}

export type DutySettingsTeamMember = {
    id: string
    name: string
    is_leader: boolean
}

export type DutySettingsTeamMembersResponse = {
    team_name: string
    leader: DutySettingsTeamLeader | null
    members: DutySettingsTeamMember[]
}

export type DutySettingsTeamSearchItem = {
    id: string
    name: string
    current_team_id: string | null
    current_team_leader: DutySettingsTeamLeader | null
}

export type DutySettingsTeamSearchResponse = {
    users: DutySettingsTeamSearchItem[]
}

export type DutySettingsTeam = {
    id: string
    rotation_position: number
    leader: DutySettingsTeamLeader
    members_count: number
}

export type DutySettingsArea = {
    id: number
    name: string
    floor: number | null
    tasks: DutySettingsTask[]
}

export type DutySettingsGroup = {
    id: string
    name: string
    dormitory_id: number
}

export type DutySettingsResponse = {
    can_manage_group_settings: boolean
    group: DutySettingsGroup
    areas: DutySettingsArea[]
    teams: DutySettingsTeam[]
    active_duty_team_id: string | null
    task_editor_state: 'active' | 'no_duties' | 'no_active_duty'
    task_editor_alert: string
    active_duty: DutySettingsActiveDuty | null
}

export type DutyParticipant = {
    participant_id: string
    full_name: string
    type: 'REGULAR' | 'TEMPORARY'
    excluded_at: string | null
    is_leader: boolean
    team_id: string | null
    team_name: string | null
}

export type DutyParticipantCandidate = {
    participant_id: string
    full_name: string
    team_id: string | null
    team_name: string | null
}

export type DutySettingsMainTab = 'tasks' | 'participants' | 'teams'

export type DutySettingsViewMode = 'list' | 'plan'
