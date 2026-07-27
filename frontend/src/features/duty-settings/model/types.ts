export type DutySettingsTask = {
    id: string
    title: string
    cost: number
    frequency: number
    last_completed_at: string | null
}

export type DutySettingsTeamLeader = {
    id: string
    name: string
}

export type DutySettingsTeam = {
    id: string
    name: string
    rotation_position: number
    leader: DutySettingsTeamLeader | null
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
    group: DutySettingsGroup
    areas: DutySettingsArea[]
    teams: DutySettingsTeam[]
    active_duty_team_id: string | null
}

export type DutySettingsMainTab = 'tasks' | 'teams' | 'next-duty'

export type DutySettingsViewMode = 'list' | 'plan'
