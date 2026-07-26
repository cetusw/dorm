export type DutySettingsTask = {
    id: string
    title: string
    cost: number
    frequency: number
    last_completed_at: string | null
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
}

export type DutySettingsMainTab = 'tasks' | 'teams' | 'next-duty'

export type DutySettingsViewMode = 'list' | 'plan'
